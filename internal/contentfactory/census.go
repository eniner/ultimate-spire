package contentfactory

import (
	"fmt"
	"strconv"
	"strings"
)

type CensusSpell struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CensusNPC struct {
	ID          int           `json:"id"`
	Name        string        `json:"name"`
	Level       int           `json:"level"`
	HP          int64         `json:"hp,omitempty"`
	Pops        int           `json:"pops,omitempty"`
	Role        string        `json:"role,omitempty"`
	Reason      string        `json:"reason,omitempty"`
	FactionID   int           `json:"factionId,omitempty"`
	Faction     string        `json:"faction,omitempty"`
	LoottableID int           `json:"loottableId"`
	NpcSpellsID int           `json:"npcSpellsId"`
	MerchantID  int           `json:"merchantId"`
	SpellSet    string        `json:"spellSet,omitempty"`
	Spells      []CensusSpell `json:"spells,omitempty"`
}

type CensusItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Slots    int    `json:"slots"`
	Classes  int    `json:"classes"`
	Slot     string `json:"slot,omitempty"`
	Class    string `json:"class,omitempty"`
	Click    int    `json:"click"`
	Proc     int    `json:"proc"`
	Worn     int    `json:"worn"`
	Focus    int    `json:"focus"`
	ClickName string `json:"clickName,omitempty"`
	ProcName  string `json:"procName,omitempty"`
}

type CensusMerchant struct {
	ID       int    `json:"id"`
	NPCName  string `json:"npcName"`
	NPCID    int    `json:"npcId"`
	Stock    int    `json:"stock"`
	NextSlot int    `json:"nextSlot"`
}

type CensusSpellSet struct {
	ID     int           `json:"id"`
	Name   string        `json:"name"`
	NPCs   int           `json:"npcs"`
	Spells []CensusSpell `json:"spells"`
}

type ZoneCensus struct {
	OK        bool             `json:"ok"`
	Error     string           `json:"error,omitempty"`
	ZoneID    int              `json:"zoneId"`
	Short     string           `json:"short"`
	Long      string           `json:"long,omitempty"`
	Npcs      []CensusNPC      `json:"npcs"`
	Items     []CensusItem     `json:"items"`
	Merchants []CensusMerchant `json:"merchants"`
	SpellSets     []CensusSpellSet `json:"spellSets"`
	Neighbors     []ZoneNeighbor   `json:"neighbors,omitempty"`
	GroundSpawns  int              `json:"groundSpawns,omitempty"`
	Forages       int              `json:"forages,omitempty"`
	Counts        map[string]int   `json:"counts"`
}

