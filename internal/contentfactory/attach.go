package contentfactory

import (
	"fmt"

	"github.com/EQEmu/spire/internal/models"
)

type AttachRequest struct {
	DryRun       bool   `json:"dryRun"`
	SkipJournal  bool   `json:"skipJournal"`
	MerchantID   int    `json:"merchantId"`
	ItemIDs      []int  `json:"itemIds"`
	Replace      bool   `json:"replace"`
	NpcIDs       []int  `json:"npcIds"`
	LoottableID  int    `json:"loottableId"`
	NpcSpellsID  int    `json:"npcSpellsId"`
	ZoneIDs      []int  `json:"zoneIds"`
	NameContains string       `json:"nameContains"`
	Target       TargetFilter `json:"target"`
	Reload       bool         `json:"reload"`
	ExportClient bool         `json:"exportClient"`
	SyncZoneJSON bool         `json:"syncZoneJson"`
}

func (s *Service) RunAttach(req AttachRequest) Plan {
	plan := emptyPlan("attach", req.DryRun)
	if s.eqemu() == nil {
		return plan.fail("PEQ database is not connected")
	}
	itemIDs := uniqueInts(req.ItemIDs)
	if req.MerchantID > 0 && len(itemIDs) > 0 {
		if err := s.stockMerchant(req.MerchantID, itemIDs, req.Replace, req.DryRun, &plan); err != nil {
			return plan.fail(err.Error())
		}
	}
	if req.Target.NameContains == "" {
		req.Target.NameContains = req.NameContains
	}
	npcIDs, classified, err := s.npcIDsFiltered(req.ZoneIDs, req.NpcIDs, req.Target)
	if err != nil {
		return plan.fail(err.Error())
	}
	roles := map[int]string{}
	for _, n := range classified {
		roles[n.ID] = n.Role
	}
	if len(npcIDs) > 0 && req.LoottableID > 0 {
		if err := s.setNpcLoot(npcIDs, req.LoottableID, req.DryRun, roles, &plan); err != nil {
			return plan.fail(err.Error())
		}
	}
	if len(npcIDs) > 0 && req.NpcSpellsID > 0 {
		if err := s.setNpcSpells(npcIDs, req.NpcSpellsID, req.DryRun, roles, &plan); err != nil {
			return plan.fail(err.Error())
		}
	}
	if len(plan.DBWrites) == 0 {
		return plan.fail("nothing to attach. Set a merchant + items, or NPCs/zones + loot table / spell set")
	}
	if req.SyncZoneJSON && len(req.ZoneIDs) > 0 {
		s.SyncZoneItemSpells(req.ZoneIDs, req.DryRun, &plan)
	}
	if req.ExportClient && !req.DryRun {
		info := s.ExportClient(false)
		plan.Export = &info
	}
	if req.Reload && !req.DryRun {
		info := s.ReloadZones(req.ZoneIDs)
		plan.Reload = &info
		plan.Warnings = append(plan.Warnings, info.Warnings...)
	}
	s.complete(&plan, req.DryRun, req.SkipJournal)
	return plan
}

func (s *Service) stockMerchant(merchantID int, itemIDs []int, replace, dry bool, plan *Plan) error {
	plan.Merchants = uniqueInts(append(plan.Merchants, merchantID))
	if replace {
		var existing []models.Merchantlist
		if err := s.eqemu().Table("merchantlist").Where("merchantid = ?", merchantID).Find(&existing).Error; err != nil {
			return err
		}
		for _, row := range existing {
			plan.DBWrites = append(plan.DBWrites, DBWrite{
				Table:  "merchantlist",
				Action: "delete",
				ID:     row.Item,
				Note:   fmt.Sprintf("merchant %d slot %d", merchantID, row.Slot),
				Extra:  map[string]int{"merchantId": merchantID, "slot": int(row.Slot), "itemId": row.Item},
			})
		}
		if !dry {
			if err := s.eqemu().Table("merchantlist").Where("merchantid = ?", merchantID).Delete(&models.Merchantlist{}).Error; err != nil {
				return err
			}
		}
	}
	var maxSlot int
	_ = s.eqemu().Table("merchantlist").Where("merchantid = ?", merchantID).Select("COALESCE(MAX(slot), 0)").Scan(&maxSlot).Error
	if replace {
		maxSlot = 0
	}
	have := map[int]bool{}
	if !replace {
		var ids []int
		_ = s.eqemu().Table("merchantlist").Where("merchantid = ?", merchantID).Pluck("item", &ids).Error
		for _, id := range ids {
			have[id] = true
		}
	}
	slot := maxSlot
	for _, itemID := range itemIDs {
		if have[itemID] {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("merchant %d already sells item %d", merchantID, itemID))
			continue
		}
		slot++
		plan.DBWrites = append(plan.DBWrites, DBWrite{
			Table:  "merchantlist",
			Action: "create",
			ID:     itemID,
			Note:   fmt.Sprintf("merchant %d slot %d", merchantID, slot),
			Extra:  map[string]int{"merchantId": merchantID, "slot": slot, "itemId": itemID},
		})
		plan.Impact = append(plan.Impact, ImpactRow{
			Kind: "merchant", ID: merchantID, Field: "slot", From: 0, To: slot,
			Note: fmt.Sprintf("add item %d", itemID),
		})
		plan.Attached++
		if dry {
			continue
		}
		row := models.Merchantlist{
			Merchantid:      merchantID,
			Slot:            uint(slot),
			Item:            itemID,
			FactionRequired: -1,
			ClassesRequired: 65535,
			Probability:     100,
		}
		if err := s.eqemu().Table("merchantlist").Create(&row).Error; err != nil {
			return fmt.Errorf("merchant item %d: %w", itemID, err)
		}
	}
	return nil
}

