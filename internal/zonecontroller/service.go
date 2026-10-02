package zonecontroller

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/EQEmu/spire/internal/database"
	"github.com/EQEmu/spire/internal/eqemuserver"
	"github.com/EQEmu/spire/internal/pathmgmt"
	"gorm.io/gorm"
)

const (
	maxJSONBytes = 8 << 20
	dataRel      = "global/ultimatedata"
	scriptRel    = "global/zone_controller.pl"
	recipeRel    = "global/ultimatedata/_spire_recipes/recipes.json"
	commandRel   = "global/ultimatedata/_spire_commands"
	traitRel     = "global/vendordata/trait_items_non_god_catalog.json"
)

type Service struct {
	pathmgmt  *pathmgmt.PathManagement
	questsDir string
	db        *database.Resolver
	world     *eqemuserver.Client
}

func NewService(pathmgmt *pathmgmt.PathManagement, db *database.Resolver, world *eqemuserver.Client) *Service {
	return &Service{pathmgmt: pathmgmt, db: db, world: world}
}

func (s *Service) eqemu() *gorm.DB {
	if s.db == nil {
		return nil
	}
	return s.db.GetEqemuDb()
}

func (s *Service) questsDirPath() string {
	if strings.TrimSpace(s.questsDir) != "" {
		return s.questsDir
	}
	if s.pathmgmt == nil {
		return ""
	}
	return strings.TrimSpace(s.pathmgmt.GetQuestsDir())
}

type FileInfo struct {
	Kind    string `json:"kind"`
	RelPath string `json:"relPath"`
	Exists  bool   `json:"exists"`
	Size    int64  `json:"size"`
}

type Status struct {
	OK           bool   `json:"ok"`
	Error        string `json:"error,omitempty"`
	QuestsDir    string `json:"questsDir"`
	DataDir      string `json:"dataDir"`
	DataRel      string `json:"dataRel"`
	ScriptRel    string `json:"scriptRel"`
	ScriptExists bool   `json:"scriptExists"`
	ZoneCount    int    `json:"zoneCount"`
	TalentCount  int    `json:"talentCount"`
	UnlockCount  int    `json:"unlockCount"`
}

type ZoneSummary struct {
	ZoneID         int               `json:"zoneId"`
	Name           string            `json:"name"`
	Objective      string            `json:"objective"`
	Tip            string            `json:"tip"`
	Respawn        string            `json:"respawn"`
	CustomCount    int               `json:"customCount"`
	IgnoreCount    int               `json:"ignoreCount"`
	DepopCount     int               `json:"depopCount"`
	LootTableCount int               `json:"lootTableCount"`
	ItemCount      int               `json:"itemCount"`
	Group          string            `json:"group,omitempty"`
	Files          map[string]FileInfo `json:"files"`
}

type LootRow struct {
	Chance string `json:"chance"`
	ID     string `json:"id"`
}

type CustomMob struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	MobID    string            `json:"mobId"`
	Static   string            `json:"static"`
	Mods     map[string]string `json:"mods"`
	Loot     []LootRow         `json:"loot"`
}

type LootTable struct {
	ID    string   `json:"id"`
	Count int      `json:"count"`
	Items []string `json:"items"`
}

type ItemRow struct {
	Name        string `json:"name"`
	ItemID      string `json:"itemId"`
	SpellID     string `json:"spellId"`
	GlobalKey   string `json:"globalKey"`
	IsGlobal    string `json:"isGlobal"`
	Description string `json:"description"`
}

type ZoneDetail struct {
	ZoneSummary
	Basedata map[string]map[string]string `json:"basedata"`
	Custom   []CustomMob                  `json:"custom"`
	Ignore   []string                     `json:"ignore"`
	Depop    []string                     `json:"depop"`
	Loot     []LootTable                  `json:"loot"`
	Items    []ItemRow                    `json:"items"`
}

type TalentRow struct {
	Key             string `json:"key"`
	ClassFamily     string `json:"classFamily"`
	BucketSuffix    string `json:"bucketSuffix"`
	LegacyGlobalKey string `json:"legacyGlobalKey"`
	ValueType       string `json:"valueType"`
	StateType       string `json:"stateType"`
}

