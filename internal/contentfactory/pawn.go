package contentfactory

import (
	"fmt"
	"strings"

	"github.com/EQEmu/spire/internal/models"
	"github.com/volatiletech/null/v8"
)

type PawnRequest struct {
	DryRun      bool    `json:"dryRun"`
	ZoneID      int     `json:"zoneId"`
	SourceNpcID int     `json:"sourceNpcId"`
	Prefix      string  `json:"prefix"`
	LoottableID int     `json:"loottableId"`
	NpcSpellsID int     `json:"npcSpellsId"`
	X           float32 `json:"x"`
	Y           float32 `json:"y"`
	Z           float32 `json:"z"`
	Heading     float32 `json:"heading"`
}

type GiveRequest struct {
	DryRun      bool   `json:"dryRun"`
	CharacterID int    `json:"characterId"`
	Name        string `json:"name"`
	ItemIDs     []int  `json:"itemIds"`
}

func (s *Service) RunPawn(req PawnRequest) Plan {
	plan := emptyPlan("pawn", req.DryRun)
	if s.eqemu() == nil {
		return plan.fail("PEQ database is not connected")
	}
	if req.ZoneID <= 0 {
		return plan.fail("zone id is required")
	}
	short, err := s.zoneShort(req.ZoneID)
	if err != nil {
		return plan.fail(err.Error())
	}
	srcID := req.SourceNpcID
	if srcID <= 0 {
		ids, classified, err := s.npcIDsFiltered([]int{req.ZoneID}, nil, TargetFilter{Roles: []string{"trash"}})
		if err != nil {
			return plan.fail(err.Error())
		}
		if len(ids) == 0 {
			return plan.fail("no trash NPC to clone for a test pawn")
		}
		srcID = ids[0]
		_ = classified
	}
	var src models.NpcType
	if err := s.eqemu().Table("npc_types").Where("id = ?", srcID).Take(&src).Error; err != nil {
		return plan.fail(fmt.Sprintf("npc %d not found", srcID))
	}
	taken, err := s.takeIDs("npc_types", "id", 0, 1, reservedFloor, &plan)
	if err != nil {
		return plan.fail(err.Error())
	}
	npcID := taken[0]
	prefix := firstNonEmpty(req.Prefix, "SPIRE_TEST")
	name := prefixedName(prefix, src.Name)
	if !strings.HasPrefix(strings.ToUpper(name), "SPIRE_TEST") {
		name = "SPIRE_TEST_" + name
	}
	dst := src
	dst.ID = npcID
	dst.Name = name
	if req.LoottableID > 0 {
		dst.LoottableId = uint(req.LoottableID)
	}
	if req.NpcSpellsID > 0 {
		dst.NpcSpellsId = uint(req.NpcSpellsID)
	}
	clearNpcAssocs(&dst)
	x, y, z, heading := req.X, req.Y, req.Z, req.Heading
	if x == 0 && y == 0 && z == 0 {
		x, y, z, heading = s.zoneSafe(req.ZoneID)
	}
	sgIDs, err := s.takeIDs("spawngroup", "id", 0, 1, reservedFloor, &plan)
	if err != nil {
		return plan.fail(err.Error())
	}
	s2IDs, err := s.takeIDs("spawn2", "id", 0, 1, reservedFloor, &plan)
	if err != nil {
		return plan.fail(err.Error())
	}
	sgID, s2ID := sgIDs[0], s2IDs[0]
	plan.Clones = append(plan.Clones, Clone{Kind: "npc", SourceID: srcID, NewID: npcID, Name: name})
	plan.DBWrites = append(plan.DBWrites,
		DBWrite{Table: "npc_types", Action: "create", ID: npcID, Note: name},
		DBWrite{Table: "spawngroup", Action: "create", ID: sgID, Note: name},
		DBWrite{Table: "spawn2", Action: "create", ID: s2ID, Note: short, Extra: map[string]int{"spawngroupId": sgID, "zoneId": req.ZoneID}},
		DBWrite{Table: "spawnentry", Action: "create", ID: npcID, Extra: map[string]int{"spawngroupId": sgID, "npcId": npcID}},
	)
	plan.Impact = append(plan.Impact, ImpactRow{
		Kind: "pawn", ID: npcID, Name: name, Field: "spawn2", To: s2ID, ZoneID: req.ZoneID,
		Key: impactKey("pawn", npcID, "spawn2"), Note: fmt.Sprintf("%s @ %.1f,%.1f,%.1f", short, x, y, z),
	})
	if !req.DryRun {
		if err := s.eqemu().Table("npc_types").Create(&dst).Error; err != nil {
			return plan.fail(err.Error())
		}
		sg := models.Spawngroup{ID: sgID, Name: name}
		if err := s.eqemu().Table("spawngroup").Create(&sg).Error; err != nil {
			return plan.fail(err.Error())
		}
		s2 := models.Spawn2{
			ID: s2ID, SpawngroupID: sgID, Zone: null.StringFrom(short),
			X: x, Y: y, Z: z, Heading: heading, Respawntime: 60,
		}
		if err := s.eqemu().Table("spawn2").Create(&s2).Error; err != nil {
			return plan.fail(err.Error())
		}
		se := models.Spawnentry{SpawngroupID: sgID, NpcID: npcID, Chance: 100}
		if err := s.eqemu().Table("spawnentry").Create(&se).Error; err != nil {
			return plan.fail(err.Error())
		}
	}
	s.complete(&plan, req.DryRun, false)
	return plan
}

