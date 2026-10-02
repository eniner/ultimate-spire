package contentfactory

import (
	"fmt"
	"strings"

	"github.com/EQEmu/spire/internal/models"
)

type LootItemIn struct {
	ItemID     int     `json:"itemId"`
	Chance     float32 `json:"chance"`
	Multiplier uint8   `json:"multiplier"`
	Charges    uint16  `json:"charges"`
}

type LootTableIn struct {
	Name        string       `json:"name"`
	Mincash     uint         `json:"mincash"`
	Maxcash     uint         `json:"maxcash"`
	Probability float32      `json:"probability"`
	Multiplier  uint8        `json:"multiplier"`
	Droplimit   uint8        `json:"droplimit"`
	Mindrop     uint8        `json:"mindrop"`
	Items       []LootItemIn `json:"items"`
	Target      string       `json:"target,omitempty"`
}

type LootMade struct {
	Name       string `json:"name"`
	TableID    int    `json:"tableId"`
	DropID     int    `json:"dropId"`
	ItemCount  int    `json:"itemCount"`
	Target     string `json:"target,omitempty"`
}

type LootRequest struct {
	DryRun       bool          `json:"dryRun"`
	StartTableID int           `json:"startTableId"`
	StartDropID  int           `json:"startDropId"`
	Tables       []LootTableIn `json:"tables"`
	AttachNpcIDs []int         `json:"attachNpcIds"`
	ZoneIDs      []int         `json:"zoneIds"`
	NameContains string        `json:"nameContains"`
	Target       TargetFilter  `json:"target"`
	RecipeID     string        `json:"recipeId"`
	SkipJournal  bool          `json:"skipJournal"`
	SkipAttach   bool          `json:"skipAttach"`
}

func lootHasItems(tables []LootTableIn) bool {
	for _, table := range tables {
		for _, it := range table.Items {
			if it.ItemID > 0 {
				return true
			}
		}
	}
	return false
}

func fillLootFromItems(tables []LootTableIn, itemIDs []int) []LootTableIn {
	if len(tables) == 0 || lootHasItems(tables) || len(itemIDs) == 0 {
		return tables
	}
	items := make([]LootItemIn, 0, len(itemIDs))
	for _, id := range itemIDs {
		items = append(items, LootItemIn{ItemID: id, Chance: 100})
	}
	tables[0].Items = items
	return tables
}