type UnlockRow struct {
	Key             string `json:"key"`
	SpellID         string `json:"spellId"`
	SpellName       string `json:"spellName"`
	BucketSuffix    string `json:"bucketSuffix"`
	ExpectedValue   string `json:"expectedValue"`
	LegacyGlobalKey string `json:"legacyGlobalKey"`
	Notes           string `json:"notes"`
}

type Systems struct {
	Talents  []TalentRow            `json:"talents"`
	Unlocks  []UnlockRow            `json:"unlocks"`
	TalentMeta map[string]interface{} `json:"talentMeta,omitempty"`
	UnlockMeta map[string]interface{} `json:"unlockMeta,omitempty"`
}

func (s *Service) dataDir() string {
	return filepath.Join(s.questsDirPath(), filepath.FromSlash(dataRel))
}

func (s *Service) Status() Status {
	st := Status{
		QuestsDir: s.questsDirPath(),
		DataDir:   s.dataDir(),
		DataRel:   dataRel,
		ScriptRel: scriptRel,
	}
	if st.QuestsDir == "" {
		st.Error = "quests directory is not set; run Spire from the EQ server folder or set SPIRE_QUESTS_ROOT"
		return st
	}
	info, err := os.Stat(st.DataDir)
	if err != nil || !info.IsDir() {
		st.Error = fmt.Sprintf("ultimatedata not found at %s", st.DataDir)
		return st
	}
	if _, err := os.Stat(filepath.Join(s.questsDirPath(), filepath.FromSlash(scriptRel))); err == nil {
		st.ScriptExists = true
	}
	zones, err := s.ListZones()
	if err != nil {
		st.Error = err.Error()
		return st
	}
	st.OK = true
	st.ZoneCount = len(zones)
	if sys, err := s.Systems(); err == nil {
		st.TalentCount = len(sys.Talents)
		st.UnlockCount = len(sys.Unlocks)
	}
	return st
}

func (s *Service) ListZones() ([]ZoneSummary, error) {
	dir := s.dataDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read ultimatedata: %w", err)
	}
	out := make([]ZoneSummary, 0)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id, err := strconv.Atoi(e.Name())
		if err != nil || id <= 0 {
			continue
		}
		sum, err := s.summarizeZone(id)
		if err != nil {
			continue
		}
		out = append(out, sum)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ZoneID < out[j].ZoneID })
	return out, nil
}

func (s *Service) GetZone(id int) (ZoneDetail, error) {
	if id <= 0 {
		return ZoneDetail{}, errors.New("invalid zone id")
	}
	sum, err := s.summarizeZone(id)
	if err != nil {
		return ZoneDetail{}, err
	}
	detail := ZoneDetail{
		ZoneSummary: sum,
		Basedata:    map[string]map[string]string{},
		Custom:      []CustomMob{},
		Ignore:      []string{},
		Depop:       []string{},
		Loot:        []LootTable{},
		Items:       []ItemRow{},
	}
	if mob, err := s.readZoneObject(id, "mob"); err == nil {
		detail.Basedata = parseBasedata(mob["basedata"])
		detail.Custom = parseCustom(mob["custom"])
		detail.Ignore = namesFromAll(mob["ignore"])
		detail.Depop = namesFromAll(mob["depop"])
	}
	if loot, err := s.readZoneObject(id, "loot"); err == nil {
		detail.Loot = parseLootTables(loot)
		detail.LootTableCount = len(detail.Loot)
	}
	if items, err := s.readZoneObject(id, "item"); err == nil {
		detail.Items = parseItems(items)
		detail.ItemCount = len(detail.Items)
	}
	return detail, nil
}

func (s *Service) Systems() (Systems, error) {
	sys := Systems{
		Talents: []TalentRow{},
		Unlocks: []UnlockRow{},
	}
	if raw, err := s.readRootJSON("talent_rank_registry.json"); err == nil {
		sys.TalentMeta = asObject(raw["metadata"])
		sys.Talents = parseTalents(raw["talent_ranks"])
	}
	if raw, err := s.readRootJSON("unlock_state_registry.json"); err == nil {
		sys.UnlockMeta = asObject(raw["metadata"])
		sys.Unlocks = parseUnlocks(raw["unlocks"])
	}
	if len(sys.Talents) == 0 && len(sys.Unlocks) == 0 {
		if _, err := os.Stat(s.dataDir()); err != nil {
			return sys, fmt.Errorf("ultimatedata not found")
		}
	}
	return sys, nil
}

