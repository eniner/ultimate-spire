package zonecontroller

import (
	"fmt"
	"strings"

	"github.com/EQEmu/spire/internal/models"
)

type ItemClone struct {
	SourceID int    `json:"sourceId"`
	NewID    int    `json:"newId"`
	Name     string `json:"name"`
	Step     int    `json:"step"`
	SpellID  int    `json:"spellId,omitempty"`
	ClickID  int    `json:"clickId,omitempty"`
	ProcID   int    `json:"procId,omitempty"`
}

type DBWrite struct {
	Table  string `json:"table"`
	Action string `json:"action"`
	ID     int    `json:"id"`
	Note   string `json:"note,omitempty"`
}

func (s *Service) loadItemsByID(ids []int) (map[int]models.Item, error) {
	out := map[int]models.Item{}
	if s.eqemu() == nil || len(ids) == 0 {
		return out, nil
	}
	var rows []models.Item
	if err := s.eqemu().Table("items").Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.ID] = row
	}
	return out, nil
}

func (s *Service) nextItemID(start int) (int, error) {
	if start > 0 {
		return start, nil
	}
	if s.eqemu() == nil {
		return 800000, nil
	}
	var max int
	if err := s.eqemu().Table("items").Select("COALESCE(MAX(id), 0)").Scan(&max).Error; err != nil {
		return 0, err
	}
	if max < 800000 {
		return 800000, nil
	}
	return max + 1, nil
}

func (s *Service) existingItemIDs(start, end int) (map[int]bool, error) {
	out := map[int]bool{}
	if s.eqemu() == nil || end < start {
		return out, nil
	}
	var ids []int
	if err := s.eqemu().Table("items").Where("id BETWEEN ? AND ?", start, end).Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

func (s *Service) cloneKitItems(kit RecipeKit, startID int, prefix string, stepIndex int, mult float64, dry bool, plan *FactoryPlan) ([]ItemClone, map[string]int, error) {
	clones := []ItemClone{}
	nameToID := map[string]int{}
	if len(kit.Items) == 0 {
		return clones, nameToID, nil
	}
	ids := make([]int, 0, len(kit.Items))
	for _, it := range kit.Items {
		if it.ItemID > 0 {
			ids = append(ids, it.ItemID)
		}
	}
	if len(ids) == 0 {
		return clones, nameToID, nil
	}
	byID, err := s.loadItemsByID(ids)
	if err != nil {
		return nil, nil, err
	}
	end := startID + len(kit.Items) - 1
	exists, err := s.existingItemIDs(startID, end)
	if err != nil {
		return nil, nil, err
	}
	next := startID
	for _, it := range kit.Items {
		src, ok := byID[it.ItemID]
		if !ok {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("source item %d not found", it.ItemID))
			continue
		}
		for exists[next] {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("item id %d already exists; skipped", next))
			next++
		}
		name := prefixedName(prefix, firstNonEmpty(it.Name, src.Name))
		dst := src
		dst.ID = next
		dst.Name = name
		if strings.TrimSpace(dst.Lore) != "" && !strings.HasPrefix(dst.Lore, prefix) {
			dst.Lore = prefixedName(prefix, dst.Lore)
		}
		scaleItemStats(&dst, mult)
		spellID := dst.Clickeffect
		if spellID <= 0 {
			spellID = dst.Proceffect
		}
		clone := ItemClone{SourceID: src.ID, NewID: next, Name: name, Step: stepIndex, SpellID: spellID, ClickID: dst.Clickeffect, ProcID: dst.Proceffect}
		clones = append(clones, clone)
		nameToID[name] = next
		if it.Name != "" {
			nameToID[it.Name] = next
		}
		plan.DBWrites = append(plan.DBWrites, DBWrite{Table: "items", Action: "create", ID: next, Note: name})
		if !dry {
			if err := s.eqemu().Table("items").Create(&dst).Error; err != nil {
				return clones, nameToID, fmt.Errorf("create item %d: %w", next, err)
			}
		}
		next++
	}
	return clones, nameToID, nil
}

func buildItemJSON(clones []ItemClone) map[string]interface{} {
	out := map[string]interface{}{}
	for _, c := range clones {
		spell := "-1"
		if c.SpellID > 0 {
			spell = fmt.Sprintf("%d", c.SpellID)
		}
		out[c.Name] = map[string]interface{}{
			"itemid":      fmt.Sprintf("%d", c.NewID),
			"spellid":     spell,
			"globalkey":   "-1",
			"isglobal":    "-1",
			"description": "",
		}
	}
	return out
}

func buildLootJSON(tables []RecipeLootTable, nameToID map[string]int, prefix string) map[string]interface{} {
	out := map[string]interface{}{}
	for _, table := range tables {
		items := make([]interface{}, 0, len(table.Items))
		for _, name := range table.Items {
			next := prefixedName(prefix, name)
			if _, ok := nameToID[next]; !ok {
				if _, ok := nameToID[name]; ok {
					next = name
				}
			}
			items = append(items, next)
		}
		out[table.ID] = items
	}
	return out
}

func lootFieldToRows(raw string) []interface{} {
	out := []interface{}{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id := part
		chance := "100"
		if i := strings.LastIndex(part, "@"); i >= 0 {
			id = strings.TrimSpace(part[:i])
			if c := strings.TrimSpace(part[i+1:]); c != "" {
				chance = c
			}
		}
		if id == "" {
			continue
		}
		out = append(out, map[string]interface{}{"id": id, "chance": chance})
	}
	return out
}

func applyBasedata(zone map[string]interface{}, scaled map[string]map[string]string) {
	basedata := asObject(zone["basedata"])
	for _, kind := range []string{"trash", "boss", "raid"} {
		row := asObject(basedata[kind])
		for k, v := range scaled[kind] {
			if k == "loot" {
				if rows := lootFieldToRows(v); len(rows) > 0 {
					row["loot"] = rows
				}
				continue
			}
			if strings.TrimSpace(v) == "" {
				continue
			}
			row[k] = coerceJSONValue(v)
		}
		basedata[kind] = row
	}
	zone["basedata"] = basedata
}
