package zonecontroller

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/EQEmu/spire/internal/models"
	"github.com/volatiletech/null/v8"
)

func (s *Service) writeMerchantKit(zoneID int, clones []ItemClone, dry bool, plan *FactoryPlan) {
	if s.eqemu() == nil || len(clones) == 0 {
		return
	}
	info, err := s.zoneInfo(zoneID)
	if err != nil {
		plan.Warnings = append(plan.Warnings, fmt.Sprintf("zone %d merchant skipped: %v", zoneID, err))
		return
	}
	type merchRow struct {
		NpcID      int `gorm:"column:id"`
		MerchantID int `gorm:"column:merchant_id"`
	}
	var found merchRow
	err = s.eqemu().Table("npc_types").
		Select("npc_types.id, npc_types.merchant_id").
		Joins("JOIN spawnentry ON spawnentry.npcID = npc_types.id").
		Joins("JOIN spawn2 ON spawn2.spawngroupID = spawnentry.spawngroupID").
		Where("spawn2.zone = ? AND npc_types.merchant_id > 0", info.short).
		Order("npc_types.merchant_id asc").
		Limit(1).
		Scan(&found).Error
	merchantID := found.MerchantID
	if err != nil || merchantID <= 0 {
		merchantID = 900000 + zoneID
		plan.Warnings = append(plan.Warnings, fmt.Sprintf("zone %d has no merchant NPC; reserved merchantid %d (set an NPC merchant_id to use it)", zoneID, merchantID))
	}
	var maxSlot int
	_ = s.eqemu().Table("merchantlist").Where("merchantid = ?", merchantID).Select("COALESCE(MAX(slot), 0)").Scan(&maxSlot).Error
	slot := maxSlot
	for _, clone := range clones {
		slot++
		row := models.Merchantlist{
			Merchantid:      merchantID,
			Slot:            uint(slot),
			Item:            clone.NewID,
			FactionRequired: -1,
			ClassesRequired: 65535,
			Probability:     100,
		}
		plan.DBWrites = append(plan.DBWrites, DBWrite{
			Table:  "merchantlist",
			Action: "create",
			ID:     clone.NewID,
			Note:   fmt.Sprintf("merchant %d slot %d", merchantID, slot),
		})
		if dry {
			continue
		}
		if err := s.eqemu().Table("merchantlist").Create(&row).Error; err != nil {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("merchant item %d: %v", clone.NewID, err))
		}
	}
}

func (s *Service) writeTraitRows(zoneID int, group, tier string, clones []ItemClone, dry bool, plan *FactoryPlan) {
	if len(clones) == 0 {
		return
	}
	info, _ := s.zoneInfo(zoneID)
	abs := filepath.Join(s.questsDirPath(), filepath.FromSlash(traitRel))
	doc := map[string]interface{}{"items": []interface{}{}}
	if raw, err := os.ReadFile(abs); err == nil {
		_ = json.Unmarshal(raw, &doc)
	}
	items := asArray(doc["items"])
	have := map[int]bool{}
	for _, raw := range items {
		have[asInt(asObject(raw)["item_id"])] = true
	}
	added := 0
	byID, _ := s.loadItemsByID(idsFromClones(clones))
	for _, clone := range clones {
		src := byID[clone.NewID]
		if src.ID == 0 {
			src = byID[clone.SourceID]
		}
		if src.Clickeffect <= 0 && src.Proceffect <= 0 {
			continue
		}
		if have[clone.NewID] {
			continue
		}
		items = append(items, map[string]interface{}{
			"item_id":         clone.NewID,
			"item_name":       clone.Name,
			"zone_id":         zoneID,
			"zone_short_name": info.short,
			"zone_long_name":  info.long,
			"tier":            firstNonEmpty(tier, group, "Tier"),
			"zone_group":      group,
			"loot_table":      "trait",
			"class_ids":       []interface{}{},
			"classes":         []interface{}{},
			"effects": map[string]interface{}{
				"click": map[string]interface{}{"has_effect": src.Clickeffect > 0, "spell_id": src.Clickeffect},
				"proc":  map[string]interface{}{"has_effect": src.Proceffect > 0, "spell_id": src.Proceffect},
				"worn":  map[string]interface{}{"has_effect": false, "spell_id": nil},
				"focus": map[string]interface{}{"has_effect": false, "spell_id": nil},
			},
			"source": map[string]interface{}{"spire": true, "factory": true},
		})
		have[clone.NewID] = true
		added++
	}
	if added == 0 {
		plan.Warnings = append(plan.Warnings, fmt.Sprintf("zone %d: no click/proc items for trait vendor", zoneID))
		return
	}
	doc["items"] = items
	doc["item_count"] = len(items)
	doc["generated_utc"] = time.Now().UTC().Format(time.RFC3339)
	plan.Writes = append(plan.Writes, WriteFile{ZoneID: zoneID, Kind: "traits", RelPath: traitRel, Action: "upsert"})
	if dry {
		return
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		plan.Warnings = append(plan.Warnings, err.Error())
		return
	}
	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		plan.Warnings = append(plan.Warnings, err.Error())
		return
	}
	if err := os.WriteFile(abs, append(body, '\n'), 0o644); err != nil {
		plan.Warnings = append(plan.Warnings, err.Error())
	}
}