func (s *Service) zoneSafe(zoneID int) (float32, float32, float32, float32) {
	type row struct {
		X, Y, Z, Heading float32 `gorm:"column:safe_x"`
	}
	var z models.Zone
	if err := s.eqemu().Table("zone").Where("zoneidnumber = ?", zoneID).Order("version asc").First(&z).Error; err != nil {
		return 0, 0, 0, 0
	}
	return z.SafeX, z.SafeY, z.SafeZ, z.SafeHeading
}

func (s *Service) RunGive(req GiveRequest) Plan {
	plan := emptyPlan("give", req.DryRun)
	if s.eqemu() == nil {
		return plan.fail("PEQ database is not connected")
	}
	charID := req.CharacterID
	if charID <= 0 && strings.TrimSpace(req.Name) != "" {
		_ = s.eqemu().Table("character_data").Select("id").Where("name = ?", strings.TrimSpace(req.Name)).Limit(1).Scan(&charID).Error
	}
	if charID <= 0 {
		return plan.fail("character id or exact name is required")
	}
	itemIDs := uniqueInts(req.ItemIDs)
	if len(itemIDs) == 0 {
		return plan.fail("add item ids to give")
	}
	var used []int
	_ = s.eqemu().Table("inventory").Where("character_id = ?", charID).Pluck("slot_id", &used).Error
	busy := map[int]bool{}
	for _, slot := range used {
		busy[slot] = true
	}
	slots := freeInvSlots(busy, len(itemIDs))
	if len(slots) < len(itemIDs) {
		return plan.fail(fmt.Sprintf("only %d free inventory slots on character %d", len(slots), charID))
	}
	for i, itemID := range itemIDs {
		slot := slots[i]
		plan.DBWrites = append(plan.DBWrites, DBWrite{
			Table: "inventory", Action: "create", ID: itemID,
			Note:  fmt.Sprintf("char %d slot %d", charID, slot),
			Extra: map[string]int{"characterId": charID, "slot": slot, "itemId": itemID},
		})
		plan.Impact = append(plan.Impact, ImpactRow{
			Kind: "character", ID: charID, Field: "slot", From: 0, To: slot,
			Key: impactKey("character", charID, fmt.Sprintf("slot%d", slot)), Note: fmt.Sprintf("item %d", itemID),
		})
		if req.DryRun {
			continue
		}
		row := models.Inventory{CharacterId: uint(charID), SlotId: uint32(slot), ItemId: null.UintFrom(uint(itemID)), Charges: null.Uint16From(1)}
		if err := s.eqemu().Table("inventory").Create(&row).Error; err != nil {
			return plan.fail(fmt.Sprintf("give item %d: %v", itemID, err))
		}
	}
	s.complete(&plan, req.DryRun, false)
	return plan
}

func freeInvSlots(busy map[int]bool, need int) []int {
	out := []int{}
	for slot := 22; slot <= 29 && len(out) < need; slot++ {
		if !busy[slot] {
			out = append(out, slot)
		}
	}
	for slot := 262; slot <= 351 && len(out) < need; slot++ {
		if !busy[slot] {
			out = append(out, slot)
		}
	}
	return out
}
