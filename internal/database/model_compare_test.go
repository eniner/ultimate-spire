package database

import "testing"

type fakeCtx struct {
	keys map[string]interface{}
}

func (f fakeCtx) Get(k string) interface{} {
	return f.keys[k]
}

type patchSample struct {
	Name string `json:"name"`
	HP   int    `json:"hp"`
}

func TestLimitDiffToJSONKeepsOnlyPostedFields(t *testing.T) {
	c := fakeCtx{keys: map[string]interface{}{
		"jsonPatchKeys": map[string]bool{"name": true},
	}}
	diff := map[string]interface{}{"Name": "x", "HP": 0}
	out := LimitDiffToJSON(c, patchSample{}, diff)
	if _, ok := out["HP"]; ok {
		t.Fatal("HP should be dropped when it was not in the JSON body")
	}
	if out["Name"] != "x" {
		t.Fatalf("Name should stay, got %#v", out)
	}
}

func TestResultDifferenceSkipsPassword(t *testing.T) {
	type account struct {
		Name     string
		Password string
	}
	diff := ResultDifference(account{Name: "a", Password: "old"}, account{Name: "b", Password: "new"})
	if _, ok := diff["Password"]; ok {
		t.Fatal("password should not be in the update map")
	}
	if diff["Name"] != "b" {
		t.Fatalf("name should update, got %#v", diff)
	}
}
