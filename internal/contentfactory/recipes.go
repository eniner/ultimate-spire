package contentfactory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

func (s *Service) recipeAbs() string {
	return filepath.Join(s.questsDirPath(), filepath.FromSlash(recipeRel))
}

func (s *Service) loadRecipes() ([]Recipe, error) {
	raw, err := os.ReadFile(s.recipeAbs())
	if err != nil {
		if os.IsNotExist(err) {
			return []Recipe{}, nil
		}
		return nil, err
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

func (s *Service) upsertRecipe(next Recipe, dry bool, plan *Plan) error {
	id := strings.ToLower(strings.TrimSpace(next.ID))
	if !recipeIDPattern.MatchString(id) {
		return fmt.Errorf("recipe id must be lowercase letters, numbers, dash, underscore")
	}
	next.ID = id
	if strings.TrimSpace(next.Name) == "" {
		next.Name = id
	}
	if next.Basedata == nil {
		next.Basedata = map[string]map[string]string{
			"trash": {},
			"boss":  {},
			"raid":  {},
		}
	}
	rows, err := s.loadRecipes()
	if err != nil {
		return err
	}
	replaced := false
	for i, row := range rows {
		if row.ID == next.ID {
			if len(next.Kit.Items) == 0 {
				next.Kit.Items = row.Kit.Items
			}
			if len(next.Kit.Tables) == 0 {
				next.Kit.Tables = row.Kit.Tables
			}
			if next.Step == 0 {
				next.Step = row.Step
			}
			if next.Prefix == "" {
				next.Prefix = row.Prefix
			}
			if next.Group == "" {
				next.Group = row.Group
			}
			rows[i] = next
			replaced = true
			break
		}
	}
	if !replaced {
		rows = append(rows, next)
	}
	plan.Writes = append(plan.Writes, FileWrite{Kind: "recipe", RelPath: recipeRel, Action: "upsert"})
	if dry {
		return nil
	}
	return writeJSON(s.recipeAbs(), RecipeFile{Recipes: rows})
}

func (s *Service) exportItemRecipe(req ItemRequest, clones []Clone, dry bool, plan *Plan) error {
	if s.questsDirPath() == "" {
		return fmt.Errorf("quests directory is not set")
	}
	items := []RecipeKitItem{}
	for _, clone := range clones {
		if clone.Step > 0 {
			continue
		}
		items = append(items, RecipeKitItem{
			ItemID:  clone.NewID,
			Name:    clone.Name,
			Slots:   clone.Slots,
			Classes: clone.Classes,
		})
	}
	recipe := Recipe{
		ID:     req.RecipeID,
		Name:   firstNonEmpty(req.RecipeName, req.RecipeID),
		Group:  firstNonEmpty(req.Group, req.RecipeID),
		Step:   req.Step,
		Prefix: req.Prefix,
		Kit:    RecipeKit{Items: items},
	}
	return s.upsertRecipe(recipe, dry, plan)
}

func (s *Service) exportLootRecipe(req LootRequest, tables []LootMade, dry bool, plan *Plan) error {
	if s.questsDirPath() == "" {
		return fmt.Errorf("quests directory is not set")
	}
	rows := []RecipeLootTable{}
	for i, table := range tables {
		id := strings.ToLower(strings.TrimSpace(table.Name))
		id = recipeIDPattern.FindString(strings.ReplaceAll(id, " ", "-"))
		if id == "" {
			id = fmt.Sprintf("table-%d", i+1)
		}
		names := []string{}
		if i < len(req.Tables) {
			for _, it := range req.Tables[i].Items {
				names = append(names, fmt.Sprintf("%d", it.ItemID))
			}
		}
		rows = append(rows, RecipeLootTable{ID: id, Count: 1, Items: names})
	}
	recipe := Recipe{
		ID:  req.RecipeID,
		Name: req.RecipeID,
		Kit: RecipeKit{Tables: rows},
	}
	return s.upsertRecipe(recipe, dry, plan)
}