func (s *Service) writeEvolvingChains(steps [][]ItemClone, zoneIDs []int, kills int64, dry bool, plan *FactoryPlan) {
	if s.eqemu() == nil || len(steps) < 2 {
		return
	}
	if kills <= 0 {
		kills = 100
	}
	kitSize := len(steps[0])
	if kitSize == 0 {
		return
	}
	for i := 0; i < kitSize; i++ {
		first := steps[0][i]
		max := len(steps)
		for stepIdx, clones := range steps {
			if i >= len(clones) {
				continue
			}
			clone := clones[i]
			plan.DBWrites = append(plan.DBWrites, DBWrite{
				Table:  "items",
				Action: "evolve",
				ID:     clone.NewID,
				Note:   fmt.Sprintf("evoid %d level %d/%d", first.NewID, stepIdx+1, max),
			})
			if !dry {
				if err := s.eqemu().Table("items").Where("id = ?", clone.NewID).Updates(map[string]interface{}{
					"evoitem":       1,
					"evoid":         first.NewID,
					"evolvinglevel": stepIdx + 1,
					"evomax":        max,
				}).Error; err != nil {
					plan.Warnings = append(plan.Warnings, fmt.Sprintf("evolve item %d: %v", clone.NewID, err))
				}
			}
			if stepIdx >= max-1 {
				continue
			}
			zoneID := 0
			if stepIdx < len(zoneIDs) {
				zoneID = zoneIDs[stepIdx]
			}
			detail := models.ItemsEvolvingDetail{
				ItemEvoId:       null.UintFrom(uint(first.NewID)),
				ItemEvolveLevel: null.UintFrom(uint(stepIdx + 1)),
				ItemId:          null.UintFrom(uint(clone.NewID)),
				Type:            null.UintFrom(4),
				SubType:         null.StringFrom(strconv.Itoa(zoneID)),
				RequiredAmount:  null.Int64From(kills),
			}
			plan.DBWrites = append(plan.DBWrites, DBWrite{
				Table:  "items_evolving_details",
				Action: "create",
				ID:     clone.NewID,
				Note:   fmt.Sprintf("type 4 zone %d x%d", zoneID, kills),
			})
			if dry {
				continue
			}
			if err := s.eqemu().Table("items_evolving_details").Create(&detail).Error; err != nil {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("evolving details %d: %v", clone.NewID, err))
			}
		}
	}
}

func idsFromClones(clones []ItemClone) []int {
	out := make([]int, 0, len(clones)*2)
	for _, c := range clones {
		out = append(out, c.NewID, c.SourceID)
	}
	return out
}
