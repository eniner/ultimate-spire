package database

import (
	"fmt"
	"github.com/EQEmu/spire/internal/structs"
	"reflect"
	"strings"
)

type ModelDifference struct {
	Field string
	Old   interface{}
	New   interface{}
}

// ResultDifference will return a map[string]interface{} based on differences between
// two of the same models to be used in database updates
func ResultDifference(v1 interface{}, v2 interface{}) map[string]interface{} {
	results := make(map[string]interface{}, 0)
	for _, d := range CompareModels(v1, v2) {
		if strings.EqualFold(d.Field, "Password") || strings.EqualFold(d.Field, "password") {
			continue
		}
		results[d.Field] = d.New
	}

	return results
}

// LimitDiffToJSON drops fields that were not present in the JSON body so a
// partial PATCH cannot zero unrelated columns or replay a stale full row.
func LimitDiffToJSON(c interface{ Get(string) interface{} }, sample interface{}, diff map[string]interface{}) map[string]interface{} {
	if c == nil || len(diff) == 0 {
		return diff
	}
	raw := c.Get("jsonPatchKeys")
	keys, _ := raw.(map[string]bool)
	if len(keys) == 0 {
		return diff
	}
	allowed := jsonFieldAliases(sample, keys)
	out := make(map[string]interface{}, len(diff))
	for field, val := range diff {
		if strings.EqualFold(field, "Password") || strings.EqualFold(field, "password") {
			continue
		}
		if allowed[field] || allowed[strings.ToLower(field)] {
			out[field] = val
		}
	}
	if len(out) == 0 {
		return map[string]interface{}{}
	}
	return out
}

func jsonFieldAliases(sample interface{}, keys map[string]bool) map[string]bool {
	allowed := make(map[string]bool, len(keys)*2)
	for key := range keys {
		allowed[key] = true
		allowed[strings.ToLower(key)] = true
	}
	if sample == nil {
		return allowed
	}
	t := reflect.TypeOf(sample)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return allowed
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonName == "" || jsonName == "-" {
			jsonName = field.Name
		}
		if keys[jsonName] || keys[strings.ToLower(jsonName)] || keys[field.Name] || keys[strings.ToLower(field.Name)] {
			allowed[field.Name] = true
			allowed[jsonName] = true
			allowed[strings.ToLower(field.Name)] = true
			allowed[strings.ToLower(jsonName)] = true
		}
	}
	return allowed
}

// CompareModels will compare two of the same data models and determine
// what fields are different
// Example
//
// []database.ModelDifference{
//  database.ModelDifference{
//    Field: "type",
//    Old:   2.000000,
//    New:   1.000000,
//  },
//}
func CompareModels(v1 interface{}, v2 interface{}) []ModelDifference {
	var v1Model = structs.Map(v1)
	var v2Model = structs.Map(v2)

	var differences []ModelDifference
	for v1Field := range v1Model {
		for v2Field := range v2Model {
			fType := fmt.Sprintf("%v", reflect.TypeOf(v1Model[v1Field]))

			// struct mapper will export interfaces
			isValidField := !strings.Contains(fType, "interface")

			if v2Field == v1Field && isValidField && v1Model[v1Field] != v2Model[v1Field] {
				differences = append(differences, ModelDifference{
					Field: v1Field,
					Old:   v1Model[v1Field],
					New:   v2Model[v1Field],
				})
			}
		}
	}

	return differences
}