func (s *Service) summarizeZone(id int) (ZoneSummary, error) {
	files := map[string]FileInfo{
		"mob":  s.statKind(id, "mob"),
		"loot": s.statKind(id, "loot"),
		"item": s.statKind(id, "item"),
	}
	if !files["mob"].Exists && !files["loot"].Exists && !files["item"].Exists {
		return ZoneSummary{}, fmt.Errorf("no json for zone %d", id)
	}
	sum := ZoneSummary{ZoneID: id, Files: files}
	if mob, err := s.readZoneObject(id, "mob"); err == nil {
		info := asObject(mob["info"])
		sum.Name = asString(info["name"])
		sum.Objective = asString(info["objective"])
		sum.Tip = asString(info["tip"])
		sum.Respawn = asString(info["respawn"])
		sum.Group = asString(info["group"])
		if sum.Group == "" {
			sum.Group = asString(info["tier_group"])
		}
		sum.CustomCount = countCustom(mob["custom"])
		sum.IgnoreCount = len(namesFromAll(mob["ignore"]))
		sum.DepopCount = len(namesFromAll(mob["depop"]))
	}
	if loot, err := s.readZoneObject(id, "loot"); err == nil {
		sum.LootTableCount = len(loot)
	}
	if items, err := s.readZoneObject(id, "item"); err == nil {
		sum.ItemCount = len(items)
	}
	return sum, nil
}

func (s *Service) statKind(id int, kind string) FileInfo {
	rel := fmt.Sprintf("%s/%d/%d_%s.json", dataRel, id, id, kind)
	abs := filepath.Join(s.questsDirPath(), filepath.FromSlash(rel))
	info := FileInfo{Kind: kind, RelPath: rel}
	if st, err := os.Stat(abs); err == nil && !st.IsDir() {
		info.Exists = true
		info.Size = st.Size()
	}
	return info
}

func (s *Service) readZoneObject(id int, kind string) (map[string]interface{}, error) {
	rel := fmt.Sprintf("%s/%d/%d_%s.json", dataRel, id, id, kind)
	raw, err := s.readJSON(filepath.Join(s.questsDirPath(), filepath.FromSlash(rel)))
	if err != nil {
		return nil, err
	}
	return pickZoneObject(raw, id)
}

func (s *Service) readRootJSON(name string) (map[string]interface{}, error) {
	return s.readJSON(filepath.Join(s.dataDir(), name))
}

func (s *Service) readJSON(abs string) (map[string]interface{}, error) {
	st, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if st.Size() > maxJSONBytes {
		return nil, fmt.Errorf("json too large: %s", filepath.Base(abs))
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", filepath.Base(abs), err)
	}
	return out, nil
}

func pickZoneObject(top map[string]interface{}, id int) (map[string]interface{}, error) {
	key := strconv.Itoa(id)
	if obj := asObject(top[key]); obj != nil {
		return obj, nil
	}
	for k, v := range top {
		if obj := asObject(v); obj != nil && k != "" {
			return obj, nil
		}
	}
	return nil, fmt.Errorf("zone %d object missing", id)
}

func parseBasedata(v interface{}) map[string]map[string]string {
	src := asObject(v)
	out := map[string]map[string]string{}
	for _, kind := range []string{"trash", "boss", "raid"} {
		row := map[string]string{}
		for k, val := range asObject(src[kind]) {
			if k == "loot" {
				continue
			}
			row[k] = asString(val)
		}
		if loot := parseLootRows(asObject(src[kind])["loot"]); len(loot) > 0 {
			parts := make([]string, 0, len(loot))
			for _, l := range loot {
				parts = append(parts, l.ID+"@"+l.Chance)
			}
			row["loot"] = strings.Join(parts, ", ")
		}
		out[kind] = row
	}
	return out
}

func parseCustom(v interface{}) []CustomMob {
	src := asObject(v)
	names := make([]string, 0, len(src))
	for name := range src {
		if strings.TrimSpace(name) == "" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]CustomMob, 0, len(names))
	for _, name := range names {
		obj := asObject(src[name])
		mob := CustomMob{
			Name:   name,
			Type:   asString(obj["type"]),
			MobID:  asString(obj["mobid"]),
			Static: asString(obj["static"]),
			Mods:   map[string]string{},
			Loot:   parseLootRows(obj["loot"]),
		}
		for k, val := range obj {
			switch k {
			case "type", "mobid", "static", "loot":
				continue
			}
			s := asString(val)
			if s == "" || s == "-1" {
				continue
			}
			mob.Mods[k] = s
		}
		out = append(out, mob)
	}
	return out
}

