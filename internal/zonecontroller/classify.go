package zonecontroller

import (
	"fmt"
	"sort"
	"strings"

	"github.com/EQEmu/spire/internal/models"
)

type ClassifyRequest struct {
	ZoneIDs   []int `json:"zoneIds"`
	Apply     bool  `json:"apply"`
	Overwrite bool  `json:"overwrite"`
	DryRun    *bool `json:"dryRun"`
}

type ClassifyMob struct {
	Name   string `json:"name"`
	NpcID  int    `json:"npcId"`
	Level  int    `json:"level"`
	HP     int64  `json:"hp"`
	Pops   int    `json:"pops"`
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type ClassifyZone struct {
	ZoneID     int           `json:"zoneId"`
	ShortName  string        `json:"shortName"`
	Mobs       []ClassifyMob `json:"mobs"`
	Trash      int           `json:"trash"`
	Boss       int           `json:"boss"`
	Raid       int           `json:"raid"`
	Ignore     int           `json:"ignore"`
	Controller bool          `json:"controller"`
}

type ClassifyPlan struct {
	WritePlan
	Zones []ClassifyZone `json:"zones"`
}

type spawnAgg struct {
	ID         int
	Name       string
	Level      int
	HP         int64
	RaidTarget int
	Pops       int
}

func (s *Service) ClassifyZones(req ClassifyRequest) ClassifyPlan {
	plan := ClassifyPlan{WritePlan: WritePlan{DryRun: dryRunValue(req.DryRun)}, Zones: []ClassifyZone{}}
	ids, err := normalizeZoneIDs(req.ZoneIDs)
	if err != nil {
		plan.Error = err.Error()
		return plan
	}
	if s.eqemu() == nil {
		plan.Error = "database is not connected"
		return plan
	}
	stamp := backupStamp()
	for _, id := range ids {
		zone, err := s.classifyOne(id)
		if err != nil {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("zone %d: %v", id, err))
			continue
		}
		plan.Zones = append(plan.Zones, zone)
		if !req.Apply {
			continue
		}
		if err := s.writeClassification(id, zone, req.Overwrite, plan.DryRun, stamp, &plan.WritePlan); err != nil {
			plan.Error = err.Error()
			return plan
		}
	}
	plan.OK = plan.Error == ""
	return plan
}

func (s *Service) classifyOne(id int) (ClassifyZone, error) {
	info, err := s.zoneInfo(id)
	if err != nil {
		return ClassifyZone{}, err
	}
	rows, err := s.zoneSpawnAgg(info.short)
	if err != nil {
		return ClassifyZone{}, err
	}
	out := ClassifyZone{ZoneID: id, ShortName: info.short, Mobs: []ClassifyMob{}}
	proposed := classifySpawnRows(rows)
	for _, mob := range proposed {
		switch mob.Type {
		case "raid":
			out.Raid++
		case "boss":
			out.Boss++
		case "ignore":
			out.Ignore++
		default:
			out.Trash++
		}
		if strings.Contains(strings.ToLower(mob.Name), "zone controller") {
			out.Controller = true
		}
	}
	out.Mobs = proposed
	return out, nil
}

func (s *Service) zoneSpawnAgg(short string) ([]spawnAgg, error) {
	type row struct {
		ID         int    `gorm:"column:id"`
		Name       string `gorm:"column:name"`
		Level      int    `gorm:"column:level"`
		HP         int64  `gorm:"column:hp"`
		RaidTarget int    `gorm:"column:raid_target"`
		Pops       int    `gorm:"column:pops"`
	}
	var rows []row
	err := s.eqemu().Table("spawn2").
		Select("npc_types.id, npc_types.name, npc_types.level, npc_types.hp, npc_types.raid_target, COUNT(spawn2.id) as pops").
		Joins("JOIN spawnentry ON spawnentry.spawngroupID = spawn2.spawngroupID").
		Joins("JOIN npc_types ON npc_types.id = spawnentry.npcID").
		Where("spawn2.zone = ?", short).
		Group("npc_types.id, npc_types.name, npc_types.level, npc_types.hp, npc_types.raid_target").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]spawnAgg, 0, len(rows))
	for _, r := range rows {
		out = append(out, spawnAgg{ID: r.ID, Name: r.Name, Level: r.Level, HP: r.HP, RaidTarget: r.RaidTarget, Pops: r.Pops})
	}
	return out, nil
}

