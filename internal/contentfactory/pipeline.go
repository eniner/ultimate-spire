package contentfactory

import "strings"

type PipelineRequest struct {
	DryRun          bool            `json:"dryRun"`
	Prefix          string          `json:"prefix"`
	Step            float64         `json:"step"`
	RecipeID        string          `json:"recipeId"`
	RecipeName      string          `json:"recipeName"`
	Group           string          `json:"group"`
	ZoneIDs         []int           `json:"zoneIds"`
	NameContains    string          `json:"nameContains"`
	Target          TargetFilter    `json:"target"`
	MerchantID      int             `json:"merchantId"`
	ReplaceMerchant bool            `json:"replaceMerchant"`
	ExportClient    bool            `json:"exportClient"`
	ExportDbStr     bool            `json:"exportDbstr"`
	Reload          bool            `json:"reload"`
	SyncZoneJSON    bool            `json:"syncZoneJson"`
	CloneNpcs       *bool           `json:"cloneNpcs"`
	RetargetSpawns  *bool           `json:"retargetSpawns"`
	Probe           *bool           `json:"probe"`
	Spells          SpellRequest    `json:"spells"`
	Items           ItemRequest     `json:"items"`
	Loot            []LootTableIn   `json:"loot"`
	NamedLoot       []LootTableIn   `json:"namedLoot"`
	NpcSpells       NpcSpellRequest `json:"npcSpells"`
	NpcSpellsNamed  NpcSpellRequest `json:"npcSpellsNamed"`
}