func parseLootRows(v interface{}) []LootRow {
	arr, ok := v.([]interface{})
	if !ok {
		return []LootRow{}
	}
	out := make([]LootRow, 0, len(arr))
	for _, item := range arr {
		obj := asObject(item)
		id := strings.TrimSpace(asString(obj["id"]))
		if id == "" {
			continue
		}
		out = append(out, LootRow{Chance: asString(obj["chance"]), ID: id})
	}
	return out
}

func parseLootTables(v map[string]interface{}) []LootTable {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]LootTable, 0, len(keys))
	for _, k := range keys {
		names := []string{}
		switch t := v[k].(type) {
		case []interface{}:
			for _, item := range t {
				if s, ok := item.(string); ok {
					if strings.TrimSpace(s) != "" {
						names = append(names, s)
					}
					continue
				}
				if n := strings.TrimSpace(asString(asObject(item)["name"])); n != "" {
					names = append(names, n)
				}
			}
		}
		out = append(out, LootTable{ID: k, Count: len(names), Items: names})
	}
	return out
}

func parseItems(v map[string]interface{}) []ItemRow {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]ItemRow, 0, len(keys))
	for _, name := range keys {
		obj := asObject(v[name])
		out = append(out, ItemRow{
			Name:        name,
			ItemID:      asString(obj["itemid"]),
			SpellID:     asString(obj["spellid"]),
			GlobalKey:   asString(obj["globalkey"]),
			IsGlobal:    asString(obj["isglobal"]),
			Description: asString(obj["description"]),
		})
	}
	return out
}

func parseTalents(v interface{}) []TalentRow {
	src := asObject(v)
	keys := make([]string, 0, len(src))
	for k := range src {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]TalentRow, 0, len(keys))
	for _, k := range keys {
		obj := asObject(src[k])
		out = append(out, TalentRow{
			Key:             k,
			ClassFamily:     asString(obj["class_family"]),
			BucketSuffix:    asString(obj["bucket_suffix"]),
			LegacyGlobalKey: asString(obj["legacy_global_key"]),
			ValueType:       asString(obj["value_type"]),
			StateType:       asString(obj["state_type"]),
		})
	}
	return out
}

func parseUnlocks(v interface{}) []UnlockRow {
	src := asObject(v)
	keys := make([]string, 0, len(src))
	for k := range src {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]UnlockRow, 0, len(keys))
	for _, k := range keys {
		obj := asObject(src[k])
		out = append(out, UnlockRow{
			Key:             k,
			SpellID:         asString(obj["spell_id"]),
			SpellName:       asString(obj["spell_name"]),
			BucketSuffix:    asString(obj["bucket_suffix"]),
			ExpectedValue:   asString(obj["expected_value"]),
			LegacyGlobalKey: asString(obj["legacy_global_key"]),
			Notes:           asString(obj["notes"]),
		})
	}
	return out
}

func countCustom(v interface{}) int {
	n := 0
	for name := range asObject(v) {
		if strings.TrimSpace(name) != "" {
			n++
		}
	}
	return n
}

func namesFromAll(v interface{}) []string {
	all, ok := asObject(v)["all"].([]interface{})
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(all))
	for _, item := range all {
		if s, ok := item.(string); ok {
			if strings.TrimSpace(s) != "" {
				out = append(out, s)
			}
			continue
		}
		if n := strings.TrimSpace(asString(asObject(item)["name"])); n != "" {
			out = append(out, n)
		}
	}
	return out
}

func asObject(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func asArray(v interface{}) []interface{} {
	if a, ok := v.([]interface{}); ok && a != nil {
		return a
	}
	return []interface{}{}
}

func asInt(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case json.Number:
		i, _ := t.Int64()
		return int(i)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(t))
		return n
	default:
		n, _ := strconv.Atoi(strings.TrimSpace(fmt.Sprint(t)))
		return n
	}
}

func asString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t {
			return "1"
		}
		return "0"
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}