func (s *Service) Census(zone string) ZoneCensus {
	out := ZoneCensus{Npcs: []CensusNPC{}, Items: []CensusItem{}, Merchants: []CensusMerchant{}, SpellSets: []CensusSpellSet{}, Counts: map[string]int{}}
	if s.eqemu() == nil {
		out.Error = "PEQ database is not connected"
		return out
	}
	id, short, long, err := s.resolveZone(zone)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	out.ZoneID = id
	out.Short = short
	out.Long = long

	classified, err := s.classifyZoneNpcs([]int{id})
	if err != nil {
		out.Error = err.Error()
		return out
	}
	if len(classified) > censusNpcLimit {
		classified = classified[:censusNpcLimit]
	}
	type npcRow = ClassifiedNPC
	npcs := classified
	factionIDs := []int{}
	for _, n := range npcs {
		if n.FactionID > 0 {
			factionIDs = append(factionIDs, n.FactionID)
		}
	}
	factionNames := map[int]string{}
	if len(factionIDs) > 0 {
		type frow struct {
			ID   int
			Name string
		}
		var frows []frow
		_ = s.eqemu().Table("faction_list").Select("id, name").Where("id IN ?", uniqueInts(factionIDs)).Scan(&frows).Error
		for _, f := range frows {
			factionNames[f.ID] = f.Name
		}
	}
	out.Neighbors = s.zoneNeighbors(short)
	var ground, forage int64
	_ = s.eqemu().Table("ground_spawns").Where("zoneid = ?", id).Count(&ground).Error
	_ = s.eqemu().Table("forage").Where("zoneid = ?", id).Count(&forage).Error
	out.GroundSpawns = int(ground)
	out.Forages = int(forage)

	spellSetIDs := uniqueInts(func() []int {
		ids := make([]int, 0, len(npcs))
		for _, n := range npcs {
			ids = append(ids, n.NpcSpellsID)
		}
		return ids
	}())
	lootIDs := uniqueInts(func() []int {
		ids := make([]int, 0, len(npcs))
		for _, n := range npcs {
			ids = append(ids, n.LoottableID)
		}
		return ids
	}())
	merchantIDs := uniqueInts(func() []int {
		ids := make([]int, 0, len(npcs))
		for _, n := range npcs {
			ids = append(ids, n.MerchantID)
		}
		return ids
	}())

	setNames := map[int]string{}
	if len(spellSetIDs) > 0 {
		type setRow struct {
			ID   int
			Name string
		}
		var sets []setRow
		_ = s.eqemu().Table("npc_spells").Select("id, name").Where("id IN ?", spellSetIDs).Scan(&sets).Error
		for _, row := range sets {
			setNames[row.ID] = row.Name
		}
	}

	spellsBySet := map[int][]CensusSpell{}
	if len(spellSetIDs) > 0 {
		type spellRow struct {
			SetID    int    `gorm:"column:npc_spells_id"`
			SpellID  int    `gorm:"column:spellid"`
			SpellName string `gorm:"column:spell_name"`
		}
		var rows []spellRow
		_ = s.eqemu().Table("npc_spells_entries").
			Select("npc_spells_entries.npc_spells_id, npc_spells_entries.spellid, spells_new.name AS spell_name").
			Joins("LEFT JOIN spells_new ON spells_new.id = npc_spells_entries.spellid").
			Where("npc_spells_entries.npc_spells_id IN ?", spellSetIDs).
			Scan(&rows).Error
		seen := map[string]bool{}
		for _, row := range rows {
			key := fmt.Sprintf("%d:%d", row.SetID, row.SpellID)
			if seen[key] {
				continue
			}
			seen[key] = true
			spellsBySet[row.SetID] = append(spellsBySet[row.SetID], CensusSpell{ID: row.SpellID, Name: row.SpellName})
		}
	}

	spellNames := map[int]string{}
	for _, list := range spellsBySet {
		for _, sp := range list {
			spellNames[sp.ID] = sp.Name
		}
	}

	setCount := map[int]int{}
	for _, n := range npcs {
		setCount[n.NpcSpellsID]++
		out.Npcs = append(out.Npcs, CensusNPC{
			ID: n.ID, Name: n.Name, Level: n.Level, HP: n.HP, Pops: n.Pops,
			Role: n.Role, Reason: n.Reason, FactionID: n.FactionID, Faction: factionNames[n.FactionID],
			LoottableID: n.LoottableID, NpcSpellsID: n.NpcSpellsID, MerchantID: n.MerchantID,
			SpellSet: setNames[n.NpcSpellsID],
			Spells:   spellsBySet[n.NpcSpellsID],
		})
	}

	if len(lootIDs) > 0 {
		type itemRow struct {
			ID          int
			Name        string `gorm:"column:Name"`
			Slots       int
			Classes     int
			Clickeffect int
			Proceffect  int
			Worneffect  int
			Focuseffect int
		}
		var items []itemRow
		_ = s.eqemu().Table("loottable_entries").
			Select("DISTINCT items.id, items.Name, items.slots, items.classes, items.clickeffect, items.proceffect, items.worneffect, items.focuseffect").
			Joins("JOIN lootdrop_entries ON lootdrop_entries.lootdrop_id = loottable_entries.lootdrop_id").
			Joins("JOIN items ON items.id = lootdrop_entries.item_id").
			Where("loottable_entries.loottable_id IN ?", lootIDs).
			Limit(400).
			Scan(&items).Error
		needNames := []int{}
		for _, it := range items {
			if it.Clickeffect > 0 {
				needNames = append(needNames, it.Clickeffect)
			}
			if it.Proceffect > 0 {
				needNames = append(needNames, it.Proceffect)
			}
		}
		for id, name := range s.spellNames(needNames) {
			spellNames[id] = name
		}
		for _, it := range items {
			out.Items = append(out.Items, CensusItem{
				ID: it.ID, Name: it.Name, Slots: it.Slots, Classes: it.Classes,
				Slot: slotName(it.Slots), Class: className(it.Classes),
				Click: it.Clickeffect, Proc: it.Proceffect, Worn: it.Worneffect, Focus: it.Focuseffect,
				ClickName: spellNames[it.Clickeffect], ProcName: spellNames[it.Proceffect],
			})
		}
	}

	if len(merchantIDs) > 0 {
		type stockRow struct {
			MerchantID int `gorm:"column:merchantid"`
			Stock      int
			NextSlot   int `gorm:"column:next_slot"`
		}
		var stocks []stockRow
		_ = s.eqemu().Table("merchantlist").
			Select("merchantid, COUNT(*) AS stock, COALESCE(MAX(slot), 0) AS next_slot").
			Where("merchantid IN ?", merchantIDs).
			Group("merchantid").
			Scan(&stocks).Error
		stockBy := map[int]stockRow{}
		for _, row := range stocks {
			stockBy[row.MerchantID] = row
		}
		seenMerch := map[int]bool{}
		for _, n := range npcs {
			if n.MerchantID <= 0 || seenMerch[n.MerchantID] {
				continue
			}
			seenMerch[n.MerchantID] = true
			st := stockBy[n.MerchantID]
			out.Merchants = append(out.Merchants, CensusMerchant{
				ID: n.MerchantID, NPCName: n.Name, NPCID: n.ID,
				Stock: st.Stock, NextSlot: st.NextSlot + 1,
			})
		}
	}

	for _, setID := range spellSetIDs {
		if setID <= 0 {
			continue
		}
		out.SpellSets = append(out.SpellSets, CensusSpellSet{
			ID: setID, Name: setNames[setID], NPCs: setCount[setID], Spells: spellsBySet[setID],
		})
	}

	out.Counts["npcs"] = len(out.Npcs)
	out.Counts["items"] = len(out.Items)
	out.Counts["merchants"] = len(out.Merchants)
	out.Counts["spellSets"] = len(out.SpellSets)
	out.Counts["groundSpawns"] = out.GroundSpawns
	out.Counts["forages"] = out.Forages
	for _, n := range out.Npcs {
		out.Counts[n.Role]++
	}
	out.OK = true
	return out
}