func (s *Service) RunPipeline(req PipelineRequest) Plan {
	plan := emptyPlan("pipeline", req.DryRun)
	prefix := strings.TrimSpace(req.Prefix)
	if req.Step <= 0 {
		req.Step = 1
	}
	recipeID := strings.TrimSpace(req.RecipeID)
	if req.Target.NameContains == "" {
		req.Target.NameContains = req.NameContains
	}
	req.ZoneIDs = s.expandZones(req.ZoneIDs, req.Target.IncludeRing)

	if len(req.Spells.Sources) > 0 {
		req.Spells.DryRun = req.DryRun
		req.Spells.SkipJournal = true
		if req.Spells.Prefix == "" {
			req.Spells.Prefix = prefix
		}
		if req.Spells.Step <= 0 {
			req.Spells.Step = req.Step
		}
		mergePlan(&plan, s.RunSpells(req.Spells))
		if plan.Error != "" {
			return plan
		}
	}

	if len(normalizeCells(req.Items.Cells)) > 0 {
		req.Items.DryRun = req.DryRun
		req.Items.SkipJournal = true
		if req.Items.Prefix == "" {
			req.Items.Prefix = prefix
		}
		if req.Items.Step <= 0 {
			req.Items.Step = req.Step
		}
		if req.Items.RecipeID == "" {
			req.Items.RecipeID = recipeID
			req.Items.RecipeName = firstNonEmpty(req.RecipeName, recipeID)
			req.Items.Group = req.Group
		}
		mergePlan(&plan, s.RunItems(req.Items))
		if plan.Error != "" {
			return plan
		}
	}

	cloneNpcs := flagOr(req.CloneNpcs, len(req.ZoneIDs) > 0)
	if cloneNpcs && len(req.ZoneIDs) > 0 {
		mergePlan(&plan, s.RunNpcClones(NpcCloneRequest{
			DryRun: req.DryRun, SkipJournal: true, Prefix: prefix, Step: req.Step,
			ZoneIDs: req.ZoneIDs, Target: req.Target, RetargetSpawns: req.RetargetSpawns,
		}))
		if plan.Error != "" {
			return plan
		}
	}
	remap := cloneRemap(plan)

	kitItems := cloneIDs(plan, "item")
	if len(req.Loot) > 0 {
		for i := range req.Loot {
			if strings.TrimSpace(req.Loot[i].Target) == "" {
				req.Loot[i].Target = "trash"
			}
		}
		req.Loot = fillLootFromItems(req.Loot, kitItems)
		if lootHasItems(req.Loot) {
			lootReq := LootRequest{
				DryRun: req.DryRun, SkipJournal: true, SkipAttach: true,
				Tables: req.Loot, RecipeID: recipeID,
			}
			if lootReq.Tables[0].Name == "" && prefix != "" {
				lootReq.Tables[0].Name = prefix + "-trash"
			}
			mergePlan(&plan, s.RunLoot(lootReq))
			if plan.Error != "" {
				return plan
			}
		} else {
			plan.Warnings = append(plan.Warnings, "trash loot skipped: no item ids")
		}
	}

	if len(req.NamedLoot) > 0 {
		for i := range req.NamedLoot {
			if strings.TrimSpace(req.NamedLoot[i].Target) == "" {
				req.NamedLoot[i].Target = "named"
			}
		}
		req.NamedLoot = fillLootFromItems(req.NamedLoot, kitItems)
		if lootHasItems(req.NamedLoot) {
			lootReq := LootRequest{
				DryRun: req.DryRun, SkipJournal: true, SkipAttach: true,
				Tables: req.NamedLoot, RecipeID: recipeID,
			}
			if lootReq.Tables[0].Name == "" && prefix != "" {
				lootReq.Tables[0].Name = prefix + "-named"
			}
			mergePlan(&plan, s.RunLoot(lootReq))
			if plan.Error != "" {
				return plan
			}
		}
	}

	if req.NpcSpells.SourceID > 0 {
		req.NpcSpells.DryRun = req.DryRun
		req.NpcSpells.SkipJournal = true
		req.NpcSpells.SkipAttach = true
		if req.NpcSpells.Prefix == "" {
			req.NpcSpells.Prefix = prefix
		}
		mergePlan(&plan, s.RunNpcSpells(req.NpcSpells))
		if plan.Error != "" {
			return plan
		}
	}
	if req.NpcSpellsNamed.SourceID > 0 {
		req.NpcSpellsNamed.DryRun = req.DryRun
		req.NpcSpellsNamed.SkipJournal = true
		req.NpcSpellsNamed.SkipAttach = true
		if req.NpcSpellsNamed.Prefix == "" {
			req.NpcSpellsNamed.Prefix = prefix + " Named"
		}
		mergePlan(&plan, s.RunNpcSpells(req.NpcSpellsNamed))
		if plan.Error != "" {
			return plan
		}
	}

	trashIDs, trashRows, _ := s.npcIDsFiltered(req.ZoneIDs, nil, withRoles(req.Target, "trash"))
	namedIDs, namedRows, _ := s.npcIDsFiltered(req.ZoneIDs, nil, withRoles(req.Target, "named"))
	trashIDs = remapIDs(trashIDs, remap)
	namedIDs = remapIDs(namedIDs, remap)
	trashRoles := roleMap(trashRows, remap)
	namedRoles := roleMap(namedRows, remap)

	trashTable, namedTable := tableForTarget(plan.Tables, "trash"), tableForTarget(plan.Tables, "named")
	if trashTable == 0 && namedTable == 0 && len(plan.Tables) > 0 {
		trashTable = plan.Tables[0].TableID
	}
	spellIDs := cloneIDs(plan, "npc_spells")
	trashSpells, namedSpells := 0, 0
	if len(spellIDs) > 0 {
		trashSpells = spellIDs[0]
	}
	if len(spellIDs) > 1 {
		namedSpells = spellIDs[1]
	} else {
		namedSpells = trashSpells
	}

	if trashTable > 0 && len(trashIDs) > 0 {
		if err := s.setNpcLoot(trashIDs, trashTable, req.DryRun, trashRoles, &plan); err != nil {
			return plan.fail(err.Error())
		}
	}
	if namedTable > 0 && len(namedIDs) > 0 {
		if err := s.setNpcLoot(namedIDs, namedTable, req.DryRun, namedRoles, &plan); err != nil {
			return plan.fail(err.Error())
		}
	}
	if trashSpells > 0 && len(trashIDs) > 0 {
		if err := s.setNpcSpells(trashIDs, trashSpells, req.DryRun, trashRoles, &plan); err != nil {
			return plan.fail(err.Error())
		}
	}
	if namedSpells > 0 && len(namedIDs) > 0 && req.NpcSpellsNamed.SourceID > 0 {
		if err := s.setNpcSpells(namedIDs, namedSpells, req.DryRun, namedRoles, &plan); err != nil {
			return plan.fail(err.Error())
		}
	}

	itemIDs := kitItems
	if req.MerchantID > 0 && len(itemIDs) > 0 {
		attach := AttachRequest{
			DryRun: req.DryRun, SkipJournal: true, MerchantID: req.MerchantID,
			ItemIDs: itemIDs, Replace: req.ReplaceMerchant,
		}
		mergePlan(&plan, s.RunAttach(attach))
		if plan.Error != "" {
			return plan
		}
	}

	if req.SyncZoneJSON && len(req.ZoneIDs) > 0 {
		s.SyncZoneItemSpells(req.ZoneIDs, req.DryRun, &plan)
	}

	if req.ExportClient && !req.DryRun {
		info := s.ExportClient(req.ExportDbStr)
		plan.Export = &info
		if !info.OK {
			plan.Warnings = append(plan.Warnings, firstNonEmpty(info.Note, "client export failed"))
		}
	} else if req.ExportClient && req.DryRun {
		plan.Export = &ExportInfo{OK: true, Note: "Client export will write spells_us.txt after a real write."}
	}

	if req.Reload && !req.DryRun {
		info := s.ReloadZones(req.ZoneIDs)
		plan.Reload = &info
		plan.Warnings = append(plan.Warnings, info.Warnings...)
	} else if req.Reload && req.DryRun {
		plan.Reload = &ReloadInfo{OK: true, Note: "Reload/repop will run after a real write."}
	}

	if flagOr(req.Probe, true) {
		if req.DryRun {
			plan.Probe = &ProbeResult{OK: true, Note: "Probe will re-read PEQ, spells_us.txt, and zone JSON after a real write."}
		} else {
			s.ProbePlan(&plan, req.ZoneIDs)
		}
	}

	if len(plan.DBWrites) == 0 && len(plan.Writes) == 0 {
		return plan.fail("pipeline had nothing to do. Add spells, kit cells, loot, an NPC spell set, or clone NPCs.")
	}
	s.complete(&plan, req.DryRun, false)
	return plan
}

func withRoles(base TargetFilter, roles ...string) TargetFilter {
	out := base
	out.Roles = roles
	return out
}

func tableForTarget(tables []LootMade, target string) int {
	for _, t := range tables {
		if roleMatch(target, t.Target) || strings.EqualFold(t.Target, target) {
			return t.TableID
		}
	}
	return 0
}

func roleMap(rows []ClassifiedNPC, remap map[int]int) map[int]string {
	out := map[int]string{}
	for _, n := range rows {
		id := n.ID
		if next, ok := remap[id]; ok {
			id = next
		}
		out[id] = n.Role
		out[n.ID] = n.Role
	}
	return out
}
