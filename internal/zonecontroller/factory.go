package zonecontroller

import (
	"fmt"
	"strconv"
	"strings"
)

type FactoryRequest struct {
	RecipeID    string   `json:"recipeId"`
	Recipe      *Recipe  `json:"recipe"`
	ZoneIDs     []int    `json:"zoneIds"`
	Mode        string   `json:"mode"`
	Step        float64  `json:"step"`
	Prefix      string   `json:"prefix"`
	ItemStartID int      `json:"itemStartId"`
	CreateZones bool     `json:"createZones"`
	CreateItems bool     `json:"createItems"`
	CopyGear    bool     `json:"copyGear"`
	Classify    bool     `json:"classify"`
	Merchants   bool     `json:"merchants"`
	Traits      bool     `json:"traits"`
	Evolving    bool     `json:"evolving"`
	Overwrite   bool     `json:"overwrite"`
	EvoKills    int64    `json:"evoKills"`
	DryRun      *bool    `json:"dryRun"`
}

type FactoryStep struct {
	Index      int                         `json:"index"`
	ZoneID     int                         `json:"zoneId"`
	Multiplier float64                     `json:"multiplier"`
	Prefix     string                      `json:"prefix"`
	Basedata   map[string]map[string]string `json:"basedata"`
	ItemStart  int                         `json:"itemStart"`
	ItemCount  int                         `json:"itemCount"`
}

type FactoryPlan struct {
	WritePlan
	RecipeID   string          `json:"recipeId,omitempty"`
	Mode       string          `json:"mode,omitempty"`
	Steps      []FactoryStep   `json:"steps,omitempty"`
	Coverage   []CoverageCell  `json:"coverage,omitempty"`
	Missing    int             `json:"coverageMissing,omitempty"`
	ItemClones []ItemClone     `json:"itemClones,omitempty"`
	Issues     []ValidateIssue `json:"issues,omitempty"`
	DBWrites   []DBWrite       `json:"dbWrites,omitempty"`
	Commands   []string        `json:"commands,omitempty"`
}

