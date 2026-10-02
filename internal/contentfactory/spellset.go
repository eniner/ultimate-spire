package contentfactory

import (
	"fmt"
	"strings"

	"github.com/EQEmu/spire/internal/models"
	"github.com/volatiletech/null/v8"
)

type SpellSetEntryIn struct {
	SpellID     int    `json:"spellId"`
	Name        string `json:"name,omitempty"`
	MinLevel    int    `json:"minLevel"`
	MaxLevel    int    `json:"maxLevel"`
	Type        int    `json:"type"`
	Manacost    int    `json:"manacost"`
	RecastDelay int    `json:"recastDelay"`
	Priority    int    `json:"priority"`
}

type SpellSetView struct {
	OK      bool              `json:"ok"`
	Error   string            `json:"error,omitempty"`
	ID      int               `json:"id"`
	Name    string            `json:"name"`
	Entries []SpellSetEntryIn `json:"entries"`
}

type SpellSetSaveRequest struct {
	DryRun  bool              `json:"dryRun"`
	ID      int               `json:"id"`
	Name    string            `json:"name"`
	Prefix  string            `json:"prefix"`
	Entries []SpellSetEntryIn `json:"entries"`
	AsNew   bool              `json:"asNew"`
}

func (s *Service) GetSpellSet(id int) SpellSetView {
	out := SpellSetView{ID: id, Entries: []SpellSetEntryIn{}}
	if s.eqemu() == nil {
		out.Error = "PEQ database is not connected"
		return out
	}
	if id <= 0 {
		out.Error = "spell set id is required"
		return out
	}
	var hdr models.NpcSpell
	if err := s.eqemu().Table("npc_spells").Where("id = ?", id).Take(&hdr).Error; err != nil {
		out.Error = fmt.Sprintf("npc_spells %d not found", id)
		return out
	}
	out.Name = hdr.Name.String
	var rows []models.NpcSpellsEntry
	_ = s.eqemu().Table("npc_spells_entries").Where("npc_spells_id = ?", id).Order("minlevel, priority desc").Find(&rows).Error
	ids := []int{}
	for _, r := range rows {
		ids = append(ids, int(r.Spellid))
	}
	names := s.spellNames(ids)
	for _, r := range rows {
		out.Entries = append(out.Entries, SpellSetEntryIn{
			SpellID: int(r.Spellid), Name: names[int(r.Spellid)],
			MinLevel: int(r.Minlevel), MaxLevel: int(r.Maxlevel), Type: int(r.Type),
			Manacost: int(r.Manacost), RecastDelay: r.RecastDelay, Priority: int(r.Priority),
		})
	}
	out.OK = true
	return out
}

func (s *Service) SaveSpellSet(req SpellSetSaveRequest) Plan {
	plan := emptyPlan("spell_set", req.DryRun)
	if s.eqemu() == nil {
		return plan.fail("PEQ database is not connected")
	}
	entries := []SpellSetEntryIn{}
	for _, e := range req.Entries {
		if e.SpellID <= 0 {
			continue
		}
		if e.MaxLevel <= 0 {
			e.MaxLevel = 255
		}
		if e.SpellID > npcSpellIDCap {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("spell %d is above %d", e.SpellID, npcSpellIDCap))
			continue
		}
		entries = append(entries, e)
	}
	if len(entries) == 0 {
		return plan.fail("add at least one spell entry")
	}
	listID := req.ID
	name := strings.TrimSpace(req.Name)
	asNew := req.AsNew || listID <= 0
	if asNew {
		if req.ID > 0 {
			var src models.NpcSpell
			if err := s.eqemu().Table("npc_spells").Where("id = ?", req.ID).Take(&src).Error; err == nil {
				if name == "" {
					name = src.Name.String
				}
				name = prefixedName(req.Prefix, name)
				dst := src
				ids, err := s.takeIDs("npc_spells", "id", 0, 1, reservedFloor, &plan)
				if err != nil {
					return plan.fail(err.Error())
				}
				listID = ids[0]
				dst.ID = uint(listID)
				dst.Name = null.StringFrom(name)
				dst.NpcSpellsEntries = nil
				dst.BotSpellsEntries = nil
				dst.NpcSpell = nil
				plan.Clones = append(plan.Clones, Clone{Kind: "npc_spells", SourceID: req.ID, NewID: listID, Name: name})
				plan.DBWrites = append(plan.DBWrites, DBWrite{Table: "npc_spells", Action: "create", ID: listID, Note: name})
				if !req.DryRun {
					if err := s.eqemu().Table("npc_spells").Create(&dst).Error; err != nil {
						return plan.fail(err.Error())
					}
				}
			}
		}
		if listID <= 0 {
			ids, err := s.takeIDs("npc_spells", "id", 0, 1, reservedFloor, &plan)
			if err != nil {
				return plan.fail(err.Error())
			}
			listID = ids[0]
			if name == "" {
				name = prefixedName(req.Prefix, "custom set")
			}
			plan.DBWrites = append(plan.DBWrites, DBWrite{Table: "npc_spells", Action: "create", ID: listID, Note: name})
			if !req.DryRun {
				row := models.NpcSpell{ID: uint(listID), Name: null.StringFrom(name)}
				if err := s.eqemu().Table("npc_spells").Create(&row).Error; err != nil {
					return plan.fail(err.Error())
				}
			}
		}
	} else {
		if name != "" && !req.DryRun {
			_ = s.eqemu().Table("npc_spells").Where("id = ?", listID).Update("name", name).Error
		}
		plan.DBWrites = append(plan.DBWrites, DBWrite{Table: "npc_spells_entries", Action: "delete", ID: listID, Note: "replace entries"})
		if !req.DryRun {
			if err := s.eqemu().Exec("DELETE FROM npc_spells_entries WHERE npc_spells_id = ?", listID).Error; err != nil {
				return plan.fail(err.Error())
			}
		}
	}
	for _, e := range entries {
		plan.DBWrites = append(plan.DBWrites, DBWrite{
			Table: "npc_spells_entries", Action: "create", ID: e.SpellID,
			Note:  fmt.Sprintf("list %d lvl %d-%d", listID, e.MinLevel, e.MaxLevel),
			Extra: map[string]int{"npcSpellsId": listID, "spellId": e.SpellID},
		})
		if req.DryRun {
			continue
		}
		row := models.NpcSpellsEntry{
			NpcSpellsId: listID, Spellid: uint16(e.SpellID), Type: uint(e.Type),
			Minlevel: uint8(e.MinLevel), Maxlevel: uint8(e.MaxLevel),
			Manacost: int16(e.Manacost), RecastDelay: e.RecastDelay, Priority: int16(e.Priority),
		}
		if err := s.eqemu().Table("npc_spells_entries").Create(&row).Error; err != nil {
			return plan.fail(err.Error())
		}
	}
	s.complete(&plan, req.DryRun, false)
	return plan
}