func (s *Service) resolveZone(raw string) (int, string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, "", "", fmt.Errorf("zone id or short name is required")
	}
	type row struct {
		ID    int    `gorm:"column:zoneidnumber"`
		Short string `gorm:"column:short_name"`
		Long  string `gorm:"column:long_name"`
	}
	var z row
	tx := s.eqemu().Table("zone").Select("zoneidnumber, short_name, long_name").Order("version asc").Limit(1)
	if id, err := strconv.Atoi(raw); err == nil && id > 0 {
		tx = tx.Where("zoneidnumber = ?", id)
	} else {
		tx = tx.Where("short_name = ?", raw)
	}
	if err := tx.Scan(&z).Error; err != nil || z.ID == 0 || strings.TrimSpace(z.Short) == "" {
		return 0, "", "", fmt.Errorf("unknown zone %q", raw)
	}
	return z.ID, z.Short, z.Long, nil
}

func (s *Service) spellNames(ids []int) map[int]string {
	out := map[int]string{}
	ids = uniqueInts(ids)
	if s.eqemu() == nil || len(ids) == 0 {
		return out
	}
	type row struct {
		ID   int
		Name string
	}
	var rows []row
	_ = s.eqemu().Table("spells_new").Select("id, name").Where("id IN ?", ids).Scan(&rows).Error
	for _, r := range rows {
		out[r.ID] = r.Name
	}
	return out
}
