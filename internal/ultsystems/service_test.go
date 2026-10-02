package ultsystems

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTalentAndRunewordDryRunThenWrite(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "global", "talentdata", "talent_catalog.json"), `{
  "schema_version": "1.0.0",
  "catalog_id": "test",
  "generated_at_utc": "2026-01-01T00:00:00Z",
  "policy": {},
  "task_selector": {},
  "point_pools": [{"pool_id":"demigod_talent","display_name":"Demigod","available_trr_key":"a","spent_trr_key":"s","max_points":80,"currency_ids":[],"access":{"min_level":1,"max_level":85,"required_tss_flags":[]}}],
  "trees": [{"tree_id":"war_t1","class_id":1,"tier":"demigod","label":"Warrior Tier 1","sort_order":10}],
  "talents": []
}`)
	mustWrite(t, filepath.Join(root, "global", "ultimatedata", "talent_rank_registry.json"), `{"metadata":{},"talent_ranks":{}}`)
	mustWrite(t, filepath.Join(root, "global", "ultimatedata", "unlock_state_registry.json"), `{"metadata":{},"unlocks":{}}`)
	mustWrite(t, filepath.Join(root, "global", "talentdata", "specializations.json"), `{"1":{}}`)
	mustWrite(t, filepath.Join(root, "global", "vendordata", "trait_items_non_god_catalog.json"), `{"generated_utc":"","scope":"test","item_count":0,"items":[]}`)

	svc := &Service{questsDir: root}
	dry := true
	plan := svc.Write(WriteRequest{
		Kind:   "talent",
		Action: "upsert",
		DryRun: &dry,
		Entry: map[string]interface{}{
			"talent_id":     "war_test_blade",
			"display_name":  "Test Blade",
			"classIds":      []interface{}{1},
			"treeId":        "war_t1",
			"pointPoolId":   "demigod_talent",
			"rankCap":       5,
			"costAmount":    2,
		},
	})
	if !plan.OK || !plan.DryRun {
		t.Fatalf("dry run talent: %#v", plan)
	}
	if _, err := os.Stat(filepath.Join(root, "global", "talentdata", "talent_catalog.json")); err != nil {
		t.Fatal(err)
	}
	cat, _ := svc.loadMap(talentRel)
	if len(asArray(cat["talents"])) != 0 {
		t.Fatalf("dry run wrote talents: %#v", cat["talents"])
	}

	write := false
	plan = svc.Write(WriteRequest{
		Kind:   "talent",
		Action: "upsert",
		DryRun: &write,
		Entry: map[string]interface{}{
			"talent_id":    "war_test_blade",
			"display_name": "Test Blade",
			"classIds":     []interface{}{1},
			"treeId":       "war_t1",
			"rankCap":      5,
		},
	})
	if !plan.OK || plan.DryRun {
		t.Fatalf("write talent: %#v", plan)
	}
	got, err := svc.Talent("war_test_blade")
	if err != nil {
		t.Fatal(err)
	}
	if asString(got["display_name"]) != "Test Blade" {
		t.Fatalf("talent name %v", got["display_name"])
	}

	plan = svc.Write(WriteRequest{
		Kind:   "rank",
		Action: "upsert",
		DryRun: &write,
		Entry:  map[string]interface{}{"key": "war_test_blade", "classFamily": "war"},
	})
	if !plan.OK {
		t.Fatalf("rank: %#v", plan)
	}
	plan = svc.Write(WriteRequest{
		Kind:   "trait",
		Action: "upsert",
		DryRun: &write,
		Entry: map[string]interface{}{
			"itemId":   149401,
			"itemName": "Executus Test",
			"zoneId":   112,
			"classes":  "WAR",
			"procSpell": 26881,
		},
	})
	if !plan.OK {
		t.Fatalf("trait: %#v", plan)
	}
	plan = svc.Write(WriteRequest{
		Kind:   "runeword_seed",
		Action: "seed",
		DryRun: &write,
		Entry:  map[string]interface{}{},
	})
	if !plan.OK {
		t.Fatalf("seed: %#v", plan)
	}
	rw := svc.Runewords()
	if !rw.Exists || len(rw.Items) < 8 || len(rw.Combos) < 10 {
		t.Fatalf("seeded runewords %#v", rw)
	}
	plan = svc.Write(WriteRequest{
		Kind:   "runeword_combo",
		Action: "upsert",
		DryRun: &write,
		Entry: map[string]interface{}{
			"comboKey":     "annihilus.normal",
			"outputItemId": 899345,
			"outputName":   "Annihilus, Blade of Destruction",
		},
	})
	if !plan.OK {
		t.Fatalf("combo: %#v", plan)
	}

	payload, err := svc.Catalog()
	if err != nil {
		t.Fatal(err)
	}
	if len(payload.Talents) != 1 || payload.Talents[0].TalentID != "war_test_blade" {
		t.Fatalf("catalog talents %#v", payload.Talents)
	}
	traits, err := svc.Traits()
	if err != nil || traits.ItemCount != 1 {
		t.Fatalf("traits %#v %v", traits, err)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	var probe interface{}
	if err := json.Unmarshal([]byte(body), &probe); err != nil {
		t.Fatal(err)
	}
}
