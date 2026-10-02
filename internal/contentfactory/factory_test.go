package contentfactory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNextFromMax(t *testing.T) {
	if nextFromMax(10, 800000) != 800000 {
		t.Fatal(nextFromMax(10, 800000))
	}
	if nextFromMax(800010, 800000) != 800011 {
		t.Fatal(nextFromMax(800010, 800000))
	}
}

func TestPrefixedNameAndLadder(t *testing.T) {
	if prefixedName("T1", "Helm") != "T1 Helm" {
		t.Fatal(prefixedName("T1", "Helm"))
	}
	if prefixedName("T1", "T1 Helm") != "T1 Helm" {
		t.Fatal("already prefixed")
	}
	if ladderMultiplier(1.5, 0) != 1 {
		t.Fatal(ladderMultiplier(1.5, 0))
	}
	if ladderMultiplier(2, 2) != 4 {
		t.Fatal(ladderMultiplier(2, 2))
	}
}

func TestParseIDs(t *testing.T) {
	got := parseIDs("17, 31 39\n44")
	if len(got) != 4 || got[0] != 17 || got[3] != 44 {
		t.Fatalf("%v", got)
	}
	if len(parseIDs("17, 17, 0, -3")) != 1 {
		t.Fatal(parseIDs("17, 17, 0, -3"))
	}
}

func TestNormalizeCells(t *testing.T) {
	got := normalizeCells([]KitCell{
		{SourceID: 0},
		{SourceID: 1001, Slot: "Head", Class: "WAR"},
		{SourceID: 1001, Slot: "Head", Class: "WAR"},
	})
	if len(got) != 1 {
		t.Fatalf("%#v", got)
	}
}

