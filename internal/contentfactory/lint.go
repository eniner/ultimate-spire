package contentfactory

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type LintItem struct {
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
	ID      int    `json:"id,omitempty"`
}

func (s *Service) lintPlan(plan *Plan) {
	if plan == nil {
		return
	}
	if plan.Lint == nil {
		plan.Lint = []LintItem{}
	}
	seen := map[string]bool{}
	add := func(level, code string, id int, msg string) {
		key := fmt.Sprintf("%s:%s:%d:%s", level, code, id, msg)
		if seen[key] {
			return
		}
		seen[key] = true
		plan.Lint = append(plan.Lint, LintItem{Level: level, Code: code, ID: id, Message: msg})
	}
	for _, c := range plan.Clones {
		if utf8.RuneCountInString(c.Name) > 64 {
			add("warning", "name_len", c.NewID, fmt.Sprintf("%s %q is longer than 64 characters", c.Kind, c.Name))
		}
		if c.Kind == "item" && (c.Slots == 0 || c.Classes == 0) {
			add("warning", "item_bits", c.NewID, fmt.Sprintf("item %d has slot/class 0", c.NewID))
		}
		if c.Kind == "spell" && c.NewID > npcSpellIDCap {
			add("error", "spellid_cap", c.NewID, fmt.Sprintf("spell %d is above %d and will not fit npc_spells_entries", c.NewID, npcSpellIDCap))
		}
	}
	for _, w := range plan.DBWrites {
		if w.Table == "npc_spells_entries" && w.ID > npcSpellIDCap {
			add("error", "spellid_cap", w.ID, fmt.Sprintf("entry spell %d is above %d", w.ID, npcSpellIDCap))
		}
		if w.Table == "npc_types" && w.Action == "update" && w.ID > 0 && w.ID < reservedFloor {
			add("warning", "vanilla_npc", w.ID, fmt.Sprintf("updates stock PEQ npc %d; clone NPCs to keep vanilla rows", w.ID))
		}
		if w.Table == "spawnentry" && w.Action == "update" && extra(w, "oldNpcId") > 0 && extra(w, "oldNpcId") < reservedFloor {
			add("warning", "retarget_spawn", w.ID, fmt.Sprintf("spawnentry will point at clone %d instead of stock npc %d", extra(w, "newNpcId"), extra(w, "oldNpcId")))
		}
	}
	needSpells := []int{}
	for _, c := range plan.Clones {
		for _, d := range c.Diff {
			if (d.Field == "click" || d.Field == "proc") && d.To != "" && d.To != "0" && d.To != "-1" {
				if n, err := strconv.Atoi(d.To); err == nil && n > 0 {
					needSpells = append(needSpells, n)
				}
			}
		}
	}
	if s.eqemu() != nil && len(needSpells) > 0 {
		names := s.spellNames(needSpells)
		for _, id := range uniqueInts(needSpells) {
			if names[id] == "" {
				add("error", "missing_spell", id, fmt.Sprintf("item points at missing spell %d", id))
			}
		}
	}
	if s.eqemu() != nil {
		for _, c := range plan.Clones {
			if c.Kind != "item" || c.NewID <= 0 || c.SourceID <= 0 {
				continue
			}
			type row struct {
				Slots, Classes, Clickeffect, Proceffect int
			}
			var src row
			if err := s.eqemu().Table("items").Select("slots, classes, clickeffect, proceffect").Where("id = ?", c.SourceID).Take(&src).Error; err != nil {
				continue
			}
			if src.Clickeffect > 0 {
				needSpells = append(needSpells, src.Clickeffect)
			}
			if src.Proceffect > 0 {
				needSpells = append(needSpells, src.Proceffect)
			}
		}
		names := s.spellNames(needSpells)
		for _, id := range uniqueInts(needSpells) {
			if id > 0 && names[id] == "" {
				add("error", "missing_spell", id, fmt.Sprintf("source click/proc spell %d is missing", id))
			}
		}
	}
}

func lintHasError(plan Plan) bool {
	for _, item := range plan.Lint {
		if item.Level == "error" {
			return true
		}
	}
	return false
}

func (s *Service) complete(plan *Plan, dry, skipJournal bool) {
	if plan == nil {
		return
	}
	s.lintPlan(plan)
	if lintHasError(*plan) && !dry {
		plan.fail("lint errors — preview only until they are fixed")
		return
	}
	if !dry && !skipJournal && plan.Error == "" {
		s.saveRun(plan)
	}
	if plan.Error == "" {
		plan.OK = true
	}
}

func diffInt(field string, from, to int) DiffField {
	if from == to {
		return DiffField{}
	}
	return DiffField{Field: field, From: strconv.Itoa(from), To: strconv.Itoa(to)}
}

func diffI64(field string, from, to int64) DiffField {
	if from == to {
		return DiffField{}
	}
	return DiffField{Field: field, From: strconv.FormatInt(from, 10), To: strconv.FormatInt(to, 10)}
}

func compactDiff(in []DiffField) []DiffField {
	out := []DiffField{}
	for _, d := range in {
		if strings.TrimSpace(d.Field) == "" || d.From == d.To {
			continue
		}
		out = append(out, d)
	}
	return out
}