func (s *Service) RunLoot(req LootRequest) Plan {
	plan := emptyPlan("loot", req.DryRun)
	if s.eqemu() == nil {
		return plan.fail("PEQ database is not connected")
	}
	tables := []LootTableIn{}
	for _, table := range req.Tables {
		items := []LootItemIn{}
		for _, it := range table.Items {
			if it.ItemID > 0 {
				if it.Chance <= 0 {
					it.Chance = 100
				}
				if it.Multiplier == 0 {
					it.Multiplier = 1
				}
				if it.Charges == 0 {
					it.Charges = 1
				}
				items = append(items, it)
			}
		}
		if len(items) == 0 {
			continue
		}
		table.Items = items
		if strings.TrimSpace(table.Name) == "" {
			table.Name = fmt.Sprintf("spire_loot_%d", len(tables)+1)
		}
		if table.Probability <= 0 {
			table.Probability = 100
		}
		if table.Multiplier == 0 {
			table.Multiplier = 1
		}
		tables = append(tables, table)
	}
	if len(tables) == 0 {
		return plan.fail("add at least one loot table with item ids")
	}
	tableIDs, err := s.takeIDs("loottable", "id", req.StartTableID, len(tables), reservedFloor, &plan)
	if err != nil {
		return plan.fail(err.Error())
	}
	dropIDs, err := s.takeIDs("lootdrop", "id", req.StartDropID, len(tables), reservedFloor, &plan)
	if err != nil {
		return plan.fail(err.Error())
	}
	for i, table := range tables {
		tableID := tableIDs[i]
		dropID := dropIDs[i]
		made := LootMade{Name: table.Name, TableID: tableID, DropID: dropID, ItemCount: len(table.Items), Target: table.Target}
		plan.Tables = append(plan.Tables, made)
		plan.DBWrites = append(plan.DBWrites, DBWrite{Table: "loottable", Action: "create", ID: tableID, Note: table.Name})
		plan.DBWrites = append(plan.DBWrites, DBWrite{Table: "lootdrop", Action: "create", ID: dropID, Note: table.Name})
		plan.DBWrites = append(plan.DBWrites, DBWrite{
			Table: "loottable_entries", Action: "create", ID: tableID,
			Note:  fmt.Sprintf("drop %d", dropID),
			Extra: map[string]int{"loottableId": tableID, "lootdropId": dropID},
		})
		for _, it := range table.Items {
			plan.DBWrites = append(plan.DBWrites, DBWrite{
				Table: "lootdrop_entries", Action: "create", ID: it.ItemID,
				Note:  fmt.Sprintf("drop %d chance %.1f", dropID, it.Chance),
				Extra: map[string]int{"lootdropId": dropID, "itemId": it.ItemID},
			})
		}
		if req.DryRun {
			continue
		}
		lt := models.Loottable{ID: uint(tableID), Name: table.Name, Mincash: table.Mincash, Maxcash: table.Maxcash}
		if err := s.eqemu().Table("loottable").Create(&lt).Error; err != nil {
			return plan.fail(fmt.Sprintf("create loottable %d: %v", tableID, err))
		}
		ld := models.Lootdrop{ID: uint(dropID), Name: table.Name}
		if err := s.eqemu().Table("lootdrop").Create(&ld).Error; err != nil {
			return plan.fail(fmt.Sprintf("create lootdrop %d: %v", dropID, err))
		}
		entry := models.LoottableEntry{
			LoottableId: uint(tableID),
			LootdropId:  uint(dropID),
			Multiplier:  table.Multiplier,
			Droplimit:   table.Droplimit,
			Mindrop:     table.Mindrop,
			Probability: table.Probability,
		}
		if err := s.eqemu().Table("loottable_entries").Create(&entry).Error; err != nil {
			return plan.fail(fmt.Sprintf("create loottable_entries %d: %v", tableID, err))
		}
		for _, it := range table.Items {
			drop := models.LootdropEntry{
				LootdropId:  uint(dropID),
				ItemId:      it.ItemID,
				ItemCharges: it.Charges,
				Chance:      it.Chance,
				Multiplier:  it.Multiplier,
			}
			if err := s.eqemu().Table("lootdrop_entries").Create(&drop).Error; err != nil {
				return plan.fail(fmt.Sprintf("create lootdrop_entries %d/%d: %v", dropID, it.ItemID, err))
			}
		}
	}
	if !req.SkipAttach {
		if req.Target.NameContains == "" {
			req.Target.NameContains = req.NameContains
		}
		for i, table := range plan.Tables {
			filter := req.Target
			if i < len(tables) && strings.TrimSpace(tables[i].Target) != "" {
				filter.Roles = []string{tables[i].Target}
			}
			npcIDs, classified, err := s.npcIDsFiltered(req.ZoneIDs, req.AttachNpcIDs, filter)
			if err != nil {
				plan.Warnings = append(plan.Warnings, err.Error())
				continue
			}
			roles := map[int]string{}
			for _, n := range classified {
				roles[n.ID] = n.Role
			}
			if len(npcIDs) > 0 {
				if err := s.setNpcLoot(npcIDs, table.TableID, req.DryRun, roles, &plan); err != nil {
					return plan.fail(err.Error())
				}
			}
		}
	}
	if strings.TrimSpace(req.RecipeID) != "" {
		if err := s.exportLootRecipe(req, plan.Tables, req.DryRun, &plan); err != nil {
			plan.Warnings = append(plan.Warnings, err.Error())
		} else {
			plan.RecipeID = strings.TrimSpace(req.RecipeID)
		}
	}
	s.complete(&plan, req.DryRun, req.SkipJournal)
	return plan
}
