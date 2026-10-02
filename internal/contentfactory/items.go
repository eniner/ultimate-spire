package contentfactory

import (
	"fmt"
	"strings"

	"github.com/EQEmu/spire/internal/models"
)

type KitCell struct {
	Slot         string `json:"slot"`
	Class        string `json:"class"`
	SourceID     int    `json:"sourceId"`
	Name         string `json:"name"`
	Slots        int    `json:"slots"`
	Classes      int    `json:"classes"`
	ClickSpellID int    `json:"clickSpellId"`
	ProcSpellID  int    `json:"procSpellId"`
	WornSpellID  int    `json:"wornSpellId"`
	FocusSpellID int    `json:"focusSpellId"`
}

type ItemRequest struct {
	DryRun     bool      `json:"dryRun"`
	StartID    int       `json:"startId"`
	Prefix     string    `json:"prefix"`
	Count      int       `json:"count"`
	Step       float64   `json:"step"`
	Cells      []KitCell `json:"cells"`
	RecipeID    string    `json:"recipeId"`
	RecipeName  string    `json:"recipeName"`
	Group       string    `json:"group"`
	SkipJournal bool      `json:"skipJournal"`
}

func (s *Service) RunItems(req ItemRequest) Plan {
	plan := emptyPlan("items", req.DryRun)
	if s.eqemu() == nil {
		return plan.fail("PEQ database is not connected")
	}
	cells := normalizeCells(req.Cells)
	if len(cells) == 0 {
		return plan.fail("fill at least one kit cell or source item")
	}
	if req.Count < 1 {
		req.Count = 1
	}
	if req.Count > 12 {
		return plan.fail("count must be 12 or less")
	}
	if req.Step <= 0 {
		req.Step = 1
	}
	ids := uniqueInts(func() []int {
		out := make([]int, 0, len(cells))
		for _, cell := range cells {
			out = append(out, cell.SourceID)
		}
		return out
	}())
	byID, err := s.loadItems(ids)
	if err != nil {
		return plan.fail(err.Error())
	}
	need := len(cells) * req.Count
	taken, err := s.takeIDs("items", "id", req.StartID, need, reservedFloor, &plan)
	if err != nil {
		return plan.fail(err.Error())
	}
	cursor := 0
	for step := 0; step < req.Count; step++ {
		mult := ladderMultiplier(req.Step, step)
		prefix := strings.TrimSpace(req.Prefix)
		if req.Count > 1 {
			if prefix == "" {
				prefix = fmt.Sprintf("T%d", step+1)
			} else {
				prefix = fmt.Sprintf("%s T%d", prefix, step+1)
			}
		}
		for _, cell := range cells {
			src, ok := byID[cell.SourceID]
			if !ok {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("source item %d not found", cell.SourceID))
				continue
			}
			if cursor >= len(taken) {
				return plan.fail("ran out of reserved item ids")
			}
			newID := taken[cursor]
			cursor++
			name := firstNonEmpty(cell.Name, src.Name)
			name = prefixedName(prefix, name)
			dst := src
			dst.ID = newID
			dst.Name = name
			if strings.TrimSpace(dst.Lore) != "" && prefix != "" && !strings.HasPrefix(dst.Lore, prefix) {
				dst.Lore = prefixedName(prefix, dst.Lore)
			}
			if bits := firstPositive(cell.Slots, slotBits(cell.Slot)); bits > 0 {
				dst.Slots = bits
			}
			if bits := firstPositive(cell.Classes, classBit(cell.Class)); bits > 0 {
				dst.Classes = bits
			}
			if cell.ClickSpellID > 0 {
				dst.Clickeffect = cell.ClickSpellID
			}
			if cell.ProcSpellID > 0 {
				dst.Proceffect = cell.ProcSpellID
			}
			if cell.WornSpellID > 0 {
				dst.Worneffect = cell.WornSpellID
			}
			if cell.FocusSpellID > 0 {
				dst.Focuseffect = cell.FocusSpellID
			}
			scaleItemStats(&dst, mult)
			clone := Clone{
				Kind:     "item",
				SourceID: src.ID,
				NewID:    newID,
				Name:     name,
				Step:     step,
				Slot:     firstNonEmpty(cell.Slot, slotName(dst.Slots)),
				Class:    firstNonEmpty(cell.Class, className(dst.Classes)),
				Slots:    dst.Slots,
				Classes:  dst.Classes,
				Diff: compactDiff([]DiffField{
					diffInt("ac", src.Ac, dst.Ac),
					diffInt("hp", src.Hp, dst.Hp),
					diffInt("mana", src.Mana, dst.Mana),
					diffInt("damage", src.Damage, dst.Damage),
					diffInt("delay", src.Delay, dst.Delay),
					diffInt("click", src.Clickeffect, dst.Clickeffect),
					diffInt("proc", src.Proceffect, dst.Proceffect),
				}),
			}
			plan.Clones = append(plan.Clones, clone)
			plan.DBWrites = append(plan.DBWrites, DBWrite{Table: "items", Action: "create", ID: newID, Note: name})
			if req.DryRun {
				continue
			}
			if err := s.eqemu().Table("items").Create(&dst).Error; err != nil {
				return plan.fail(fmt.Sprintf("create item %d: %v", newID, err))
			}
		}
	}
	if strings.TrimSpace(req.RecipeID) != "" {
		if err := s.exportItemRecipe(req, plan.Clones, req.DryRun, &plan); err != nil {
			plan.Warnings = append(plan.Warnings, err.Error())
		} else {
			plan.RecipeID = strings.TrimSpace(req.RecipeID)
		}
	}
	s.complete(&plan, req.DryRun, req.SkipJournal)
	return plan
}

func normalizeCells(cells []KitCell) []KitCell {
	out := []KitCell{}
	seen := map[string]bool{}
	for _, cell := range cells {
		if cell.SourceID <= 0 {
			continue
		}
		key := fmt.Sprintf("%d|%s|%s|%s", cell.SourceID, cell.Slot, cell.Class, cell.Name)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, cell)
	}
	return out
}

func firstPositive(vals ...int) int {
	for _, v := range vals {
		if v > 0 {
			return v
		}
	}
	return 0
}

func (s *Service) loadItems(ids []int) (map[int]models.Item, error) {
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
