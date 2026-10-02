package contentfactory

import (
	"fmt"
	"sort"
	"strings"
)

type TargetFilter struct {
	Roles         []string `json:"roles"`
	ExcludeNpcIDs []int    `json:"excludeNpcIds"`
	OnlyNpcIDs    []int    `json:"onlyNpcIds"`
	NameContains  string   `json:"nameContains"`
	SkipMerchants *bool    `json:"skipMerchants"`
	IncludeRing   bool     `json:"includeRing"`
}

type ClassifiedNPC struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Reason      string `json:"reason,omitempty"`
	Level       int    `json:"level"`
	HP          int64  `json:"hp"`
	Pops        int    `json:"pops"`
	MerchantID  int    `json:"merchantId"`
	LoottableID int    `json:"loottableId"`
	NpcSpellsID int    `json:"npcSpellsId"`
	FactionID   int    `json:"factionId"`
	RaidTarget  int    `json:"raidTarget,omitempty"`
}

type ZoneNeighbor struct {
	ID    int    `json:"id"`
	Short string `json:"short"`
	Long  string `json:"long,omitempty"`
}

func (f TargetFilter) roles() []string {
	out := []string{}
	for _, r := range f.Roles {
		r = strings.ToLower(strings.TrimSpace(r))
		if r != "" {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return []string{"trash"}
	}
	return uniqueStrings(out)
}

func (f TargetFilter) skipMerchants() bool {
	if f.SkipMerchants == nil {
		return true
	}
	return *f.SkipMerchants
}

func (f TargetFilter) allows(role string) bool {
	for _, r := range f.roles() {
		if r == "all" || r == role {
			return true
		}
		if r == "named" && (role == "boss" || role == "raid") {
			return true
		}
	}
	return false
}

func roleMatch(want, have string) bool {
	want = strings.ToLower(strings.TrimSpace(want))
	have = strings.ToLower(strings.TrimSpace(have))
	if want == "" || want == "all" {
		return true
	}
	if want == "named" {
		return have == "boss" || have == "raid"
	}
	return want == have
}

func (s *Service) expandZones(zoneIDs []int, ring bool) []int {
	ids := uniqueInts(zoneIDs)
	if !ring || s.eqemu() == nil {
		return ids
	}
	for _, id := range ids {
		short, err := s.zoneShort(id)
		if err != nil {
			continue
		}
		var targets []int
		_ = s.eqemu().Table("zone_points").Select("DISTINCT target_zone_id").
			Where("zone = ? AND target_zone_id > 0", short).
			Pluck("target_zone_id", &targets).Error
		ids = append(ids, targets...)
	}
	return uniqueInts(ids)
}

func (s *Service) zoneNeighbors(short string) []ZoneNeighbor {
	out := []ZoneNeighbor{}
	if s.eqemu() == nil || short == "" {
		return out
	}
	type row struct {
		ID    int    `gorm:"column:zoneidnumber"`
		Short string `gorm:"column:short_name"`
		Long  string `gorm:"column:long_name"`
	}
	var rows []row
	_ = s.eqemu().Table("zone_points").
		Select("DISTINCT zone.zoneidnumber, zone.short_name, zone.long_name").
		Joins("JOIN zone ON zone.zoneidnumber = zone_points.target_zone_id").
		Where("zone_points.zone = ? AND zone_points.target_zone_id > 0", short).
		Order("zone.short_name").
		Scan(&rows).Error
	seen := map[int]bool{}
	for _, r := range rows {
		if r.ID <= 0 || seen[r.ID] {
			continue
		}
		seen[r.ID] = true
		out = append(out, ZoneNeighbor{ID: r.ID, Short: r.Short, Long: r.Long})
	}
	return out
}

func (s *Service) classifyZoneNpcs(zoneIDs []int) ([]ClassifiedNPC, error) {
	out := []ClassifiedNPC{}
	if s.eqemu() == nil {
		return out, fmt.Errorf("PEQ database is not connected")
	}
	seen := map[int]bool{}
	for _, zoneID := range uniqueInts(zoneIDs) {
		short, err := s.zoneShort(zoneID)
		if err != nil {
			return out, err
		}
		type row struct {
			ID          int
			Name        string
			Level       int
			HP          int64
			RaidTarget  int    `gorm:"column:raid_target"`
			Pops        int    `gorm:"column:pops"`
			MerchantID  int    `gorm:"column:merchant_id"`
			LoottableID int    `gorm:"column:loottable_id"`
			NpcSpellsID int    `gorm:"column:npc_spells_id"`
			FactionID   int    `gorm:"column:npc_faction_id"`
		}
		var rows []row
		err = s.eqemu().Table("spawn2").
			Select("npc_types.id, npc_types.name, npc_types.level, npc_types.hp, npc_types.raid_target, npc_types.merchant_id, npc_types.loottable_id, npc_types.npc_spells_id, npc_types.npc_faction_id, COUNT(spawn2.id) AS pops").
			Joins("JOIN spawnentry ON spawnentry.spawngroupID = spawn2.spawngroupID").
			Joins("JOIN npc_types ON npc_types.id = spawnentry.npcID").
			Where("spawn2.zone = ?", short).
			Group("npc_types.id, npc_types.name, npc_types.level, npc_types.hp, npc_types.raid_target, npc_types.merchant_id, npc_types.loottable_id, npc_types.npc_spells_id, npc_types.npc_faction_id").
			Scan(&rows).Error
		if err != nil {
			return out, err
		}
		agg := make([]spawnAgg, 0, len(rows))
		byID := map[int]row{}
		for _, r := range rows {
			byID[r.ID] = r
			agg = append(agg, spawnAgg{ID: r.ID, Name: r.Name, Level: r.Level, HP: r.HP, RaidTarget: r.RaidTarget, Pops: r.Pops})
		}
		for _, mob := range classifySpawnRows(agg) {
			if seen[mob.ID] {
				continue
			}
			seen[mob.ID] = true
			src := byID[mob.ID]
			out = append(out, ClassifiedNPC{
				ID: mob.ID, Name: src.Name, Role: mob.Role, Reason: mob.Reason,
				Level: src.Level, HP: src.HP, Pops: src.Pops, MerchantID: src.MerchantID,
				LoottableID: src.LoottableID, NpcSpellsID: src.NpcSpellsID, FactionID: src.FactionID,
				RaidTarget: src.RaidTarget,
			})
		}
	}
	return out, nil
}

func (s *Service) npcIDsFiltered(zoneIDs []int, extra []int, filter TargetFilter) ([]int, []ClassifiedNPC, error) {
	zoneIDs = s.expandZones(zoneIDs, filter.IncludeRing)
	classified, err := s.classifyZoneNpcs(zoneIDs)
	if err != nil {
		return nil, nil, err
	}
	if len(extra) > 0 {
		have := map[int]bool{}
		for _, n := range classified {
			have[n.ID] = true
		}
		missing := []int{}
		for _, id := range uniqueInts(extra) {
			if !have[id] {
				missing = append(missing, id)
			}
		}
		if len(missing) > 0 {
			type row struct {
				ID          int
				Name        string
				Level       int
				HP          int64
				MerchantID  int `gorm:"column:merchant_id"`
				LoottableID int `gorm:"column:loottable_id"`
				NpcSpellsID int `gorm:"column:npc_spells_id"`
				FactionID   int `gorm:"column:npc_faction_id"`
				RaidTarget  int `gorm:"column:raid_target"`
			}
			var rows []row
			_ = s.eqemu().Table("npc_types").
				Select("id, name, level, hp, merchant_id, loottable_id, npc_spells_id, npc_faction_id, raid_target").
				Where("id IN ?", missing).Scan(&rows).Error
			for _, r := range rows {
				role := "trash"
				if isIgnoreName(r.Name) {
					role = "ignore"
				} else if r.RaidTarget > 0 {
					role = "raid"
				} else if looksNamed(r.Name) {
					role = "boss"
				}
				classified = append(classified, ClassifiedNPC{
					ID: r.ID, Name: r.Name, Role: role, Level: r.Level, HP: r.HP,
					MerchantID: r.MerchantID, LoottableID: r.LoottableID, NpcSpellsID: r.NpcSpellsID,
					FactionID: r.FactionID, RaidTarget: r.RaidTarget, Pops: 1,
				})
			}
		}
	}
	only := map[int]bool{}
	for _, id := range filter.OnlyNpcIDs {
		only[id] = true
	}
	exclude := map[int]bool{}
	for _, id := range filter.ExcludeNpcIDs {
		exclude[id] = true
	}
	name := strings.ToLower(strings.TrimSpace(filter.NameContains))
	kept := []ClassifiedNPC{}
	ids := []int{}
	for _, n := range classified {
		if exclude[n.ID] {
			continue
		}
		if len(only) > 0 && !only[n.ID] {
			continue
		}
		if filter.skipMerchants() && n.MerchantID > 0 {
			continue
		}
		if !filter.allows(n.Role) {
			continue
		}
		if name != "" && !strings.Contains(strings.ToLower(n.Name), name) {
			continue
		}
		kept = append(kept, n)
		ids = append(ids, n.ID)
	}
	return uniqueInts(ids), kept, nil
}

type spawnAgg struct {
	ID         int
	Name       string
	Level      int
	HP         int64
	RaidTarget int
	Pops       int
}

type classMob struct {
	ID     int
	Name   string
	Role   string
	Reason string
}

func classifySpawnRows(rows []spawnAgg) []classMob {
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
	out := make([]classMob, 0, len(rows))
	for _, r := range rows {
		mob := classMob{ID: r.ID, Name: r.Name}
		switch {
		case isIgnoreName(r.Name):
			mob.Role = "ignore"
			mob.Reason = "controller, pet, or placeholder"
		case r.RaidTarget > 0:
			mob.Role = "raid"
			mob.Reason = "raid_target"
		case r.Pops <= 1 && looksNamed(r.Name) && r.HP >= median:
			mob.Role = "boss"
			mob.Reason = "unique named"
		case r.Pops <= 1 && r.HP > 0 && median > 0 && r.HP >= median*3:
			mob.Role = "boss"
			mob.Reason = "unique high HP"
		default:
			mob.Role = "trash"
			mob.Reason = "common spawn"
		}
		out = append(out, mob)
	}
	return out
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
	return n[0] >= 'A' && n[0] <= 'Z'
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, s := range in {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func boolPtr(v bool) *bool {
	return &v
}

func impactKey(kind string, id int, field string) string {
	return fmt.Sprintf("%s:%d:%s", kind, id, field)
}
