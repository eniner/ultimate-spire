package contentfactory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (s *Service) SyncZoneItemSpells(zoneIDs []int, dry bool, plan *Plan) {
	if plan == nil {
		return
	}
	ids := uniqueInts(zoneIDs)
	if len(ids) == 0 {
		plan.Warnings = append(plan.Warnings, "no zones to sync item spell ids")
		return
	}
	if s.questsDirPath() == "" {
		plan.Warnings = append(plan.Warnings, "quests directory is not set; cannot sync zone item JSON")
		return
	}
	for _, id := range ids {
		n, err := s.syncOneZoneItems(id, dry)
		if err != nil {
			plan.Warnings = append(plan.Warnings, err.Error())
			continue
		}
		plan.Synced += n
		rel := fmt.Sprintf("global/ultimatedata/%d/%d_item.json", id, id)
		plan.Writes = append(plan.Writes, FileWrite{Kind: "item-json", RelPath: rel, Action: "sync-spellid"})
	}
}

func (s *Service) syncOneZoneItems(zoneID int, dry bool) (int, error) {
	abs := filepath.Join(s.questsDirPath(), "global", "ultimatedata", strconv.Itoa(zoneID), fmt.Sprintf("%d_item.json", zoneID))
	raw, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("zone %d has no item JSON", zoneID)
		}
		return 0, err
	}
	var top map[string]interface{}
	if err := json.Unmarshal(raw, &top); err != nil {
		return 0, fmt.Errorf("parse %s: %w", filepath.Base(abs), err)
	}
	items := pickItemMap(top, zoneID)
	if len(items) == 0 {
		return 0, fmt.Errorf("zone %d item JSON has no rows", zoneID)
	}
	ids := []int{}
	for _, rawItem := range items {
		obj := asObject(rawItem)
		if n, err := strconv.Atoi(strings.TrimSpace(asString(obj["itemid"]))); err == nil && n > 0 {
			ids = append(ids, n)
		}
	}
	byID, err := s.loadItems(uniqueInts(ids))
	if err != nil {
		return 0, err
	}
	changed := 0
	for name, rawItem := range items {
		obj := asObject(rawItem)
		itemID, _ := strconv.Atoi(strings.TrimSpace(asString(obj["itemid"])))
		src, ok := byID[itemID]
		if !ok {
			continue
		}
		spell := src.Clickeffect
		if spell <= 0 {
			spell = src.Proceffect
		}
		next := "-1"
		if spell > 0 {
			next = strconv.Itoa(spell)
		}
		if asString(obj["spellid"]) == next {
			continue
		}
		obj["spellid"] = next
		items[name] = obj
		changed++
	}
	if changed == 0 || dry {
		return changed, nil
	}
	if _, ok := top[strconv.Itoa(zoneID)]; ok {
		top[strconv.Itoa(zoneID)] = items
		return changed, writeJSON(abs, top)
	}
	if looksLikeItemRoot(top) {
		return changed, writeJSON(abs, items)
	}
	return changed, writeJSON(abs, map[string]interface{}{strconv.Itoa(zoneID): items})
}

func pickItemMap(top map[string]interface{}, zoneID int) map[string]interface{} {
	if inner, ok := top[strconv.Itoa(zoneID)].(map[string]interface{}); ok {
		return inner
	}
	if looksLikeItemRoot(top) {
		return top
	}
	return map[string]interface{}{}
}

func looksLikeItemRoot(top map[string]interface{}) bool {
	for _, v := range top {
		obj := asObject(v)
		if asString(obj["itemid"]) != "" || asString(obj["spellid"]) != "" {
			return true
		}
	}
	return false
}

func asObject(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func asString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	case float64:
		if t == float64(int(t)) {
			return strconv.Itoa(int(t))
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case json.Number:
		return t.String()
	default:
		if v == nil {
			return ""
		}
		return fmt.Sprintf("%v", v)
	}
}
