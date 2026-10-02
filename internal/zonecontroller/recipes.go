package zonecontroller

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var recipeIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,47}$`)

type RecipeKitItem struct {
	ItemID  int    `json:"itemId"`
	Name    string `json:"name"`
	Slots   int    `json:"slots"`
	Classes int    `json:"classes"`
}

type RecipeLootTable struct {
	ID    string   `json:"id"`
	Count int      `json:"count"`
	Items []string `json:"items"`
}

type RecipeKit struct {
	SourceZone int               `json:"sourceZone"`
	Items      []RecipeKitItem   `json:"items"`
	Tables     []RecipeLootTable `json:"tables"`
}

type Recipe struct {
	ID       string                       `json:"id"`
	Name     string                       `json:"name"`
	Group    string                       `json:"group"`
	Step     float64                      `json:"step"`
	Prefix   string                       `json:"prefix"`
	Basedata map[string]map[string]string `json:"basedata"`
	Kit      RecipeKit                    `json:"kit"`
}

type RecipeFile struct {
	Recipes []Recipe `json:"recipes"`
}

type RecipeList struct {
	OK      bool     `json:"ok"`
	Error   string   `json:"error,omitempty"`
	RelPath string   `json:"relPath"`
	Recipes []Recipe `json:"recipes"`
}

func (s *Service) recipeAbs() string {
	return filepath.Join(s.questsDirPath(), filepath.FromSlash(recipeRel))
}

func (s *Service) ListRecipes() RecipeList {
	out := RecipeList{RelPath: recipeRel, Recipes: []Recipe{}}
	if s.questsDirPath() == "" {
		out.Error = "quests directory is not set"
		return out
	}
	rows, err := s.loadRecipes()
	if err != nil && !os.IsNotExist(err) {
		out.Error = err.Error()
		return out
	}
	out.OK = true
	out.Recipes = rows
	return out
}

func (s *Service) loadRecipes() ([]Recipe, error) {
	abs := s.recipeAbs()
	raw, err := os.ReadFile(abs)
	if err != nil {
		return []Recipe{}, err
	}
	var doc RecipeFile
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse recipes: %w", err)
	}
	if doc.Recipes == nil {
		doc.Recipes = []Recipe{}
	}
	return doc.Recipes, nil
}

func (s *Service) SaveRecipe(req Recipe, dry bool) FactoryPlan {
	plan := FactoryPlan{WritePlan: WritePlan{DryRun: dry}}
	req, err := normalizeRecipe(req)
	if err != nil {
		plan.Error = err.Error()
		return plan
	}
	rows, err := s.loadRecipes()
	if err != nil && !os.IsNotExist(err) {
		plan.Error = err.Error()
		return plan
	}
	replaced := false
	for i, row := range rows {
		if row.ID == req.ID {
			rows[i] = req
			replaced = true
			break
		}
	}
	if !replaced {
		rows = append(rows, req)
	}
	plan.Writes = append(plan.Writes, WriteFile{Kind: "recipe", RelPath: recipeRel, Action: "upsert"})
	if dry {
		plan.OK = true
		return plan
	}
	if err := s.writeRecipeFile(rows, &plan.WritePlan); err != nil {
		plan.Error = err.Error()
		return plan
	}
	plan.OK = true
	return plan
}

func (s *Service) DeleteRecipe(id string, dry bool) FactoryPlan {
	plan := FactoryPlan{WritePlan: WritePlan{DryRun: dry}}
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		plan.Error = "recipe id is required"
		return plan
	}
	rows, err := s.loadRecipes()
	if err != nil {
		plan.Error = err.Error()
		return plan
	}
	next := make([]Recipe, 0, len(rows))
	found := false
	for _, row := range rows {
		if row.ID == id {
			found = true
			continue
		}
		next = append(next, row)
	}
	if !found {
		plan.Error = "recipe not found"
		return plan
	}
	plan.Writes = append(plan.Writes, WriteFile{Kind: "recipe", RelPath: recipeRel, Action: "delete"})
	if dry {
		plan.OK = true
		return plan
	}
	if err := s.writeRecipeFile(next, &plan.WritePlan); err != nil {
		plan.Error = err.Error()
		return plan
	}
	plan.OK = true
	return plan
}

func (s *Service) writeRecipeFile(rows []Recipe, plan *WritePlan) error {
	abs := s.recipeAbs()
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(abs); err == nil {
		backupRel := fmt.Sprintf("%s/_spire_backups/%s/recipes.json", dataRel, backupStamp())
		dst := filepath.Join(s.questsDirPath(), filepath.FromSlash(backupRel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		b, err := os.ReadFile(abs)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			return err
		}
		plan.Backups = append(plan.Backups, backupRel)
	}
	body, err := json.MarshalIndent(RecipeFile{Recipes: rows}, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return os.WriteFile(abs, body, 0o644)
}

func (s *Service) RecipeFromZone(id int) (Recipe, error) {
	detail, err := s.GetZone(id)
	if err != nil {
		return Recipe{}, err
	}
	recipe := Recipe{
		ID:       "zone-" + strconv.Itoa(id),
		Name:     firstNonEmpty(detail.Name, "Zone "+strconv.Itoa(id)),
		Group:    detail.Group,
		Step:     1.35,
		Prefix:   "T1 ",
		Basedata: detail.Basedata,
		Kit: RecipeKit{
			SourceZone: id,
			Items:      []RecipeKitItem{},
			Tables:     []RecipeLootTable{},
		},
	}
	for _, it := range detail.Items {
		itemID, _ := strconv.Atoi(strings.TrimSpace(it.ItemID))
		if itemID <= 0 && strings.TrimSpace(it.Name) == "" {
			continue
		}
		recipe.Kit.Items = append(recipe.Kit.Items, RecipeKitItem{
			ItemID: itemID,
			Name:   it.Name,
		})
	}
	for _, t := range detail.Loot {
		recipe.Kit.Tables = append(recipe.Kit.Tables, RecipeLootTable{
			ID:    t.ID,
			Count: t.Count,
			Items: append([]string{}, t.Items...),
		})
	}
	s.hydrateKitItems(&recipe.Kit)
	return recipe, nil
}

func (s *Service) findRecipe(id string) (Recipe, error) {
	id = strings.ToLower(strings.TrimSpace(id))
	rows, err := s.loadRecipes()
	if err != nil && !os.IsNotExist(err) {
		return Recipe{}, err
	}
	for _, row := range rows {
		if row.ID == id {
			return row, nil
		}
	}
	return Recipe{}, fmt.Errorf("recipe %s not found", id)
}

func (s *Service) hydrateKitItems(kit *RecipeKit) {
	if kit == nil || s.eqemu() == nil {
		return
	}
	ids := make([]int, 0, len(kit.Items))
	for _, it := range kit.Items {
		if it.ItemID > 0 {
			ids = append(ids, it.ItemID)
		}
	}
	if len(ids) == 0 {
		return
	}
	byID, err := s.loadItemsByID(ids)
	if err != nil {
		return
	}
	for i, it := range kit.Items {
		src, ok := byID[it.ItemID]
		if !ok {
			continue
		}
		if strings.TrimSpace(kit.Items[i].Name) == "" {
			kit.Items[i].Name = src.Name
		}
		kit.Items[i].Slots = src.Slots
		kit.Items[i].Classes = src.Classes
	}
}

func normalizeRecipe(req Recipe) (Recipe, error) {
	req.ID = strings.ToLower(strings.TrimSpace(req.ID))
	req.Name = strings.TrimSpace(req.Name)
	req.Group = strings.TrimSpace(req.Group)
	req.Prefix = strings.TrimSpace(req.Prefix)
	if req.ID == "" {
		return req, fmt.Errorf("recipe id is required")
	}
	if !recipeIDPattern.MatchString(req.ID) {
		return req, fmt.Errorf("recipe id must be lowercase letters, numbers, underscore, or hyphen")
	}
	if req.Name == "" {
		req.Name = req.ID
	}
	if req.Step <= 0 {
		req.Step = 1.35
	}
	if req.Basedata == nil {
		req.Basedata = map[string]map[string]string{}
	}
	if req.Kit.Items == nil {
		req.Kit.Items = []RecipeKitItem{}
	}
	if req.Kit.Tables == nil {
		req.Kit.Tables = []RecipeLootTable{}
	}
	return req, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