func (s *Service) RunFactory(req FactoryRequest) FactoryPlan {
	plan := FactoryPlan{WritePlan: WritePlan{DryRun: dryRunValue(req.DryRun)}, Steps: []FactoryStep{}, ItemClones: []ItemClone{}, DBWrites: []DBWrite{}}
	recipe, err := s.resolveRecipe(req)
	if err != nil {
		plan.Error = err.Error()
		return plan
	}
	ids, err := normalizeZoneIDs(req.ZoneIDs)
	if err != nil {
		plan.Error = err.Error()
		return plan
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode != "ladder" {
		mode = "stamp"
	}
	step := req.Step
	if step <= 0 {
		step = recipe.Step
	}
	if step <= 0 {
		step = 1.35
	}
	s.hydrateKitItems(&recipe.Kit)
	plan.RecipeID = recipe.ID
	plan.Mode = mode
	plan.Coverage = kitCoverage(recipe.Kit.Items)
	plan.Missing = coverageMissing(plan.Coverage)

	start, err := s.nextItemID(req.ItemStartID)
	if err != nil {
		plan.Error = err.Error()
		return plan
	}
	kitCount := 1
	if mode == "ladder" {
		kitCount = len(ids)
	}
	if req.CreateItems && len(recipe.Kit.Items) > 0 && start+len(recipe.Kit.Items)*kitCount-1 >= 2000000000 {
		plan.Error = "item id range is too large"
		return plan
	}

	for i, id := range ids {
		mult := step
		if mode == "ladder" {
			mult = ladderMultiplier(step, i)
		} else if req.Step <= 0 {
			mult = 1
		}
		prefix := strings.TrimSpace(req.Prefix)
		if prefix == "" {
			prefix = strings.TrimSpace(recipe.Prefix)
		}
		if mode == "ladder" {
			stepName := fmt.Sprintf("T%d", i+1)
			extra := strings.TrimSpace(prefix)
			if len(extra) >= 2 && (extra[0] == 'T' || extra[0] == 't') && extra[1] >= '0' && extra[1] <= '9' {
				extra = strings.TrimSpace(extra[2:])
			}
			if extra == "" {
				prefix = stepName
			} else {
				prefix = stepName + " " + extra
			}
		}
		plan.Steps = append(plan.Steps, FactoryStep{
			Index:      i,
			ZoneID:     id,
			Multiplier: mult,
			Prefix:     strings.TrimSpace(prefix),
			Basedata:   scaleBasedata(recipe.Basedata, mult),
			ItemStart:  start + i*len(recipe.Kit.Items),
			ItemCount:  len(recipe.Kit.Items),
		})
	}

	stamp := backupStamp()
	stepClones := make([][]ItemClone, 0, len(plan.Steps))
	sharedClones := []ItemClone{}
	sharedNames := map[string]int{}

	if mode == "stamp" && req.CreateItems && len(recipe.Kit.Items) > 0 {
		clones, names, err := s.cloneKitItems(recipe.Kit, start, strings.TrimSpace(firstNonEmpty(req.Prefix, recipe.Prefix)), 0, plan.Steps[0].Multiplier, plan.DryRun, &plan)
		if err != nil {
			plan.Error = err.Error()
			return plan
		}
		sharedClones = clones
		sharedNames = names
		plan.ItemClones = append(plan.ItemClones, clones...)
	}

	for i, stepRow := range plan.Steps {
		if req.CreateZones {
			if _, err := s.readKindRaw(stepRow.ZoneID, "mob"); err != nil {
				create := s.CreateZones(CreateRequest{
					ZoneIDs:   []int{stepRow.ZoneID},
					Source:    "template",
					Overwrite: req.Overwrite,
					DryRun:    req.DryRun,
					Name:      recipe.Name,
				})
				plan.Writes = append(plan.Writes, create.Writes...)
				plan.Backups = append(plan.Backups, create.Backups...)
				plan.Warnings = append(plan.Warnings, create.Warnings...)
				if create.Error != "" {
					plan.Error = create.Error
					return plan
				}
			}
		}
		raw, err := s.readKindRaw(stepRow.ZoneID, "mob")
		if err != nil {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("zone %d has no mob json; enable create zones", stepRow.ZoneID))
			continue
		}
		mob := remapTopKey(raw, stepRow.ZoneID)
		zone := pickZoneMap(mob, stepRow.ZoneID)
		applyBasedata(zone, stepRow.Basedata)
		if recipe.Group != "" {
			info := asObject(zone["info"])
			info["group"] = recipe.Group
			zone["info"] = info
		}
		mob[strconv.Itoa(stepRow.ZoneID)] = zone
		if err := s.commitKind(stepRow.ZoneID, "mob", mob, true, plan.DryRun, stamp, &plan.WritePlan); err != nil {
			plan.Error = err.Error()
			return plan
		}

		clones := sharedClones
		names := sharedNames
		prefix := stepRow.Prefix
		if mode == "ladder" && req.CreateItems && len(recipe.Kit.Items) > 0 {
			var err error
			clones, names, err = s.cloneKitItems(recipe.Kit, stepRow.ItemStart, prefix, i, stepRow.Multiplier, plan.DryRun, &plan)
			if err != nil {
				plan.Error = err.Error()
				return plan
			}
			plan.ItemClones = append(plan.ItemClones, clones...)
			stepClones = append(stepClones, clones)
		} else if mode == "stamp" {
			stepClones = [][]ItemClone{sharedClones}
		}

		if req.CopyGear && (len(clones) > 0 || len(recipe.Kit.Tables) > 0) {
			itemBody := remapTopKey(map[string]interface{}{"TEMPZONEID": buildItemJSON(clones)}, stepRow.ZoneID)
			if err := s.commitKind(stepRow.ZoneID, "item", itemBody, true, plan.DryRun, stamp, &plan.WritePlan); err != nil {
				plan.Error = err.Error()
				return plan
			}
			lootBody := remapTopKey(map[string]interface{}{"TEMPZONEID": buildLootJSON(recipe.Kit.Tables, names, prefix)}, stepRow.ZoneID)
			if err := s.commitKind(stepRow.ZoneID, "loot", lootBody, true, plan.DryRun, stamp, &plan.WritePlan); err != nil {
				plan.Error = err.Error()
				return plan
			}
		}

		if req.Classify {
			classPlan := s.ClassifyZones(ClassifyRequest{
				ZoneIDs:   []int{stepRow.ZoneID},
				Apply:     true,
				Overwrite: req.Overwrite,
				DryRun:    req.DryRun,
			})
			plan.Writes = append(plan.Writes, classPlan.Writes...)
			plan.Warnings = append(plan.Warnings, classPlan.Warnings...)
			if classPlan.Error != "" {
				plan.Warnings = append(plan.Warnings, classPlan.Error)
			}
		}
		if req.Merchants && len(clones) > 0 {
			s.writeMerchantKit(stepRow.ZoneID, clones, plan.DryRun, &plan)
		}
		if req.Traits && len(clones) > 0 {
			s.writeTraitRows(stepRow.ZoneID, recipe.Group, stepRow.Prefix, clones, plan.DryRun, &plan)
		}
	}

	if req.Evolving && mode == "ladder" && req.CreateItems {
		s.writeEvolvingChains(stepClones, ids, req.EvoKills, plan.DryRun, &plan)
	}

	plan.Commands = applyCommands(ids)
	plan.Issues = s.ValidateZones(ids).Issues
	plan.OK = plan.Error == ""
	return plan
}

func (s *Service) resolveRecipe(req FactoryRequest) (Recipe, error) {
	if req.Recipe != nil && strings.TrimSpace(req.Recipe.ID+req.Recipe.Name) != "" {
		recipe := *req.Recipe
		if recipe.ID == "" {
			recipe.ID = "inline"
		}
		if recipe.Kit.SourceZone > 0 && len(recipe.Kit.Items) == 0 {
			from, err := s.RecipeFromZone(recipe.Kit.SourceZone)
			if err == nil {
				recipe.Kit = from.Kit
				if len(recipe.Basedata) == 0 {
					recipe.Basedata = from.Basedata
				}
			}
		}
		return normalizeRecipe(recipe)
	}
	if strings.TrimSpace(req.RecipeID) == "" {
		return Recipe{}, fmt.Errorf("recipe id is required")
	}
	recipe, err := s.findRecipe(req.RecipeID)
	if err != nil {
		return Recipe{}, err
	}
	if recipe.Kit.SourceZone > 0 && len(recipe.Kit.Items) == 0 {
		from, err := s.RecipeFromZone(recipe.Kit.SourceZone)
		if err == nil {
			recipe.Kit = from.Kit
		}
	}
	return recipe, nil
}
