package zonecontroller

import (
	"encoding/json"
	"testing"
)

func TestPickZoneObjectAndCustom(t *testing.T) {
	raw := []byte(`{"17":{"info":{"name":"Blackburrow","objective":"Kill Fippy"},"ignore":{"all":[{"name":"zone controller"},{"name":""}]},"depop":{"all":[{"name":""}]},"custom":{"":{"type":"boss"},"Fippy Darkpaw":{"mobid":"2001","type":"boss","static":"1","max_hp_mod":700000,"aggro_mod":"-1","loot":[{"chance":"300","id":"bosshead"}]}},"basedata":{"trash":{"level":"75","loot":[{"chance":3,"id":"spells"}]}}}}`)
	var top map[string]interface{}
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatal(err)
	}
	obj, err := pickZoneObject(top, 17)
	if err != nil {
		t.Fatal(err)
	}
	if asString(asObject(obj["info"])["name"]) != "Blackburrow" {
		t.Fatalf("name: %#v", obj["info"])
	}
	custom := parseCustom(obj["custom"])
	if len(custom) != 1 || custom[0].Name != "Fippy Darkpaw" || custom[0].MobID != "2001" {
		t.Fatalf("custom: %#v", custom)
	}
	if len(custom[0].Loot) != 1 || custom[0].Loot[0].ID != "bosshead" {
		t.Fatalf("loot: %#v", custom[0].Loot)
	}
	ignore := namesFromAll(obj["ignore"])
	if len(ignore) != 1 || ignore[0] != "zone controller" {
		t.Fatalf("ignore: %#v", ignore)
	}
	if n := countCustom(obj["custom"]); n != 1 {
		t.Fatalf("count custom %d", n)
	}
}
