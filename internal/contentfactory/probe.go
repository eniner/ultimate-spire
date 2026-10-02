package contentfactory

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ProbeCheck struct {
	Name string `json:"name"`
	OK   bool   `json:"ok"`
	Note string `json:"note,omitempty"`
}

type ProbeResult struct {
	OK      bool         `json:"ok"`
	Note    string       `json:"note,omitempty"`
	Checks  []ProbeCheck `json:"checks"`
	ZoneIDs []int        `json:"zoneIds,omitempty"`
}

type ProbeRequest struct {
	ZoneIDs  []int `json:"zoneIds"`
	SpellIDs []int `json:"spellIds"`
	ItemIDs  []int `json:"itemIds"`
	NpcIDs   []int `json:"npcIds"`
}

func (s *Service) Probe(req ProbeRequest) ProbeResult {
	out := ProbeResult{Checks: []ProbeCheck{}, ZoneIDs: uniqueInts(req.ZoneIDs)}
	add := func(name string, ok bool, note string) {
		out.Checks = append(out.Checks, ProbeCheck{Name: name, OK: ok, Note: note})
	}
	if s.eqemu() == nil {
		add("database", false, "PEQ is not connected")
		out.Note = "Cannot probe without PEQ"
		return out
	}
	add("database", true, "PEQ connected")

	npcIDs := uniqueInts(req.NpcIDs)
	if len(npcIDs) > 0 {
		type row struct {
			ID          int
			Name        string
			LoottableID int `gorm:"column:loottable_id"`
			NpcSpellsID int `gorm:"column:npc_spells_id"`
		}
		var rows []row
		_ = s.eqemu().Table("npc_types").Select("id, name, loottable_id, npc_spells_id").Where("id IN ?", npcIDs).Scan(&rows).Error
		have := map[int]row{}
		for _, r := range rows {
			have[r.ID] = r
		}
		missing := 0
		for _, id := range npcIDs {
			if _, ok := have[id]; !ok {
				missing++
			}
		}
		add("npc_types", missing == 0, fmt.Sprintf("%d/%d npc rows present", len(have), len(npcIDs)))
		if len(out.ZoneIDs) > 0 {
			var pointed int64
			for _, zoneID := range out.ZoneIDs {
				short, err := s.zoneShort(zoneID)
				if err != nil {
					continue
				}
				var n int64
				_ = s.eqemu().Table("spawn2").
					Joins("JOIN spawnentry ON spawnentry.spawngroupID = spawn2.spawngroupID").
					Where("spawn2.zone = ? AND spawnentry.npcID IN ?", short, npcIDs).
					Count(&n).Error
				pointed += n
			}
			add("spawnentry", pointed > 0, fmt.Sprintf("%d spawn rows point at probed NPCs", pointed))
		}
	}

	itemIDs := uniqueInts(req.ItemIDs)
	if len(itemIDs) > 0 {
		var have []int
		_ = s.eqemu().Table("items").Where("id IN ?", itemIDs).Pluck("id", &have).Error
		add("items", len(have) == len(itemIDs), fmt.Sprintf("%d/%d items present", len(have), len(itemIDs)))
	}

	spellIDs := uniqueInts(req.SpellIDs)
	if len(spellIDs) > 0 {
		var have []int
		_ = s.eqemu().Table("spells_new").Where("id IN ?", spellIDs).Pluck("id", &have).Error
		add("spells_new", len(have) == len(spellIDs), fmt.Sprintf("%d/%d spells present in PEQ", len(have), len(spellIDs)))
		found, path := s.spellsFileContains(spellIDs)
		if path == "" {
			add("spells_us.txt", false, "no exported spells_us.txt yet — write with export checked")
		} else {
			add("spells_us.txt", found == len(spellIDs), fmt.Sprintf("%d/%d new spell ids in %s", found, len(spellIDs), path))
		}
	}

	for _, zoneID := range out.ZoneIDs {
		n, note := s.probeZoneJSON(zoneID, itemIDs)
		add(fmt.Sprintf("zone %d json", zoneID), n >= 0, note)
	}

	fail := 0
	for _, c := range out.Checks {
		if !c.OK {
			fail++
		}
	}
	out.OK = fail == 0
	if out.OK {
		out.Note = "PEQ and export match the probed IDs. Repop still required if the zone is already up."
	} else {
		out.Note = fmt.Sprintf("%d probe check(s) failed", fail)
	}
	return out
}

func (s *Service) ProbePlan(plan *Plan, zoneIDs []int) ProbeResult {
	if plan == nil {
		return ProbeResult{Note: "no plan"}
	}
	req := ProbeRequest{ZoneIDs: zoneIDs}
	for _, c := range plan.Clones {
		switch c.Kind {
		case "spell":
			req.SpellIDs = append(req.SpellIDs, c.NewID)
		case "item":
			req.ItemIDs = append(req.ItemIDs, c.NewID)
		case "npc", "npc_types":
			req.NpcIDs = append(req.NpcIDs, c.NewID)
		}
	}
	for _, row := range plan.Impact {
		if row.Kind == "npc" {
			req.NpcIDs = append(req.NpcIDs, row.ID)
		}
	}
	out := s.Probe(req)
	plan.Probe = &out
	return out
}

func (s *Service) spellsFileContains(ids []int) (int, string) {
	paths := []string{}
	if s.pathmgmt != nil {
		if dir := strings.TrimSpace(s.pathmgmt.GetExportDir()); dir != "" {
			paths = append(paths, filepath.Join(dir, "spells_us.txt"))
		}
	}
	if s.questsDirPath() != "" {
		paths = append(paths, filepath.Join(s.questsDirPath(), filepath.FromSlash(exportRel), "spells_us.txt"))
	}
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		text := string(raw)
		hit := 0
		for _, id := range ids {
			if strings.Contains(text, strconv.Itoa(id)+"^") || strings.HasPrefix(text, strconv.Itoa(id)+"^") {
				hit++
			}
		}
		return hit, p
	}
	return 0, ""
}

func (s *Service) probeZoneJSON(zoneID int, itemIDs []int) (int, string) {
	if s.pathmgmt == nil && s.questsDirPath() == "" {
		return -1, "no quests dir"
	}
	rel := fmt.Sprintf("%d_item.json", zoneID)
	// Sync already knows how to find zone JSON; count leftover -1 if we can read it.
	candidates := []string{}
	if s.questsDirPath() != "" {
		candidates = append(candidates,
			filepath.Join(s.questsDirPath(), "global", "ultimatedata", rel),
			filepath.Join(s.questsDirPath(), "global", "zone_controller", rel),
		)
	}
	for _, p := range candidates {
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		text := string(raw)
		neg := strings.Count(text, `"spellid": "-1"`) + strings.Count(text, `"spellid":-1`)
		hits := 0
		for _, id := range itemIDs {
			if strings.Contains(text, strconv.Itoa(id)) {
				hits++
			}
		}
		return hits, fmt.Sprintf("%s: %d minted item ids, %d leftover spellid -1", filepath.Base(p), hits, neg)
	}
	return 0, "zone item JSON not found (ok if this zone has no controller item file)"
}
