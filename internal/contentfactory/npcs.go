package contentfactory

import (
	"fmt"
	"strings"

	"github.com/EQEmu/spire/internal/models"
)

type NpcCloneRequest struct {
	DryRun         bool         `json:"dryRun"`
	SkipJournal    bool         `json:"skipJournal"`
	Prefix         string       `json:"prefix"`
	StartID        int          `json:"startId"`
	ZoneIDs        []int        `json:"zoneIds"`
	NpcIDs         []int        `json:"npcIds"`
	Target         TargetFilter `json:"target"`
	RetargetSpawns *bool        `json:"retargetSpawns"`
	Step           float64      `json:"step"`
}

func (s *Service) RunNpcClones(req NpcCloneRequest) Plan {
	plan := emptyPlan("npcs", req.DryRun)
	if s.eqemu() == nil {
		return plan.fail("PEQ database is not connected")
	}
	ids, classified, err := s.npcIDsFiltered(req.ZoneIDs, req.NpcIDs, req.Target)
	if err != nil {
		return plan.fail(err.Error())
	}
	if len(ids) == 0 {
		return plan.fail("no NPCs match the target filter")
	}
	byID := map[int]ClassifiedNPC{}
	for _, n := range classified {
		byID[n.ID] = n
	}
	var srcs []models.NpcType
	if err := s.eqemu().Table("npc_types").Where("id IN ?", ids).Find(&srcs).Error; err != nil {
		return plan.fail(err.Error())
	}
	if len(srcs) == 0 {
		return plan.fail("source npc_types not found")
	}
	taken, err := s.takeIDs("npc_types", "id", req.StartID, len(srcs), reservedFloor, &plan)
	if err != nil {
		return plan.fail(err.Error())
	}
	retarget := flagOr(req.RetargetSpawns, true)
	remap := map[int]int{}
	for i, src := range srcs {
		newID := taken[i]
		name := prefixedName(req.Prefix, src.Name)
		dst := src
		dst.ID = newID
		dst.Name = name
		clearNpcAssocs(&dst)
		if req.Step > 1 {
			dst.Hp = scaleI64(src.Hp, req.Step)
			dst.Mindmg = uint(scaleU(uint(src.Mindmg), req.Step))
			dst.Maxdmg = uint(scaleU(uint(src.Maxdmg), req.Step))
		}
		info := byID[src.ID]
		clone := Clone{Kind: "npc", SourceID: src.ID, NewID: newID, Name: name}
		clone.Diff = compactDiff([]DiffField{
			diffI64("hp", src.Hp, dst.Hp),
			diffInt("mindmg", int(src.Mindmg), int(dst.Mindmg)),
			diffInt("maxdmg", int(src.Maxdmg), int(dst.Maxdmg)),
			diffInt("level", int(src.Level), int(dst.Level)),
			diffInt("ac", int(src.AC), int(dst.AC)),
		})
		plan.Clones = append(plan.Clones, clone)
		plan.DBWrites = append(plan.DBWrites, DBWrite{Table: "npc_types", Action: "create", ID: newID, Note: name})
		plan.Impact = append(plan.Impact, ImpactRow{
			Kind: "npc", ID: src.ID, Name: src.Name, Field: "clone", From: src.ID, To: newID,
			Role: info.Role, Key: impactKey("npc", src.ID, "clone"), Note: "mint " + name,
		})
		remap[src.ID] = newID
		if req.DryRun {
			continue
		}
		if err := s.eqemu().Table("npc_types").Create(&dst).Error; err != nil {
			return plan.fail(fmt.Sprintf("create npc %d: %v", newID, err))
		}
	}
	if retarget && len(req.ZoneIDs) > 0 {
		if err := s.retargetSpawns(req.ZoneIDs, remap, req.DryRun, &plan); err != nil {
			return plan.fail(err.Error())
		}
	}
	s.complete(&plan, req.DryRun, req.SkipJournal)
	return plan
}

func clearNpcAssocs(n *models.NpcType) {
	if n == nil {
		return
	}
	n.AlternateCurrency = nil
	n.Merchantlists = nil
	n.NpcFactions = nil
	n.NpcSpell = nil
	n.Spawnentries = nil
	n.NpcEmotes = nil
	n.NpcTypesTint = nil
	n.Loottable = nil
}

func (s *Service) retargetSpawns(zoneIDs []int, remap map[int]int, dry bool, plan *Plan) error {
	for _, zoneID := range uniqueInts(zoneIDs) {
		short, err := s.zoneShort(zoneID)
		if err != nil {
			return err
		}
		for oldID, newID := range remap {
			var n int64
			_ = s.eqemu().Table("spawnentry").
				Joins("JOIN spawn2 ON spawn2.spawngroupID = spawnentry.spawngroupID").
				Where("spawn2.zone = ? AND spawnentry.npcID = ?", short, oldID).
				Count(&n).Error
			if n == 0 {
				continue
			}
			plan.DBWrites = append(plan.DBWrites, DBWrite{
				Table:  "spawnentry",
				Action: "update",
				ID:     newID,
				Note:   fmt.Sprintf("%s npc %d -> %d (%d rows)", short, oldID, newID, n),
				Extra:  map[string]int{"oldNpcId": oldID, "newNpcId": newID, "zoneId": zoneID},
			})
			plan.Impact = append(plan.Impact, ImpactRow{
				Kind: "spawn", ID: oldID, Field: "npcID", From: oldID, To: newID, ZoneID: zoneID,
				Key: impactKey("spawn", oldID, "npcID"), Note: short,
			})
			if dry {
				continue
			}
			if err := s.eqemu().Exec(
				"UPDATE spawnentry se JOIN spawn2 s2 ON s2.spawngroupID = se.spawngroupID SET se.npcID = ? WHERE s2.zone = ? AND se.npcID = ?",
				newID, short, oldID,
			).Error; err != nil {
				return fmt.Errorf("retarget %s %d: %w", short, oldID, err)
			}
		}
	}
	return nil
}

func scaleI64(v int64, mult float64) int64 {
	if v <= 0 || mult <= 1 {
		return v
	}
	out := int64(float64(v) * mult)
	if out < 1 {
		return 1
	}
	return out
}

func scaleU(v uint, mult float64) uint {
	if v == 0 || mult <= 1 {
		return v
	}
	out := uint(float64(v) * mult)
	if out < 1 {
		return 1
	}
	return out
}

func cloneRemap(plan Plan) map[int]int {
	out := map[int]int{}
	for _, c := range plan.Clones {
		if c.Kind == "npc" && c.SourceID > 0 && c.NewID > 0 {
			out[c.SourceID] = c.NewID
		}
	}
	return out
}

func remapIDs(ids []int, remap map[int]int) []int {
	if len(remap) == 0 {
		return ids
	}
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if next, ok := remap[id]; ok {
			out = append(out, next)
		} else {
			out = append(out, id)
		}
	}
	return uniqueInts(out)
}

func prefixedNPC(prefix, name string) string {
	return prefixedName(strings.TrimSpace(prefix), name)
}
