package contentfactory

import (
	"fmt"
	"strings"

	"github.com/EQEmu/spire/internal/models"
	"github.com/volatiletech/null/v8"
)

type SpellSource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SpellScale struct {
	Effects  *bool `json:"effects"`
	Mana     *bool `json:"mana"`
	Duration *bool `json:"duration"`
	Recast   *bool `json:"recast"`
	Max      *bool `json:"max"`
}

type SpellAttach struct {
	ItemID int    `json:"itemId"`
	Field  string `json:"field"`
}

type SpellRequest struct {
	DryRun       bool          `json:"dryRun"`
	SkipJournal  bool          `json:"skipJournal"`
	StartID      int           `json:"startId"`
	Prefix       string        `json:"prefix"`
	Count        int           `json:"count"`
	Step         float64       `json:"step"`
	NpcCastable  bool          `json:"npcCastable"`
	ExportClient bool          `json:"exportClient"`
	ExportDbStr  bool          `json:"exportDbstr"`
	Sources      []SpellSource `json:"sources"`
	Scale        SpellScale    `json:"scale"`
	Attach       []SpellAttach `json:"attach"`
}

func flagOr(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}

func (s *Service) RunSpells(req SpellRequest) Plan {
	plan := emptyPlan("spells", req.DryRun)
	if s.eqemu() == nil {
		return plan.fail("PEQ database is not connected")
	}
	if req.Count < 1 {
		req.Count = 1
	}
	if req.Count > 20 {
		return plan.fail("count must be 20 or less")
	}
	if req.Step <= 0 {
		req.Step = 1
	}
	ids := uniqueInts(func() []int {
		out := make([]int, 0, len(req.Sources))
		for _, src := range req.Sources {
			out = append(out, src.ID)
		}
		return out
	}())
	if len(ids) == 0 {
		return plan.fail("add at least one source spell")
	}
	need := len(ids) * req.Count
	var taken []int
	var err error
	if req.NpcCastable {
		taken, err = s.takeIDsCapped("spells_new", "id", req.StartID, need, npcCastSpellFloor, npcSpellIDCap, &plan)
	} else {
		taken, err = s.takeIDs("spells_new", "id", req.StartID, need, reservedFloor, &plan)
	}
	if err != nil {
		return plan.fail(err.Error())
	}
	var rows []models.SpellsNew
	if err := s.eqemu().Table("spells_new").Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return plan.fail(err.Error())
	}
	byID := map[int]models.SpellsNew{}
	for _, row := range rows {
		byID[row.ID] = row
	}
	wantEffects := flagOr(req.Scale.Effects, true)
	wantMana := flagOr(req.Scale.Mana, true)
	wantDur := flagOr(req.Scale.Duration, true)
	wantRecast := flagOr(req.Scale.Recast, false)
	wantMax := flagOr(req.Scale.Max, true)
	cursor := 0
	created := map[int]int{}
	for _, srcReq := range req.Sources {
		src, ok := byID[srcReq.ID]
		if !ok {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("source spell %d not found", srcReq.ID))
			continue
		}
		baseName := firstNonEmpty(srcReq.Name, src.Name.String)
		for step := 0; step < req.Count; step++ {
			if cursor >= len(taken) {
				return plan.fail("ran out of reserved spell ids")
			}
			newID := taken[cursor]
			cursor++
			if newID > npcSpellIDCap {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("spell %d is above %d and cannot be used in npc_spells_entries", newID, npcSpellIDCap))
			}
			mult := ladderMultiplier(req.Step, step)
			name := prefixedName(req.Prefix, baseName)
			if req.Count > 1 {
				name = prefixedName(fmt.Sprintf("%sS%d", strings.TrimSpace(req.Prefix), step+1), baseName)
				if strings.TrimSpace(req.Prefix) == "" {
					name = fmt.Sprintf("S%d %s", step+1, baseName)
				}
			}
			dst := src
			dst.ID = newID
			dst.Name = null.StringFrom(name)
			dst.Items = nil
			dst.NpcSpellsEntries = nil
			dst.BotSpellsEntries = nil
			if wantEffects {
				scaleInt(&dst.EffectBaseValue1, mult)
				scaleInt(&dst.EffectBaseValue2, mult)
				scaleInt(&dst.EffectBaseValue3, mult)
				scaleInt(&dst.EffectBaseValue4, mult)
				scaleInt(&dst.EffectBaseValue5, mult)
				scaleInt(&dst.EffectBaseValue6, mult)
				scaleInt(&dst.EffectBaseValue7, mult)
				scaleInt(&dst.EffectBaseValue8, mult)
				scaleInt(&dst.EffectBaseValue9, mult)
				scaleInt(&dst.EffectBaseValue10, mult)
				scaleInt(&dst.EffectBaseValue11, mult)
				scaleInt(&dst.EffectBaseValue12, mult)
			}
			if wantMax {
				scaleInt(&dst.Max1, mult)
				scaleInt(&dst.Max2, mult)
				scaleInt(&dst.Max3, mult)
				scaleInt(&dst.Max4, mult)
				scaleInt(&dst.Max5, mult)
				scaleInt(&dst.Max6, mult)
				scaleInt(&dst.Max7, mult)
				scaleInt(&dst.Max8, mult)
				scaleInt(&dst.Max9, mult)
				scaleInt(&dst.Max10, mult)
				scaleInt(&dst.Max11, mult)
				scaleInt(&dst.Max12, mult)
			}
			if wantMana {
				scalePositive(&dst.Mana, mult)
			}
			if wantDur {
				scalePositive(&dst.Buffduration, mult)
			}
			if wantRecast {
				scalePositive(&dst.RecastTime, mult)
			}
			clone := Clone{Kind: "spell", SourceID: src.ID, NewID: newID, Name: name, Step: step,
				Diff: compactDiff([]DiffField{
					diffInt("mana", src.Mana, dst.Mana),
					diffInt("duration", src.Buffduration, dst.Buffduration),
					diffInt("recast", src.RecastTime, dst.RecastTime),
					diffInt("effect1", src.EffectBaseValue1, dst.EffectBaseValue1),
				}),
			}
			plan.Clones = append(plan.Clones, clone)
			plan.DBWrites = append(plan.DBWrites, DBWrite{Table: "spells_new", Action: "create", ID: newID, Note: name})
			created[src.ID] = newID
			if req.DryRun {
				continue
			}
			if err := s.eqemu().Table("spells_new").Create(&dst).Error; err != nil {
				return plan.fail(fmt.Sprintf("create spell %d: %v", newID, err))
			}
		}
	}
	if err := s.attachSpells(req.Attach, created, req.DryRun, &plan); err != nil {
		return plan.fail(err.Error())
	}
	if req.ExportClient && !req.DryRun {
		info := s.ExportClient(req.ExportDbStr)
		plan.Export = &info
		if !info.OK {
			plan.Warnings = append(plan.Warnings, firstNonEmpty(info.Note, "client export failed"))
		}
	}
	s.complete(&plan, req.DryRun, req.SkipJournal)
	return plan
}

