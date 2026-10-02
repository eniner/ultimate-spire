package zonecontroller

import (
	"sort"
	"strconv"
	"strings"
)

type ValidateIssue struct {
	ZoneID   int    `json:"zoneId"`
	Severity string `json:"severity"`
	Kind     string `json:"kind"`
	Message  string `json:"message"`
}

type ValidateReport struct {
	OK       bool            `json:"ok"`
	Error    string          `json:"error,omitempty"`
	Zones    int             `json:"zones"`
	Errors   int             `json:"errors"`
	Warnings int             `json:"warnings"`
	Issues   []ValidateIssue `json:"issues"`
}

func (s *Service) ValidateZones(ids []int) ValidateReport {
	out := ValidateReport{Issues: []ValidateIssue{}}
	if len(ids) == 0 {
		zones, err := s.ListZones()
		if err != nil {
			out.Error = err.Error()
			return out
		}
		for _, z := range zones {
			ids = append(ids, z.ZoneID)
		}
	}
	clean, err := normalizeZoneIDs(ids)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	out.Zones = len(clean)
	for _, id := range clean {
		out.Issues = append(out.Issues, s.validateOne(id)...)
	}
	for _, issue := range out.Issues {
		if issue.Severity == "error" {
			out.Errors++
		} else {
			out.Warnings++
		}
	}
	sort.Slice(out.Issues, func(i, j int) bool {
		if out.Issues[i].ZoneID != out.Issues[j].ZoneID {
			return out.Issues[i].ZoneID < out.Issues[j].ZoneID
		}
		if out.Issues[i].Severity != out.Issues[j].Severity {
			return out.Issues[i].Severity == "error"
		}
		return out.Issues[i].Kind < out.Issues[j].Kind
	})
	out.OK = true
	return out
}

func (s *Service) validateOne(id int) []ValidateIssue {
	issues := []ValidateIssue{}
	detail, err := s.GetZone(id)
	if err != nil {
		return []ValidateIssue{{ZoneID: id, Severity: "error", Kind: "json", Message: err.Error()}}
	}
	itemIDs := map[string]string{}
	for _, it := range detail.Items {
		name := strings.TrimSpace(it.Name)
		if name != "" {
			itemIDs[name] = strings.TrimSpace(it.ItemID)
		}
		if strings.TrimSpace(it.ItemID) == "" || it.ItemID == "-1" {
			issues = append(issues, ValidateIssue{ZoneID: id, Severity: "error", Kind: "item", Message: "item " + it.Name + " has no item id"})
		} else if n, _ := strconv.Atoi(it.ItemID); n > 0 && s.eqemu() != nil {
			var count int64
			s.eqemu().Table("items").Where("id = ?", n).Count(&count)
			if count == 0 {
				issues = append(issues, ValidateIssue{ZoneID: id, Severity: "error", Kind: "item", Message: "item id " + it.ItemID + " (" + it.Name + ") is not in items"})
			}
		}
	}
	for _, table := range detail.Loot {
		if len(table.Items) == 0 {
			issues = append(issues, ValidateIssue{ZoneID: id, Severity: "warn", Kind: "loot", Message: "loot table " + table.ID + " has no items"})
			continue
		}
		for _, name := range table.Items {
			if _, ok := itemIDs[name]; !ok {
				issues = append(issues, ValidateIssue{ZoneID: id, Severity: "error", Kind: "loot", Message: "loot table " + table.ID + " lists " + name + " which is not in item JSON"})
			}
		}
	}
	info, err := s.zoneInfo(id)
	if err != nil {
		issues = append(issues, ValidateIssue{ZoneID: id, Severity: "warn", Kind: "zone", Message: err.Error()})
	} else {
		if !s.zoneHasController(info.short) {
			issues = append(issues, ValidateIssue{ZoneID: id, Severity: "error", Kind: "controller", Message: "no zone controller NPC spawned in " + info.short})
		}
		if names, err := s.zoneSpawnNames(info.short); err == nil {
			have := map[string]bool{}
			for _, n := range names {
				have[strings.ToLower(n)] = true
			}
			for _, mob := range detail.Custom {
				if !have[strings.ToLower(mob.Name)] {
					issues = append(issues, ValidateIssue{ZoneID: id, Severity: "warn", Kind: "custom", Message: "custom mob " + mob.Name + " is not in spawn2"})
				}
			}
		}
	}
	if detail.CustomCount == 0 {
		issues = append(issues, ValidateIssue{ZoneID: id, Severity: "warn", Kind: "custom", Message: "no custom mobs; every spawn will use trash basedata"})
	}
	if detail.ItemCount == 0 {
		issues = append(issues, ValidateIssue{ZoneID: id, Severity: "warn", Kind: "gear", Message: "item JSON is empty"})
	}
	return issues
}

func (s *Service) zoneSpawnNames(short string) ([]string, error) {
	if s.eqemu() == nil {
		return nil, nil
	}
	var names []string
	err := s.eqemu().Table("npc_types").
		Select("DISTINCT npc_types.name").
		Joins("JOIN spawnentry ON spawnentry.npcID = npc_types.id").
		Joins("JOIN spawn2 ON spawn2.spawngroupID = spawnentry.spawngroupID").
		Where("spawn2.zone = ?", short).
		Pluck("npc_types.name", &names).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(names))
	for _, n := range names {
		if cleaned := cleanMobName(n); cleaned != "" {
			out = append(out, cleaned)
		}
	}
	return out, nil
}