func (s *Service) setNpcLoot(npcIDs []int, tableID int, dry bool, roles map[int]string, plan *Plan) error {
	return s.updateNpcs(npcIDs, map[string]interface{}{"loottable_id": tableID}, "loottable", tableID, dry, roles, plan)
}

func (s *Service) setNpcSpells(npcIDs []int, spellsID int, dry bool, roles map[int]string, plan *Plan) error {
	return s.updateNpcs(npcIDs, map[string]interface{}{"npc_spells_id": spellsID}, "npc_spells", spellsID, dry, roles, plan)
}

func (s *Service) updateNpcs(npcIDs []int, values map[string]interface{}, field string, value int, dry bool, roles map[int]string, plan *Plan) error {
	type row struct {
		ID          int
		Name        string
		LoottableID int `gorm:"column:loottable_id"`
		NpcSpellsID int `gorm:"column:npc_spells_id"`
		MerchantID  int `gorm:"column:merchant_id"`
	}
	var rows []row
	if err := s.eqemu().Table("npc_types").Select("id, name, loottable_id, npc_spells_id, merchant_id").Where("id IN ?", npcIDs).Find(&rows).Error; err != nil {
		return err
	}
	have := map[int]row{}
	for _, r := range rows {
		have[r.ID] = r
	}
	for _, id := range npcIDs {
		prev, ok := have[id]
		if !ok {
			if src, name := plannedNpc(plan, id); src > 0 {
				if base, found := have[src]; found {
					prev = base
				} else {
					var srcRow row
					if err := s.eqemu().Table("npc_types").Select("id, name, loottable_id, npc_spells_id, merchant_id").Where("id = ?", src).Take(&srcRow).Error; err == nil {
						prev = srcRow
					}
				}
				prev.ID = id
				if name != "" {
					prev.Name = name
				}
				ok = true
			}
		}
		if !ok {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("npc %d not found", id))
			continue
		}
		from := prev.LoottableID
		if field == "npc_spells" {
			from = prev.NpcSpellsID
		}
		plan.DBWrites = append(plan.DBWrites, DBWrite{
			Table:  "npc_types",
			Action: "update",
			ID:     id,
			Note:   fmt.Sprintf("%s = %d (was %d)", field, value, from),
			Extra: map[string]int{
				"loottableId": prev.LoottableID,
				"npcSpellsId": prev.NpcSpellsID,
				"merchantId":  prev.MerchantID,
			},
		})
		if from != value {
			role := ""
			if roles != nil {
				role = roles[id]
			}
			plan.Impact = append(plan.Impact, ImpactRow{
				Kind: "npc", ID: id, Name: prev.Name, Field: field, From: from, To: value,
				Role: role, Key: impactKey("npc", id, field),
			})
		}
		plan.Attached++
		if dry {
			continue
		}
		if err := s.eqemu().Table("npc_types").Where("id = ?", id).Updates(values).Error; err != nil {
			return fmt.Errorf("update npc %d: %w", id, err)
		}
	}
	return nil
}

func plannedNpc(plan *Plan, newID int) (int, string) {
	if plan == nil {
		return 0, ""
	}
	for _, c := range plan.Clones {
		if (c.Kind == "npc" || c.Kind == "npc_types") && c.NewID == newID {
			return c.SourceID, c.Name
		}
	}
	return 0, ""
}
