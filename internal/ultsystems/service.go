package ultsystems

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/EQEmu/spire/internal/database"
	"github.com/EQEmu/spire/internal/pathmgmt"
)

const (
	maxJSONBytes     = 8 << 20
	talentRel        = "global/talentdata/talent_catalog.json"
	talentsLegacyRel = "global/talentdata/talents.json"
	specRel          = "global/talentdata/specializations.json"
	rankRel          = "global/ultimatedata/talent_rank_registry.json"
	unlockRel        = "global/ultimatedata/unlock_state_registry.json"
	traitRel         = "global/vendordata/trait_items_non_god_catalog.json"
	runewordRel      = "global/ultimatedata/runeword_catalog.json"
	zcScriptRel      = "global/zone_controller.pl"
)

type Service struct {
	pathmgmt  *pathmgmt.PathManagement
	questsDir string
	db        *database.Resolver
}

func NewService(pathmgmt *pathmgmt.PathManagement, db *database.Resolver) *Service {
	return &Service{pathmgmt: pathmgmt, db: db}
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

func (s *Service) abs(rel string) string {
	return filepath.Join(s.questsDirPath(), filepath.FromSlash(rel))
}

type FileInfo struct {
	Kind    string `json:"kind"`
	RelPath string `json:"relPath"`
	Exists  bool   `json:"exists"`
	Size    int64  `json:"size"`
}

type Status struct {
	OK            bool                `json:"ok"`
	Error         string              `json:"error,omitempty"`
	QuestsDir     string              `json:"questsDir"`
	TalentCount   int                 `json:"talentCount"`
	TreeCount     int                 `json:"treeCount"`
	PoolCount     int                 `json:"poolCount"`
	RankCount     int                 `json:"rankCount"`
	UnlockCount   int                 `json:"unlockCount"`
	SpecCount     int                 `json:"specCount"`
	TraitCount    int                 `json:"traitCount"`
	RunewordItems int                 `json:"runewordItems"`
	RunewordCombos int                `json:"runewordCombos"`
	ZoneJSONCount int                 `json:"zoneJsonCount"`
	Files         map[string]FileInfo `json:"files"`
	RunewordTables RunewordTables     `json:"runewordTables"`
}

type RunewordTables struct {
	Available bool           `json:"available"`
	Error     string         `json:"error,omitempty"`
	Counts    map[string]int `json:"counts,omitempty"`
}

type WritePlan struct {
	OK       bool     `json:"ok"`
	DryRun   bool     `json:"dryRun"`
	Kind     string   `json:"kind"`
	Action   string   `json:"action"`
	RelPath  string   `json:"relPath"`
	Key      string   `json:"key,omitempty"`
	Backups  []string `json:"backups,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Error    string   `json:"error,omitempty"`
}

type WriteRequest struct {
	Kind   string                 `json:"kind"`
	Action string                 `json:"action"`
	DryRun *bool                  `json:"dryRun"`
	Entry  map[string]interface{} `json:"entry"`
}

func (s *Service) Status() Status {
	st := Status{
		QuestsDir: s.questsDirPath(),
		Files:     map[string]FileInfo{},
	}
	if st.QuestsDir == "" {
		st.Error = "quests directory is not set; run Spire from the EQ server folder or set SPIRE_QUESTS_ROOT"
		return st
	}
	for kind, rel := range map[string]string{
		"catalog": talentRel,
		"legacy":  talentsLegacyRel,
		"specs":   specRel,
		"ranks":   rankRel,
		"unlocks": unlockRel,
		"traits":  traitRel,
		"runewords": runewordRel,
		"zone_controller": zcScriptRel,
	} {
		info := FileInfo{Kind: kind, RelPath: rel}
		if fi, err := os.Stat(s.abs(rel)); err == nil {
			info.Exists = true
			info.Size = fi.Size()
		}
		st.Files[kind] = info
	}
	if cat, err := s.loadMap(talentRel); err == nil {
		st.TalentCount = len(asArray(cat["talents"]))
		st.TreeCount = len(asArray(cat["trees"]))
		st.PoolCount = len(asArray(cat["point_pools"]))
	}
	if ranks, err := s.loadMap(rankRel); err == nil {
		st.RankCount = len(asObject(ranks["talent_ranks"]))
	}
	if unlocks, err := s.loadMap(unlockRel); err == nil {
		st.UnlockCount = len(asObject(unlocks["unlocks"]))
	}
	if specs, err := s.loadMap(specRel); err == nil {
		for _, classNode := range specs {
			st.SpecCount += len(asObject(classNode))
		}
	}
	if traits, err := s.loadMap(traitRel); err == nil {
		st.TraitCount = len(asArray(traits["items"]))
	}
	rw := s.runewordCatalog()
	st.RunewordItems = len(rw.Items)
	st.RunewordCombos = len(rw.Combos)
	st.ZoneJSONCount = s.countZoneDirs()
	st.RunewordTables = s.probeRunewordTables()
	st.OK = st.Files["catalog"].Exists || st.Files["ranks"].Exists
	if !st.OK {
		st.Error = "ultimate talent/runeword JSON not found under " + st.QuestsDir
	}
	return st
}

func (s *Service) countZoneDirs() int {
	dir := filepath.Join(s.questsDirPath(), filepath.FromSlash("global/ultimatedata"))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() && isDigits(e.Name()) {
			n++
		}
	}
	return n
}

func (s *Service) Write(req WriteRequest) WritePlan {
	plan := WritePlan{
		DryRun:  dryRunValue(req.DryRun),
		Kind:    strings.TrimSpace(req.Kind),
		Action:  strings.ToLower(strings.TrimSpace(req.Action)),
	}
	if plan.Action == "" {
		plan.Action = "upsert"
	}
	if plan.Action != "upsert" && plan.Action != "delete" && plan.Action != "seed" {
		plan.Error = "action must be upsert, delete, or seed"
		return plan
	}
	if req.Entry == nil {
		req.Entry = map[string]interface{}{}
	}

	var err error
	switch plan.Kind {
	case "talent":
		err = s.writeTalent(&plan, req)
	case "tree":
		err = s.writeTree(&plan, req)
	case "pool":
		err = s.writePool(&plan, req)
	case "rank":
		err = s.writeRank(&plan, req)
	case "unlock":
		err = s.writeUnlock(&plan, req)
	case "spec":
		err = s.writeSpec(&plan, req)
	case "trait":
		err = s.writeTrait(&plan, req)
	case "runeword_item":
		err = s.writeRunewordItem(&plan, req)
	case "runeword_combo":
		err = s.writeRunewordCombo(&plan, req)
	case "runeword_seed":
		err = s.seedRunewords(&plan, req)
	default:
		err = fmt.Errorf("unknown kind %q", plan.Kind)
	}
	if err != nil {
		plan.Error = err.Error()
		plan.OK = false
		return plan
	}
	plan.OK = true
	return plan
}

func (s *Service) loadMap(rel string) (map[string]interface{}, error) {
	path := s.abs(rel)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) > maxJSONBytes {
		return nil, fmt.Errorf("%s is too large", rel)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", rel, err)
	}
	if out == nil {
		out = map[string]interface{}{}
	}
	return out, nil
}

func (s *Service) commitJSON(plan *WritePlan, rel string, payload interface{}) error {
	plan.RelPath = rel
	path := s.abs(rel)
	if plan.DryRun {
		if _, err := os.Stat(path); err == nil {
			plan.Action = "would-update"
		} else {
			plan.Action = "would-create"
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		backup, err := s.backupFile(rel)
		if err != nil {
			return err
		}
		if backup != "" {
			plan.Backups = append(plan.Backups, backup)
		}
		plan.Action = "updated"
	} else {
		plan.Action = "created"
	}
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return os.WriteFile(path, body, 0o644)
}

func (s *Service) backupFile(rel string) (string, error) {
	src := s.abs(rel)
	stamp := time.Now().Format("20060102-150405")
	destRel := filepath.ToSlash(filepath.Join("global/_spire_backups", stamp, filepath.Base(rel)))
	dest := s.abs(destRel)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	raw, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(dest, raw, 0o644); err != nil {
		return "", err
	}
	return destRel, nil
}

func dryRunValue(flag *bool) bool {
	if flag == nil {
		return true
	}
	return *flag
}

func asObject(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok && m != nil {
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

func asString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	case json.Number:
		return t.String()
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", t))
	}
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
		var n int
		fmt.Sscanf(strings.TrimSpace(t), "%d", &n)
		return n
	default:
		return 0
	}
}

func textOf(entry map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if s := strings.TrimSpace(asString(entry[k])); s != "" {
			return s
		}
	}
	return ""
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func slugOK(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func mergeEntry(dst map[string]interface{}, src map[string]interface{}) {
	for k, v := range src {
		if v == nil {
			continue
		}
		dst[k] = v
	}
}

func intSlice(v interface{}) []int {
	out := []int{}
	switch t := v.(type) {
	case []interface{}:
		for _, item := range t {
			if n := asInt(item); n > 0 {
				out = append(out, n)
			}
		}
	case []int:
		for _, n := range t {
			if n > 0 {
				out = append(out, n)
			}
		}
	}
	return out
}
