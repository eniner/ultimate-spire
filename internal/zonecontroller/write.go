package zonecontroller

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const maxBulkZones = 80

type WriteFile struct {
	ZoneID  int    `json:"zoneId"`
	Kind    string `json:"kind"`
	RelPath string `json:"relPath"`
	Action  string `json:"action"`
}

type WritePlan struct {
	OK       bool        `json:"ok"`
	DryRun   bool        `json:"dryRun"`
	Writes   []WriteFile `json:"writes"`
	Backups  []string    `json:"backups,omitempty"`
	Warnings []string    `json:"warnings,omitempty"`
	Error    string      `json:"error,omitempty"`
}

func remapTopKey(raw map[string]interface{}, destID int) map[string]interface{} {
	key := strconv.Itoa(destID)
	var body interface{}
	if v, ok := raw[key]; ok {
		body = v
	} else if v, ok := raw["TEMPZONEID"]; ok {
		body = v
	} else {
		for _, v := range raw {
			body = v
			break
		}
	}
	if body == nil {
		body = map[string]interface{}{}
	}
	return map[string]interface{}{key: deepCopy(body)}
}

func deepCopy(v interface{}) interface{} {
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var out interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return v
	}
	return out
}

func defaultCustomMob(mobType string) map[string]interface{} {
	if mobType != "boss" && mobType != "raid" {
		mobType = "trash"
	}
	return map[string]interface{}{
		"type":                        mobType,
		"static":                      "-1",
		"mobid":                       "-1",
		"level_mod":                   "-1",
		"max_hp_mod":                  "-1",
		"min_hit_mod":                 "-1",
		"max_hit_mod":                 "-1",
		"aggro_mod":                   "-1",
		"assist_mod":                  "-1",
		"attack_speed_mod":            "-1",
		"special_attacks_override":    "-1",
		"special_abilities_override":  "-1",
		"see_invis_override":          "-1",
		"slow_mitigation_override":    "-1",
		"size_mod":                    "-1",
		"cash_override":               "-1",
		"loot":                        []interface{}{},
	}
}

func (s *Service) zoneRel(id int, kind string) string {
	return fmt.Sprintf("%s/%d/%d_%s.json", dataRel, id, id, kind)
}

func (s *Service) zoneAbs(id int, kind string) string {
	return filepath.Join(s.questsDirPath(), filepath.FromSlash(s.zoneRel(id, kind)))
}

func (s *Service) templateAbs(kind string) string {
	rel := fmt.Sprintf("%s/templates/_%s.json", dataRel, kind)
	return filepath.Join(s.questsDirPath(), filepath.FromSlash(rel))
}

func (s *Service) readKindRaw(id int, kind string) (map[string]interface{}, error) {
	return s.readJSON(s.zoneAbs(id, kind))
}

func (s *Service) readTemplateRaw(kind string) (map[string]interface{}, error) {
	return s.readJSON(s.templateAbs(kind))
}

func (s *Service) writeJSON(abs string, raw map[string]interface{}) error {
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(abs, b, 0o644)
}

func (s *Service) backupExisting(id int, kind, stamp string, plan *WritePlan) error {
	abs := s.zoneAbs(id, kind)
	st, err := os.Stat(abs)
	if err != nil || st.IsDir() {
		return nil
	}
	rel := fmt.Sprintf("%s/_spire_backups/%s/%d/%d_%s.json", dataRel, stamp, id, id, kind)
	dst := filepath.Join(s.questsDirPath(), filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, b, 0o644); err != nil {
		return err
	}
	plan.Backups = append(plan.Backups, rel)
	return nil
}

func (s *Service) commitKind(id int, kind string, raw map[string]interface{}, overwrite, dry bool, stamp string, plan *WritePlan) error {
	abs := s.zoneAbs(id, kind)
	rel := s.zoneRel(id, kind)
	_, err := os.Stat(abs)
	exists := err == nil
	action := "create"
	if exists {
		if !overwrite {
			plan.Writes = append(plan.Writes, WriteFile{ZoneID: id, Kind: kind, RelPath: rel, Action: "skip"})
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("zone %d %s exists; skipped", id, kind))
			return nil
		}
		action = "replace"
	}
	plan.Writes = append(plan.Writes, WriteFile{ZoneID: id, Kind: kind, RelPath: rel, Action: action})
	if dry {
		return nil
	}
	if exists {
		if err := s.backupExisting(id, kind, stamp, plan); err != nil {
			return err
		}
	}
	return s.writeJSON(abs, remapTopKey(raw, id))
}

func normalizeZoneIDs(ids []int) ([]int, error) {
	seen := map[int]bool{}
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid zone ids")
	}
	if len(out) > maxBulkZones {
		return nil, fmt.Errorf("too many zones (%d); max is %d", len(out), maxBulkZones)
	}
	return out, nil
}

func dryRunValue(flag *bool) bool {
	if flag == nil {
		return true
	}
	return *flag
}

func backupStamp() string {
	return time.Now().Format("20060102-150405")
}

func pickZoneMap(raw map[string]interface{}, id int) map[string]interface{} {
	obj, err := pickZoneObject(raw, id)
	if err != nil {
		return map[string]interface{}{}
	}
	return obj
}

func setInfoFields(zone map[string]interface{}, name, objective, tip, respawn string) {
	info := asObject(zone["info"])
	if strings.TrimSpace(name) != "" {
		info["name"] = name
	}
	if strings.TrimSpace(objective) != "" {
		info["objective"] = objective
	}
	if strings.TrimSpace(tip) != "" {
		info["tip"] = tip
	}
	if strings.TrimSpace(respawn) != "" {
		info["respawn"] = respawn
	}
	if info["name"] == "Zone name" && strings.TrimSpace(name) != "" {
		info["name"] = name
	}
	zone["info"] = info
}

func ensureIgnoreController(zone map[string]interface{}) {
	ignore := asObject(zone["ignore"])
	all, _ := ignore["all"].([]interface{})
	for _, item := range all {
		n := strings.ToLower(strings.TrimSpace(asString(asObject(item)["name"])))
		if n == "zone controller" {
			zone["ignore"] = ignore
			return
		}
	}
	cleaned := make([]interface{}, 0, len(all)+1)
	for _, item := range all {
		n := strings.TrimSpace(asString(asObject(item)["name"]))
		if n == "" || strings.EqualFold(n, "test controller") {
			continue
		}
		cleaned = append(cleaned, item)
	}
	cleaned = append(cleaned, map[string]interface{}{"name": "zone controller"})
	ignore["all"] = cleaned
	zone["ignore"] = ignore
}

func patchBasedataFields(zone map[string]interface{}, kind string, fields map[string]string) {
	if kind != "trash" && kind != "boss" && kind != "raid" {
		return
	}
	basedata := asObject(zone["basedata"])
	row := asObject(basedata[kind])
	for k, v := range fields {
		k = strings.TrimSpace(k)
		if k == "" || k == "loot" {
			continue
		}
		row[k] = coerceJSONValue(v)
	}
	basedata[kind] = row
	zone["basedata"] = basedata
}

func coerceJSONValue(v string) interface{} {
	v = strings.TrimSpace(v)
	if v == "" {
		return v
	}
	if i, err := strconv.ParseInt(v, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f
	}
	return v
}
