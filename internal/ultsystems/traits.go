package ultsystems

import (
	"fmt"
	"strings"
	"time"
)

type TraitEffect struct {
	HasEffect bool   `json:"hasEffect"`
	SpellID   int    `json:"spellId"`
	SpellName string `json:"spellName"`
}

type TraitRow struct {
	ItemID        int    `json:"itemId"`
	ItemName      string `json:"itemName"`
	ZoneID        int    `json:"zoneId"`
	ZoneShortName string `json:"zoneShortName"`
	ZoneLongName  string `json:"zoneLongName"`
	Tier          string `json:"tier"`
	ZoneGroup     string `json:"zoneGroup"`
	LootTable     string `json:"lootTable"`
	Classes       string `json:"classes"`
	ClickSpell    int    `json:"clickSpell"`
	ProcSpell     int    `json:"procSpell"`
	WornSpell     int    `json:"wornSpell"`
	FocusSpell    int    `json:"focusSpell"`
}

type TraitPayload struct {
	GeneratedUTC string     `json:"generatedUtc"`
	Scope        string     `json:"scope"`
	ItemCount    int        `json:"itemCount"`
	Items        []TraitRow `json:"items"`
}

func (s *Service) Traits() (TraitPayload, error) {
	out := TraitPayload{Items: []TraitRow{}}
	doc, err := s.loadMap(traitRel)
	if err != nil {
		return out, err
	}
	out.GeneratedUTC = asString(doc["generated_utc"])
	out.Scope = asString(doc["scope"])
	for _, raw := range asArray(doc["items"]) {
		row := asObject(raw)
		effects := asObject(row["effects"])
		classes := []string{}
		for _, item := range asArray(row["classes"]) {
			if txt := strings.TrimSpace(asString(item)); txt != "" {
				classes = append(classes, txt)
			}
		}
		out.Items = append(out.Items, TraitRow{
			ItemID:        asInt(row["item_id"]),
			ItemName:      asString(row["item_name"]),
			ZoneID:        asInt(row["zone_id"]),
			ZoneShortName: asString(row["zone_short_name"]),
			ZoneLongName:  asString(row["zone_long_name"]),
			Tier:          asString(row["tier"]),
			ZoneGroup:     asString(row["zone_group"]),
			LootTable:     asString(row["loot_table"]),
			Classes:       strings.Join(classes, ", "),
			ClickSpell:    effectSpell(effects, "click"),
			ProcSpell:     effectSpell(effects, "proc"),
			WornSpell:     effectSpell(effects, "worn"),
			FocusSpell:    effectSpell(effects, "focus"),
		})
	}
	out.ItemCount = len(out.Items)
	return out, nil
}

func effectSpell(effects map[string]interface{}, key string) int {
	return asInt(asObject(effects[key])["spell_id"])
}

func (s *Service) writeTrait(plan *WritePlan, req WriteRequest) error {
	itemID := asInt(req.Entry["item_id"])
	if itemID == 0 {
		itemID = asInt(req.Entry["itemId"])
	}
	if itemID <= 0 {
		return fmt.Errorf("item_id is required")
	}
	plan.Key = fmt.Sprintf("%d", itemID)
	doc, err := s.loadMap(traitRel)
	if err != nil {
		return err
	}
	items := asArray(doc["items"])
	idx := -1
	var current map[string]interface{}
	for i, raw := range items {
		row := asObject(raw)
		if asInt(row["item_id"]) == itemID {
			idx = i
			current = row
			break
		}
	}
	if req.Action == "delete" {
		if idx < 0 {
			return fmt.Errorf("trait item %d not found", itemID)
		}
		items = append(items[:idx], items[idx+1:]...)
		doc["items"] = items
		doc["item_count"] = len(items)
		return s.commitJSON(plan, traitRel, doc)
	}
	if current == nil {
		current = map[string]interface{}{
			"item_id":         itemID,
			"item_name":       "New Trait Item",
			"zone_id":         0,
			"zone_short_name": "",
			"zone_long_name":  "",
			"tier":            "Demigod",
			"zone_group":      "",
			"loot_table":      "trait",
			"class_ids":       []interface{}{},
			"classes":         []interface{}{},
			"class_names":     []interface{}{},
			"effects": map[string]interface{}{
				"click": emptyEffect(),
				"proc":  emptyEffect(),
				"worn":  emptyEffect(),
				"focus": emptyEffect(),
			},
			"source": map[string]interface{}{"spire": true},
		}
	}
	if v := textOf(req.Entry, "item_name", "itemName"); v != "" {
		current["item_name"] = v
	}
	if req.Entry["zone_id"] != nil || req.Entry["zoneId"] != nil {
		n := asInt(req.Entry["zone_id"])
		if n == 0 {
			n = asInt(req.Entry["zoneId"])
		}
		current["zone_id"] = n
	}
	if v := textOf(req.Entry, "zone_short_name", "zoneShortName"); v != "" {
		current["zone_short_name"] = v
	}
	if v := textOf(req.Entry, "zone_long_name", "zoneLongName"); v != "" {
		current["zone_long_name"] = v
	}
	if v := textOf(req.Entry, "tier"); v != "" {
		current["tier"] = v
	}
	if v := textOf(req.Entry, "zone_group", "zoneGroup"); v != "" {
		current["zone_group"] = v
	}
	if v := textOf(req.Entry, "loot_table", "lootTable"); v != "" {
		current["loot_table"] = v
	}
	if req.Entry["classes"] != nil {
		current["classes"] = splitToArray(req.Entry["classes"])
	}
	if req.Entry["class_ids"] != nil || req.Entry["classIds"] != nil {
		raw := req.Entry["class_ids"]
		if raw == nil {
			raw = req.Entry["classIds"]
		}
		ids := intSlice(raw)
		arr := make([]interface{}, 0, len(ids))
		for _, n := range ids {
			arr = append(arr, n)
		}
		current["class_ids"] = arr
	}
	effects := asObject(current["effects"])
	setEffect(effects, "click", req.Entry, "clickSpell", "click_spell")
	setEffect(effects, "proc", req.Entry, "procSpell", "proc_spell")
	setEffect(effects, "worn", req.Entry, "wornSpell", "worn_spell")
	setEffect(effects, "focus", req.Entry, "focusSpell", "focus_spell")
	current["effects"] = effects
	if idx >= 0 {
		items[idx] = current
	} else {
		items = append(items, current)
	}
	doc["items"] = items
	doc["item_count"] = len(items)
	doc["generated_utc"] = time.Now().UTC().Format(time.RFC3339)
	return s.commitJSON(plan, traitRel, doc)
}

func emptyEffect() map[string]interface{} {
	return map[string]interface{}{
		"has_effect": false,
		"spell_id":   nil,
		"spell_name": nil,
	}
}

func setEffect(effects map[string]interface{}, key string, src map[string]interface{}, names ...string) {
	var found bool
	var n int
	for _, name := range names {
		if src[name] != nil {
			found = true
			n = asInt(src[name])
			break
		}
	}
	if !found {
		return
	}
	row := asObject(effects[key])
	if n > 0 {
		row["has_effect"] = true
		row["spell_id"] = n
	} else {
		row["has_effect"] = false
		row["spell_id"] = nil
	}
	effects[key] = row
}

func splitToArray(v interface{}) []interface{} {
	out := []interface{}{}
	switch t := v.(type) {
	case []interface{}:
		for _, item := range t {
			if s := strings.TrimSpace(asString(item)); s != "" {
				out = append(out, s)
			}
		}
	case string:
		for _, part := range strings.Split(t, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}