func spellItemField(field string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "click", "clickeffect":
		return "clickeffect", nil
	case "proc", "proceffect":
		return "proceffect", nil
	case "worn", "worneffect":
		return "worneffect", nil
	case "scroll", "scrolleffect":
		return "scrolleffect", nil
	case "focus", "focuseffect":
		return "focuseffect", nil
	default:
		return "", fmt.Errorf("unknown item spell field %q", field)
	}
}

func (s *Service) attachSpells(rows []SpellAttach, created map[int]int, dry bool, plan *Plan) error {
	if len(rows) == 0 {
		return nil
	}
	for _, row := range rows {
		if row.ItemID <= 0 {
			continue
		}
		col, err := spellItemField(row.Field)
		if err != nil {
			plan.Warnings = append(plan.Warnings, err.Error())
			continue
		}
		spellID := 0
		if len(plan.Clones) > 0 {
			spellID = plan.Clones[len(plan.Clones)-1].NewID
		}
		if len(created) == 1 {
			for _, id := range created {
				spellID = id
			}
		}
		if len(plan.Clones) > 0 && len(created) > 1 {
			spellID = plan.Clones[0].NewID
		}
		if spellID <= 0 {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("no cloned spell to attach to item %d", row.ItemID))
			continue
		}
		plan.DBWrites = append(plan.DBWrites, DBWrite{
			Table:  "items",
			Action: "update",
			ID:     row.ItemID,
			Note:   fmt.Sprintf("%s = %d", col, spellID),
			Extra:  map[string]int{"spellId": spellID},
		})
		plan.Attached++
		if dry {
			continue
		}
		if err := s.eqemu().Table("items").Where("id = ?", row.ItemID).Update(col, spellID).Error; err != nil {
			return fmt.Errorf("attach item %d: %w", row.ItemID, err)
		}
	}
	return nil
}