func TestSpellItemField(t *testing.T) {
	col, err := spellItemField("click")
	if err != nil || col != "clickeffect" {
		t.Fatal(col, err)
	}
	if _, err := spellItemField("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestSlotClassBits(t *testing.T) {
	if slotBits("Head") != 4 || classBit("WAR") != 1 {
		t.Fatal(slotBits("Head"), classBit("WAR"))
	}
	if classBit("ALL") != 65535 {
		t.Fatal(classBit("ALL"))
	}
}

func TestScaleInt(t *testing.T) {
	n := 10
	scaleInt(&n, 1.5)
	if n != 15 {
		t.Fatal(n)
	}
	neg := -8
	scaleInt(&neg, 2)
	if neg != -16 {
		t.Fatal(neg)
	}
}

func TestRecipeUpsert(t *testing.T) {
	root := t.TempDir()
	s := &Service{questsDir: root}
	plan := emptyPlan("items", false)
	err := s.upsertRecipe(Recipe{ID: "classic-t1", Name: "Classic T1", Prefix: "T1", Kit: RecipeKit{
		Items: []RecipeKitItem{{ItemID: 800001, Name: "T1 Helm", Slots: 4, Classes: 1}},
	}}, false, &plan)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := s.loadRecipes()
	if err != nil || len(rows) != 1 || rows[0].Kit.Items[0].ItemID != 800001 {
		t.Fatalf("%v %#v", err, rows)
	}
	err = s.upsertRecipe(Recipe{ID: "classic-t1", Kit: RecipeKit{
		Tables: []RecipeLootTable{{ID: "trash", Items: []string{"T1 Helm"}}},
	}}, false, &plan)
	if err != nil {
		t.Fatal(err)
	}
	rows, _ = s.loadRecipes()
	if len(rows) != 1 || len(rows[0].Kit.Items) != 1 || len(rows[0].Kit.Tables) != 1 {
		t.Fatalf("merge %#v", rows[0].Kit)
	}
}

func TestRunJournal(t *testing.T) {
	root := t.TempDir()
	s := &Service{questsDir: root}
	plan := emptyPlan("items", false)
	plan.DBWrites = []DBWrite{{Table: "items", Action: "create", ID: 800001, Note: "Helm"}}
	plan.Clones = []Clone{{Kind: "item", NewID: 800001, Name: "Helm"}}
	s.saveRun(&plan)
	if plan.RunID == "" {
		t.Fatal("missing run id")
	}
	list := s.ListRuns()
	if !list.OK || len(list.Runs) != 1 || list.Runs[0].Clones[0].NewID != 800001 {
		t.Fatalf("%#v", list)
	}
	raw, err := os.ReadFile(filepath.Join(root, "global", "ultimatedata", "_spire_runs", plan.RunID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var rec RunRecord
	if err := json.Unmarshal(raw, &rec); err != nil || rec.Kind != "items" {
		t.Fatal(err, rec)
	}
}

func TestTakeIDsSkipsExisting(t *testing.T) {
	s := &Service{}
	plan := emptyPlan("items", true)
	ids, err := s.takeIDs("items", "id", 800000, 3, reservedFloor, &plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 3 || ids[0] != 800000 || ids[2] != 800002 {
		t.Fatalf("%v", ids)
	}
}

func TestFactoryMeta(t *testing.T) {
	meta := FactoryMeta()
	if len(meta.Slots) != 18 || len(meta.Classes) != 16 {
		t.Fatalf("%d slots %d classes", len(meta.Slots), len(meta.Classes))
	}
	if meta.NpcCastFloor != 50000 || meta.SpellIDCap != 65535 {
		t.Fatal(meta.NpcCastFloor, meta.SpellIDCap)
	}
}

func TestTakeIDsCapped(t *testing.T) {
	s := &Service{}
	plan := emptyPlan("spells", true)
	ids, err := s.takeIDsCapped("spells_new", "id", 50000, 2, 50000, 65535, &plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != 50000 || ids[1] != 50001 {
		t.Fatalf("%v", ids)
	}
	if _, err := s.takeIDsCapped("spells_new", "id", 65535, 2, 50000, 65535, &plan); err == nil {
		t.Fatal("expected cap error")
	}
}

func TestPickKitCells(t *testing.T) {
	cells := pickKitCells([]CensusItem{
		{ID: 1001, Name: "Helm", Slots: 4, Classes: 1, Slot: "Head", Class: "WAR", Click: 13},
		{ID: 1002, Name: "Sword", Slots: 8192, Classes: 1, Slot: "Primary", Click: 0},
		{ID: 9, Name: "Trash", Slots: 0},
	})
	if len(cells) != 2 {
		t.Fatalf("%#v", cells)
	}
	if cells[0].Slot != "Head" || cells[0].Class != "WAR" || cells[0].ClickSpellID != 13 {
		t.Fatalf("%#v", cells[0])
	}
}

func TestMergePlan(t *testing.T) {
	dst := emptyPlan("pipeline", true)
	src := emptyPlan("items", true)
	src.Clones = []Clone{{Kind: "item", NewID: 8}}
	src.Impact = []ImpactRow{{Kind: "npc", ID: 1, From: 2, To: 3}}
	mergePlan(&dst, src)
	if len(dst.Clones) != 1 || len(dst.Impact) != 1 {
		t.Fatalf("%#v", dst)
	}
}

func TestFillLootFromItems(t *testing.T) {
	empty := []LootTableIn{{Name: "t1-trash"}}
	if lootHasItems(empty) {
		t.Fatal("empty table should have no items")
	}
	filled := fillLootFromItems(empty, []int{800001, 800002})
	if !lootHasItems(filled) || len(filled[0].Items) != 2 || filled[0].Items[0].ItemID != 800001 {
		t.Fatalf("%#v", filled)
	}
	keep := []LootTableIn{{Name: "manual", Items: []LootItemIn{{ItemID: 13, Chance: 50}}}}
	got := fillLootFromItems(keep, []int{800001})
	if len(got[0].Items) != 1 || got[0].Items[0].ItemID != 13 {
		t.Fatalf("should keep explicit loot items: %#v", got)
	}
	if lootHasItems(fillLootFromItems([]LootTableIn{{Name: "x"}}, nil)) {
		t.Fatal("no item ids should leave loot empty")
	}
}

func TestClassifyAndTarget(t *testing.T) {
	rows := []spawnAgg{
		{ID: 1, Name: "a_gnoll", Level: 5, HP: 100, Pops: 8},
		{ID: 2, Name: "Lord_Elgnub", Level: 16, HP: 800, Pops: 1},
		{ID: 3, Name: "zone_controller", Level: 1, HP: 1, Pops: 1},
		{ID: 4, Name: "Raid_Boss", Level: 50, HP: 9000, RaidTarget: 1, Pops: 1},
	}
	got := classifySpawnRows(rows)
	byID := map[int]string{}
	for _, m := range got {
		byID[m.ID] = m.Role
	}
	if byID[1] != "trash" || byID[2] != "boss" || byID[3] != "ignore" || byID[4] != "raid" {
		t.Fatalf("%#v", byID)
	}
	f := TargetFilter{Roles: []string{"trash"}}
	if !f.allows("trash") || f.allows("boss") || f.allows("ignore") {
		t.Fatal(f.roles())
	}
	named := TargetFilter{Roles: []string{"named"}}
	if !named.allows("boss") || !named.allows("raid") || named.allows("trash") {
		t.Fatal("named should be boss+raid")
	}
	if !looksNamed("Lord_Elgnub") || looksNamed("a_gnoll") {
		t.Fatal("named detect")
	}
}

func TestFreeInvSlots(t *testing.T) {
	busy := map[int]bool{22: true, 23: true}
	got := freeInvSlots(busy, 3)
	if len(got) != 3 || got[0] != 24 {
		t.Fatalf("%v", got)
	}
}

func TestPlannedNpc(t *testing.T) {
	plan := emptyPlan("npcs", true)
	plan.Clones = []Clone{{Kind: "npc", SourceID: 10, NewID: 800010, Name: "T1_gnoll"}}
	src, name := plannedNpc(&plan, 800010)
	if src != 10 || name != "T1_gnoll" {
		t.Fatalf("%d %s", src, name)
	}
	if src, _ := plannedNpc(&plan, 11); src != 0 {
		t.Fatal(src)
	}
}

func TestRoleMatchAndRemap(t *testing.T) {
	if !roleMatch("named", "boss") || roleMatch("trash", "boss") {
		t.Fatal("roleMatch")
	}
	got := remapIDs([]int{1, 2, 3}, map[int]int{2: 800002})
	if len(got) != 3 || got[1] != 800002 {
		t.Fatalf("%v", got)
	}
}
