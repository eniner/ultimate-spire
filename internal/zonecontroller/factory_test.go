package zonecontroller

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestScaleBasedataAndNames(t *testing.T) {
	src := map[string]map[string]string{
		"trash": {"level": "70", "max_hp": "100000", "loot": "trash@3"},
		"boss":  {"level": "75", "max_hp": "500000"},
		"raid":  {"level": "80"},
	}
	out := scaleBasedata(src, 1.5)
	if out["trash"]["level"] != "105" || out["trash"]["max_hp"] != "150000" {
		t.Fatalf("scaled %#v", out["trash"])
	}
	if out["trash"]["loot"] != "trash@3" {
		t.Fatalf("loot mutated %#v", out["trash"])
	}
	if prefixedName("T2", "Rusty Sword") != "T2 Rusty Sword" {
		t.Fatal(prefixedName("T2", "Rusty Sword"))
	}
	if ladderMultiplier(1.35, 0) != 1 || ladderMultiplier(2, 2) != 4 {
		t.Fatalf("ladder %v %v", ladderMultiplier(1.35, 0), ladderMultiplier(2, 2))
	}
}

func TestClassifyHeuristics(t *testing.T) {
	rows := []spawnAgg{
		{ID: 1, Name: "zone_controller", HP: 1, Pops: 1},
		{ID: 2, Name: "a_gnoll", HP: 100, Pops: 12},
		{ID: 3, Name: "Fippy_Darkpaw", HP: 400, Pops: 1},
		{ID: 4, Name: "Lord_Elgnub", HP: 900, Pops: 1, RaidTarget: 1},
	}
	got := classifySpawnRows(rows)
	by := map[string]string{}
	for _, m := range got {
		by[m.Name] = m.Type
	}
	if by["zone controller"] != "ignore" {
		t.Fatalf("controller %#v", got)
	}
	if by["a gnoll"] != "trash" {
		t.Fatalf("trash %#v", got)
	}
	if by["Fippy Darkpaw"] != "boss" {
		t.Fatalf("boss %#v", got)
	}
	if by["Lord Elgnub"] != "raid" {
		t.Fatalf("raid %#v", got)
	}
}

func TestKitCoverage(t *testing.T) {
	cells := kitCoverage([]RecipeKitItem{
		{Name: "Helm", Slots: 4, Classes: 1},
		{Name: "Chest", Slots: 131072, Classes: 65535},
	})
	helmWAR := false
	chestCLR := false
	legsWAR := true
	for _, c := range cells {
		if c.Slot == "Head" && c.Class == "WAR" {
			helmWAR = c.Filled
		}
		if c.Slot == "Chest" && c.Class == "CLR" {
			chestCLR = c.Filled
		}
		if c.Slot == "Legs" && c.Class == "WAR" {
			legsWAR = c.Filled
		}
	}
	if !helmWAR || !chestCLR || legsWAR {
		t.Fatalf("coverage helm=%v chest=%v legs=%v missing=%d", helmWAR, chestCLR, legsWAR, coverageMissing(cells))
	}
}

func TestRecipeAndFactoryDryRun(t *testing.T) {
	root := t.TempDir()
	data := filepath.Join(root, "global", "ultimatedata")
	if err := os.MkdirAll(filepath.Join(data, "templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name string, v interface{}) {
		b, _ := json.Marshal(v)
		if err := os.WriteFile(filepath.Join(data, "templates", name), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("_mob.json", map[string]interface{}{
		"TEMPZONEID": map[string]interface{}{
			"info":     map[string]interface{}{"name": "Zone name"},
			"depop":    map[string]interface{}{"all": []interface{}{}},
			"ignore":   map[string]interface{}{"all": []interface{}{}},
			"basedata": map[string]interface{}{"trash": map[string]interface{}{"level": "70", "max_hp": "1000"}, "boss": map[string]interface{}{"level": "75"}, "raid": map[string]interface{}{"level": "80"}},
			"custom":   map[string]interface{}{},
		},
	})
	write("_loot.json", map[string]interface{}{"TEMPZONEID": map[string]interface{}{}})
	write("_item.json", map[string]interface{}{"TEMPZONEID": map[string]interface{}{}})

	s := &Service{questsDir: root}
	dry := false
	if plan := s.CreateZones(CreateRequest{ZoneIDs: []int{17}, Source: "template", DryRun: &dry}); !plan.OK {
		t.Fatalf("create %#v", plan)
	}
	from, err := s.RecipeFromZone(17)
	if err != nil {
		t.Fatal(err)
	}
	from.ID = "classic-t1"
	from.Group = "classic-t1"
	from.Basedata["trash"]["level"] = "71"
	if plan := s.SaveRecipe(from, false); !plan.OK {
		t.Fatalf("save %#v", plan)
	}
	listed := s.ListRecipes()
	if !listed.OK || len(listed.Recipes) != 1 || listed.Recipes[0].ID != "classic-t1" {
		t.Fatalf("list %#v", listed)
	}

	preview := true
	plan := s.RunFactory(FactoryRequest{
		RecipeID:    "classic-t1",
		ZoneIDs:     []int{31, 39},
		Mode:        "ladder",
		Step:        2,
		CreateZones: true,
		CopyGear:    true,
		DryRun:      &preview,
	})
	if !plan.OK || plan.Error != "" {
		t.Fatalf("factory %#v", plan)
	}
	if len(plan.Steps) != 2 {
		t.Fatalf("steps %#v", plan.Steps)
	}
	if plan.Steps[0].Basedata["trash"]["level"] != "71" {
		t.Fatalf("t1 %#v", plan.Steps[0].Basedata)
	}
	if plan.Steps[1].Basedata["trash"]["level"] != "142" {
		t.Fatalf("t2 %#v", plan.Steps[1].Basedata)
	}
	if len(plan.Commands) != 2 {
		t.Fatalf("commands %#v", plan.Commands)
	}

	wrote := s.RunFactory(FactoryRequest{
		RecipeID:    "classic-t1",
		ZoneIDs:     []int{31},
		Mode:        "stamp",
		CreateZones: true,
		CopyGear:    false,
		DryRun:      &dry,
	})
	if !wrote.OK {
		t.Fatalf("write %#v", wrote)
	}
	d31, err := s.GetZone(31)
	if err != nil {
		t.Fatal(err)
	}
	if d31.Basedata["trash"]["level"] != "71" {
		t.Fatalf("stamped %#v", d31.Basedata)
	}
	if d31.Group != "classic-t1" {
		t.Fatalf("group %q", d31.Group)
	}
	rep := s.ValidateZones([]int{31})
	if !rep.OK {
		t.Fatalf("validate %#v", rep)
	}
}