func classifySpawnRows(rows []spawnAgg) []ClassifyMob {
	hps := make([]int64, 0, len(rows))
	for _, r := range rows {
		if !isIgnoreName(r.Name) {
			hps = append(hps, r.HP)
		}
	}
	sort.Slice(hps, func(i, j int) bool { return hps[i] < hps[j] })
	median := int64(0)
	if len(hps) > 0 {
		median = hps[len(hps)/2]
	}
	out := make([]ClassifyMob, 0, len(rows))
	for _, r := range rows {
		mob := ClassifyMob{Name: cleanMobName(r.Name), NpcID: r.ID, Level: r.Level, HP: r.HP, Pops: r.Pops}
		switch {
		case isIgnoreName(r.Name):
			mob.Type = "ignore"
			mob.Reason = "controller, pet, or placeholder"
		case r.RaidTarget > 0:
			mob.Type = "raid"
			mob.Reason = "raid_target"
		case r.Pops <= 1 && looksNamed(r.Name) && r.HP >= median:
			mob.Type = "boss"
			mob.Reason = "unique named"
		case r.Pops <= 1 && r.HP > 0 && median > 0 && r.HP >= median*3:
			mob.Type = "boss"
			mob.Reason = "unique high HP"
		default:
			mob.Type = "trash"
			mob.Reason = "common spawn"
		}
		out = append(out, mob)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Type != out[j].Type {
			return typeOrder(out[i].Type) < typeOrder(out[j].Type)
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func (s *Service) writeClassification(id int, zone ClassifyZone, overwrite, dry bool, stamp string, plan *WritePlan) error {
	raw, err := s.readKindRaw(id, "mob")
	if err != nil {
		plan.Warnings = append(plan.Warnings, fmt.Sprintf("zone %d has no mob json; use Create first", id))
		return nil
	}
	mob := remapTopKey(raw, id)
	body := pickZoneMap(mob, id)
	custom := asObject(body["custom"])
	ignore := asObject(body["ignore"])
	ignoreAll, _ := ignore["all"].([]interface{})
	added := 0
	for _, row := range zone.Mobs {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		if row.Type == "ignore" {
			if !ignoreHas(ignoreAll, name) {
				ignoreAll = append(ignoreAll, map[string]interface{}{"name": name})
			}
			continue
		}
		if existing := asObject(custom[name]); existing["type"] != nil && asString(existing["type"]) != "" && !overwrite {
			continue
		}
		entry := defaultCustomMob(row.Type)
		if row.NpcID > 0 {
			entry["mobid"] = fmt.Sprintf("%d", row.NpcID)
		}
		if row.Type == "boss" || row.Type == "raid" {
			entry["static"] = "1"
		}
		custom[name] = entry
		added++
	}
	body["custom"] = custom
	ignore["all"] = ignoreAll
	body["ignore"] = ignore
	ensureIgnoreController(body)
	mob[fmt.Sprintf("%d", id)] = body
	_ = added
	return s.commitKind(id, "mob", mob, true, dry, stamp, plan)
}

func ignoreHas(all []interface{}, name string) bool {
	want := strings.ToLower(strings.TrimSpace(name))
	for _, item := range all {
		if strings.ToLower(strings.TrimSpace(asString(asObject(item)["name"]))) == want {
			return true
		}
	}
	return false
}

func isIgnoreName(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	n = strings.TrimPrefix(n, "#")
	if n == "" || n == "zone_controller" || n == "zone controller" || strings.Contains(n, "zone controller") {
		return true
	}
	if strings.Contains(n, "_pet") || strings.HasSuffix(n, "pet") && strings.Contains(n, "a ") {
		return true
	}
	if strings.Contains(n, "placeholder") || strings.Contains(n, "corpse") {
		return true
	}
	return false
}

func looksNamed(name string) bool {
	n := strings.TrimSpace(strings.ReplaceAll(name, "_", " "))
	n = strings.TrimPrefix(n, "#")
	low := strings.ToLower(n)
	if strings.HasPrefix(low, "a ") || strings.HasPrefix(low, "an ") || strings.HasPrefix(low, "the ") {
		return false
	}
	if n == "" {
		return false
	}
	r := n[0]
	return r >= 'A' && r <= 'Z'
}

func cleanMobName(name string) string {
	n := strings.TrimSpace(strings.ReplaceAll(name, "_", " "))
	n = strings.TrimPrefix(n, "#")
	return strings.TrimSpace(n)
}

func typeOrder(t string) int {
	switch t {
	case "raid":
		return 0
	case "boss":
		return 1
	case "trash":
		return 2
	default:
		return 3
	}
}

type zoneInfo struct {
	id    int
	short string
	long  string
}

func (s *Service) zoneInfo(id int) (zoneInfo, error) {
	info := zoneInfo{id: id}
	if db := s.eqemu(); db != nil {
		var z models.Zone
		if err := db.Where("zoneidnumber = ?", id).Order("version asc").First(&z).Error; err == nil {
			info.short = strings.TrimSpace(z.ShortName.String)
			info.long = strings.TrimSpace(z.LongName)
		}
	}
	if info.short == "" {
		return info, fmt.Errorf("no PEQ short_name for zone %d", id)
	}
	return info, nil
}

func (s *Service) zoneHasController(short string) bool {
	if s.eqemu() == nil || short == "" {
		return false
	}
	var n int64
	s.eqemu().Table("spawn2").
		Joins("JOIN spawnentry ON spawnentry.spawngroupID = spawn2.spawngroupID").
		Joins("JOIN npc_types ON npc_types.id = spawnentry.npcID").
		Where("spawn2.zone = ? AND LOWER(npc_types.name) LIKE ?", short, "%zone controller%").
		Count(&n)
	return n > 0
}
