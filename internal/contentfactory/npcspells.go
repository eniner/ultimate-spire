package contentfactory

import (
	"fmt"

	"github.com/EQEmu/spire/internal/models"
	"github.com/volatiletech/null/v8"
)

type NpcSpellRequest struct {
	DryRun       bool       `json:"dryRun"`
	SkipJournal  bool       `json:"skipJournal"`
	SourceID     int        `json:"sourceId"`
	Prefix       string     `json:"prefix"`
	StartListID  int        `json:"startListId"`
	StartSpellID int        `json:"startSpellId"`
	CloneSpells  *bool      `json:"cloneSpells"`
	Scale        SpellScale `json:"scale"`
	AttachNpcIDs []int      `json:"attachNpcIds"`
	ZoneIDs      []int      `json:"zoneIds"`
	NameContains string            `json:"nameContains"`
	Target       TargetFilter      `json:"target"`
	SkipAttach   bool              `json:"skipAttach"`
	Entries      []SpellSetEntryIn `json:"entries"`
}

func (s *Service) RunNpcSpells(req NpcSpellRequest) Plan {
	plan := emptyPlan("npc_spells", req.DryRun)
	if s.eqemu() == nil {
		return plan.fail("PEQ database is not connected")
	}
	if req.SourceID <= 0 {
		return plan.fail("source npc_spells id is required")
	}
	var src models.NpcSpell
	if err := s.eqemu().Table("npc_spells").Where("id = ?", req.SourceID).Take(&src).Error; err != nil {
		return plan.fail(fmt.Sprintf("npc_spells %d not found", req.SourceID))
	}
	var entries []models.NpcSpellsEntry
	if err := s.eqemu().Table("npc_spells_entries").Where("npc_spells_id = ?", req.SourceID).Find(&entries).Error; err != nil {
		return plan.fail(err.Error())
	}
	if len(req.Entries) > 0 {
		entries = entries[:0]
		for _, e := range req.Entries {
			if e.SpellID <= 0 {
				continue
			}
			maxLvl := e.MaxLevel
			if maxLvl <= 0 {
				maxLvl = 255
			}
			entries = append(entries, models.NpcSpellsEntry{
				Spellid: uint16(e.SpellID), Type: uint(e.Type),
				Minlevel: uint8(e.MinLevel), Maxlevel: uint8(maxLvl),
				Manacost: int16(e.Manacost), RecastDelay: e.RecastDelay, Priority: int16(e.Priority),
			})
		}
	}
	cloneSpells := flagOr(req.CloneSpells, true)
	spellIDs := uniqueInts(func() []int {
		out := make([]int, 0, len(entries))
		for _, e := range entries {
			out = append(out, int(e.Spellid))
		}
		return out
	}())
	remap := map[int]int{}
	if cloneSpells && len(spellIDs) > 0 {
		taken, err := s.takeIDsCapped("spells_new", "id", req.StartSpellID, len(spellIDs), npcCastSpellFloor, npcSpellIDCap, &plan)
		if err != nil {
			return plan.fail(err.Error())
		}
		var spells []models.SpellsNew
		if err := s.eqemu().Table("spells_new").Where("id IN ?", spellIDs).Find(&spells).Error; err != nil {
			return plan.fail(err.Error())
		}
		byID := map[int]models.SpellsNew{}
		for _, row := range spells {
			byID[row.ID] = row
		}
		for i, oldID := range spellIDs {
			srcSpell, ok := byID[oldID]
			if !ok {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("spell %d not found; entry will keep the old id", oldID))
				remap[oldID] = oldID
				continue
			}
			newID := taken[i]
			if newID > npcSpellIDCap {
				return plan.fail(fmt.Sprintf("cloned spell %d is above %d", newID, npcSpellIDCap))
			}
			name := prefixedName(req.Prefix, srcSpell.Name.String)
			dst := srcSpell
			dst.ID = newID
			dst.Name = null.StringFrom(name)
			dst.Items = nil
			dst.NpcSpellsEntries = nil
			dst.BotSpellsEntries = nil
			remap[oldID] = newID
			plan.Clones = append(plan.Clones, Clone{Kind: "spell", SourceID: oldID, NewID: newID, Name: name})
			plan.DBWrites = append(plan.DBWrites, DBWrite{Table: "spells_new", Action: "create", ID: newID, Note: name})
			if req.DryRun {
				continue
			}
			if err := s.eqemu().Table("spells_new").Create(&dst).Error; err != nil {
				return plan.fail(fmt.Sprintf("create spell %d: %v", newID, err))
			}
		}
	} else {
		for _, id := range spellIDs {
			remap[id] = id
		}
	}

	listIDs, err := s.takeIDs("npc_spells", "id", req.StartListID, 1, reservedFloor, &plan)
	if err != nil {
		return plan.fail(err.Error())
	}
	listID := listIDs[0]
	listName := prefixedName(req.Prefix, firstNonEmpty(src.Name.String, fmt.Sprintf("set %d", src.ID)))
	dst := src
	dst.ID = uint(listID)
	dst.Name = null.StringFrom(listName)
	dst.NpcSpellsEntries = nil
	dst.BotSpellsEntries = nil
	dst.NpcSpell = nil
	plan.Clones = append(plan.Clones, Clone{Kind: "npc_spells", SourceID: req.SourceID, NewID: listID, Name: listName})
	plan.DBWrites = append(plan.DBWrites, DBWrite{Table: "npc_spells", Action: "create", ID: listID, Note: listName})
	if !req.DryRun {
		if err := s.eqemu().Table("npc_spells").Create(&dst).Error; err != nil {
			return plan.fail(fmt.Sprintf("create npc_spells %d: %v", listID, err))
		}
	}
	for _, entry := range entries {
		newSpell := remap[int(entry.Spellid)]
		if newSpell <= 0 {
			newSpell = int(entry.Spellid)
		}
		if newSpell > npcSpellIDCap {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("entry spell %d is above %d and will not fit", newSpell, npcSpellIDCap))
			continue
		}
		plan.DBWrites = append(plan.DBWrites, DBWrite{
			Table: "npc_spells_entries", Action: "create", ID: newSpell,
			Note:  fmt.Sprintf("list %d spell %d", listID, newSpell),
			Extra: map[string]int{"npcSpellsId": listID, "spellId": newSpell},
		})
		if req.DryRun {
			continue
		}
		row := entry
		row.ID = 0
		row.NpcSpellsId = listID
		row.Spellid = uint16(newSpell)
		row.SpellsNew = nil
		if err := s.eqemu().Table("npc_spells_entries").Create(&row).Error; err != nil {
			return plan.fail(fmt.Sprintf("create npc_spells_entries %d/%d: %v", listID, newSpell, err))
		}
	}

	if !req.SkipAttach {
		if req.Target.NameContains == "" {
			req.Target.NameContains = req.NameContains
		}
		npcIDs, classified, err := s.npcIDsFiltered(req.ZoneIDs, req.AttachNpcIDs, req.Target)
		if err != nil {
			plan.Warnings = append(plan.Warnings, err.Error())
		} else if len(npcIDs) > 0 {
			roles := map[int]string{}
			for _, n := range classified {
				roles[n.ID] = n.Role
			}
			if err := s.setNpcSpells(npcIDs, listID, req.DryRun, roles, &plan); err != nil {
				return plan.fail(err.Error())
			}
		}
	}
	s.complete(&plan, req.DryRun, req.SkipJournal)
	return plan
}

func lastCloneID(plan Plan, kind string) int {
	for i := len(plan.Clones) - 1; i >= 0; i-- {
		if plan.Clones[i].Kind == kind {
			return plan.Clones[i].NewID
		}
	}
	return 0
}

func cloneIDs(plan Plan, kind string) []int {
	out := []int{}
	for _, c := range plan.Clones {
		if c.Kind == kind && c.NewID > 0 {
			out = append(out, c.NewID)
		}
	}
	return uniqueInts(out)
}
