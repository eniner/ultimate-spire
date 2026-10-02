package zonecontroller

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRemapTopKey(t *testing.T) {
	raw := map[string]interface{}{"TEMPZONEID": map[string]interface{}{"info": map[string]interface{}{"name": "x"}}}
	out := remapTopKey(raw, 99)
	z := pickZoneMap(out, 99)
	if asString(asObject(z["info"])["name"]) != "x" {
		t.Fatalf("%#v", out)
	}
	if _, ok := out["TEMPZONEID"]; ok {
		t.Fatal("template key left behind")
	}
}

func TestCreateFromTemplateAndTier(t *testing.T) {
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
			"info":     map[string]interface{}{"name": "Zone name", "objective": "o", "tip": "t", "respawn": "r"},
			"depop":    map[string]interface{}{"all": []interface{}{}},
			"ignore":   map[string]interface{}{"all": []interface{}{map[string]interface{}{"name": "test controller"}}},
			"basedata": map[string]interface{}{"trash": map[string]interface{}{"level": "75", "max_hp": "10"}, "boss": map[string]interface{}{"level": "78"}, "raid": map[string]interface{}{"level": "80"}},
			"custom":   map[string]interface{}{},
		},
	})
	write("_loot.json", map[string]interface{}{"TEMPZONEID": map[string]interface{}{}})
	write("_item.json", map[string]interface{}{"TEMPZONEID": map[string]interface{}{}})

	s := &Service{questsDir: root}
	dry := false
	plan := s.CreateZones(CreateRequest{
		ZoneIDs:   []int{17, 17, 44},
		Source:    "template",
		Name:      "Blackburrow",
		Overwrite: false,
		DryRun:    &dry,
	})
	if !plan.OK || plan.Error != "" {
		t.Fatalf("%#v", plan)
	}
	detail, err := s.GetZone(17)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Name != "Blackburrow" {
		t.Fatalf("name %q", detail.Name)
	}
	if len(detail.Ignore) != 1 || detail.Ignore[0] != "zone controller" {
		t.Fatalf("ignore %#v", detail.Ignore)
	}

	plan2 := s.ApplyTier(TierRequest{
		ZoneIDs: []int{17, 44},
		Trash:   map[string]string{"level": "71", "max_hp": "500000"},
		DryRun:  &dry,
	})
	if !plan2.OK {
		t.Fatalf("%#v", plan2)
	}
	d17, _ := s.GetZone(17)
	if d17.Basedata["trash"]["level"] != "71" || d17.Basedata["trash"]["max_hp"] != "500000" {
		t.Fatalf("tier %#v", d17.Basedata["trash"])
	}

	plan3 := s.AddMobs(MobsRequest{
		ZoneIDs: []int{17},
		Type:    "boss",
		Names:   []string{"Fippy Darkpaw", "Fippy Darkpaw", " an elite gnoll guard "},
		DryRun:  &dry,
	})
	if !plan3.OK {
		t.Fatalf("%#v", plan3)
	}
	d17, _ = s.GetZone(17)
	if len(d17.Custom) != 2 {
		t.Fatalf("custom %#v", d17.Custom)
	}
}
