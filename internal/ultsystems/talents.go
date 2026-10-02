package ultsystems

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type CatalogMeta struct {
	SchemaVersion  string                 `json:"schemaVersion"`
	CatalogID      string                 `json:"catalogId"`
	GeneratedAtUTC string                 `json:"generatedAtUtc"`
	Policy         map[string]interface{} `json:"policy,omitempty"`
	TaskSelector   map[string]interface{} `json:"taskSelector,omitempty"`
}

type TalentSummary struct {
	TalentID          string   `json:"talentId"`
	DisplayName       string   `json:"displayName"`
	Description       string   `json:"description"`
	ClassIDs          []int    `json:"classIds"`
	TreeID            string   `json:"treeId"`
	PointPoolID       string   `json:"pointPoolId"`
	RankCap           int      `json:"rankCap"`
	CostType          string   `json:"costType"`
	CostAmount        int      `json:"costAmount"`
	SortOrder         int      `json:"sortOrder"`
	ProgressionPhase  string   `json:"progressionPhase"`
	Hidden            bool     `json:"hidden"`
	RequiredTalents   []string `json:"requiredTalents"`
}

type TreeRow struct {
	TreeID    string `json:"treeId"`
	ClassID   int    `json:"classId"`
	Tier      string `json:"tier"`
	Label     string `json:"label"`
	SortOrder int    `json:"sortOrder"`
}

type PoolRow struct {
	PoolID         string `json:"poolId"`
	DisplayName    string `json:"displayName"`
	AvailableKey   string `json:"availableTrrKey"`
	SpentKey       string `json:"spentTrrKey"`
	MaxPoints      int    `json:"maxPoints"`
	MinLevel       int    `json:"minLevel"`
	MaxLevel       int    `json:"maxLevel"`
}

type RankRow struct {
	Key             string `json:"key"`
	ClassFamily     string `json:"classFamily"`
	BucketSuffix    string `json:"bucketSuffix"`
	LegacyGlobalKey string `json:"legacyGlobalKey"`
	ValueType       string `json:"valueType"`
	StateType       string `json:"stateType"`
	RuntimeManaged  bool   `json:"runtimeManaged"`
}

type UnlockRow struct {
	Key             string   `json:"key"`
	SpellID         int      `json:"spellId"`
	SpellName       string   `json:"spellName"`
	BucketSuffix    string   `json:"bucketSuffix"`
	ExpectedValue   string   `json:"expectedValue"`
	LegacyGlobalKey string   `json:"legacyGlobalKey"`
	Notes           string   `json:"notes"`
	Sources         []string `json:"sources"`
}

type SpecRow struct {
	ClassID     int    `json:"classId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DataKey     string `json:"dataKey"`
	Cost        int    `json:"cost"`
	RankCount   int    `json:"rankCount"`
	ScrollID    int    `json:"scrollId"`
	SpellID     int    `json:"spellId"`
}

type CatalogPayload struct {
	Meta    CatalogMeta      `json:"meta"`
	Pools   []PoolRow        `json:"pools"`
	Trees   []TreeRow        `json:"trees"`
	Talents []TalentSummary  `json:"talents"`
	Ranks   []RankRow        `json:"ranks"`
	Unlocks []UnlockRow      `json:"unlocks"`
	Specs   []SpecRow        `json:"specs"`
}

func (s *Service) Catalog() (CatalogPayload, error) {
	out := CatalogPayload{
		Pools:   []PoolRow{},
		Trees:   []TreeRow{},
		Talents: []TalentSummary{},
		Ranks:   []RankRow{},
		Unlocks: []UnlockRow{},
		Specs:   []SpecRow{},
	}
	if cat, err := s.loadMap(talentRel); err == nil {
		out.Meta = CatalogMeta{
			SchemaVersion:  asString(cat["schema_version"]),
			CatalogID:      asString(cat["catalog_id"]),
			GeneratedAtUTC: asString(cat["generated_at_utc"]),
			Policy:         asObject(cat["policy"]),
			TaskSelector:   asObject(cat["task_selector"]),
		}
		for _, raw := range asArray(cat["point_pools"]) {
			row := asObject(raw)
			access := asObject(row["access"])
			out.Pools = append(out.Pools, PoolRow{
				PoolID:       asString(row["pool_id"]),
				DisplayName:  asString(row["display_name"]),
				AvailableKey: asString(row["available_trr_key"]),
				SpentKey:     asString(row["spent_trr_key"]),
				MaxPoints:    asInt(row["max_points"]),
				MinLevel:     asInt(access["min_level"]),
				MaxLevel:     asInt(access["max_level"]),
			})
		}
		for _, raw := range asArray(cat["trees"]) {
			row := asObject(raw)
			out.Trees = append(out.Trees, TreeRow{
				TreeID:    asString(row["tree_id"]),
				ClassID:   asInt(row["class_id"]),
				Tier:      asString(row["tier"]),
				Label:     asString(row["label"]),
				SortOrder: asInt(row["sort_order"]),
			})
		}
		for _, raw := range asArray(cat["talents"]) {
			out.Talents = append(out.Talents, summarizeTalent(asObject(raw)))
		}
	} else if s.questsDirPath() == "" {
		return out, fmt.Errorf("quests directory is not set")
	}
	if ranks, err := s.loadMap(rankRel); err == nil {
		obj := asObject(ranks["talent_ranks"])
		for _, key := range sortedKeys(obj) {
			row := asObject(obj[key])
			out.Ranks = append(out.Ranks, RankRow{
				Key:             key,
				ClassFamily:     asString(row["class_family"]),
				BucketSuffix:    asString(row["bucket_suffix"]),
				LegacyGlobalKey: asString(row["legacy_global_key"]),
				ValueType:       asString(row["value_type"]),
				StateType:       asString(row["state_type"]),
				RuntimeManaged:  asString(row["runtime_managed"]) == "true" || row["runtime_managed"] == true,
			})
		}
	}
	if unlocks, err := s.loadMap(unlockRel); err == nil {
		obj := asObject(unlocks["unlocks"])
		for _, key := range sortedKeys(obj) {
			row := asObject(obj[key])
			sources := []string{}
			for _, item := range asArray(row["sources"]) {
				if txt := strings.TrimSpace(asString(item)); txt != "" {
					sources = append(sources, txt)
				}
			}
			out.Unlocks = append(out.Unlocks, UnlockRow{
				Key:             key,
				SpellID:         asInt(row["spell_id"]),
				SpellName:       asString(row["spell_name"]),
				BucketSuffix:    asString(row["bucket_suffix"]),
				ExpectedValue:   asString(row["expected_value"]),
				LegacyGlobalKey: asString(row["legacy_global_key"]),
				Notes:           asString(row["notes"]),
				Sources:         sources,
			})
		}
	}
	if specs, err := s.loadMap(specRel); err == nil {
		for _, classKey := range sortedKeys(specs) {
			classID := asInt(classKey)
			if classID <= 0 {
				continue
			}
			for _, name := range sortedKeys(asObject(specs[classKey])) {
				row := asObject(asObject(specs[classKey])[name])
				out.Specs = append(out.Specs, SpecRow{
					ClassID:     classID,
					Name:        name,
					Description: asString(row["description"]),
					DataKey:     asString(row["datakey"]),
					Cost:        asInt(row["cost"]),
					RankCount:   asInt(row["rankcount"]),
					ScrollID:    asInt(row["scrollid"]),
					SpellID:     asInt(row["spellid"]),
				})
			}
		}
	}
	sort.Slice(out.Talents, func(i, j int) bool {
		if firstClass(out.Talents[i].ClassIDs) != firstClass(out.Talents[j].ClassIDs) {
			return firstClass(out.Talents[i].ClassIDs) < firstClass(out.Talents[j].ClassIDs)
		}
		return out.Talents[i].DisplayName < out.Talents[j].DisplayName
	})
	return out, nil
}

func firstClass(ids []int) int {
	if len(ids) == 0 {
		return 0
	}
	return ids[0]
}

func summarizeTalent(row map[string]interface{}) TalentSummary {
	cost := asObject(row["cost_model"])
	ui := asObject(row["ui"])
	prereq := asObject(row["prerequisites"])
	required := []string{}
	for _, item := range asArray(prereq["required_talents"]) {
		id := asString(asObject(item)["talent_id"])
		if id != "" {
			required = append(required, id)
		}
	}
	return TalentSummary{
		TalentID:         asString(row["talent_id"]),
		DisplayName:      asString(row["display_name"]),
		Description:      asString(row["description"]),
		ClassIDs:         intSlice(row["class_ids"]),
		TreeID:           asString(row["tree_id"]),
		PointPoolID:      asString(row["point_pool_id"]),
		RankCap:          asInt(row["rank_cap"]),
		CostType:         asString(cost["type"]),
		CostAmount:       asInt(cost["amount"]),
		SortOrder:        asInt(row["sort_order"]),
		ProgressionPhase: asString(row["progression_phase"]),
		Hidden:           ui["hidden"] == true,
		RequiredTalents:  required,
	}
}

func (s *Service) Talent(id string) (map[string]interface{}, error) {
	id = strings.TrimSpace(id)
	cat, err := s.loadMap(talentRel)
	if err != nil {
		return nil, err
	}
	for _, raw := range asArray(cat["talents"]) {
		row := asObject(raw)
		if asString(row["talent_id"]) == id {
			return row, nil
		}
	}
	return nil, fmt.Errorf("talent %s not found", id)
}

func (s *Service) writeTalent(plan *WritePlan, req WriteRequest) error {
	id := strings.TrimSpace(textOf(req.Entry, "talent_id", "talentId", "id"))
	if id == "" {
		return fmt.Errorf("talent_id is required")
	}
	if !slugOK(id) {
		return fmt.Errorf("talent_id must be lowercase letters, numbers, or underscore")
	}
	plan.Key = id
	cat, err := s.loadMap(talentRel)
	if err != nil {
		return err
	}
	talents := asArray(cat["talents"])
	idx := -1
	var current map[string]interface{}
	for i, raw := range talents {
		row := asObject(raw)
		if asString(row["talent_id"]) == id {
			idx = i
			current = row
			break
		}
	}
	if req.Action == "delete" {
		if idx < 0 {
			return fmt.Errorf("talent %s not found", id)
		}
		cat["talents"] = append(talents[:idx], talents[idx+1:]...)
		return s.commitJSON(plan, talentRel, cat)
	}
	if current == nil {
		current = defaultTalent(id)
	}
	applyTalentFields(current, req.Entry)
	if idx >= 0 {
		talents[idx] = current
	} else {
		talents = append(talents, current)
	}
	cat["talents"] = talents
	cat["generated_at_utc"] = time.Now().UTC().Format(time.RFC3339)
	return s.commitJSON(plan, talentRel, cat)
}

func defaultTalent(id string) map[string]interface{} {
	return map[string]interface{}{
		"talent_id":     id,
		"trr_key":       id,
		"class_ids":     []interface{}{},
		"tree_id":       "",
		"point_pool_id": "demigod_talent",
		"display_name":  id,
		"description":   "New talent",
		"sort_order":    10,
		"rank_cap":      1,
		"cost_model": map[string]interface{}{
			"type":   "flat",
			"amount": 1,
		},
		"prerequisites": map[string]interface{}{
			"min_level":            1,
			"points_spent_at_least": 0,
			"required_talents":     []interface{}{},
			"required_unlocks_all": []interface{}{},
		},
		"ui": map[string]interface{}{
			"say_token":   id,
			"task_offset": 0,
			"hidden":      false,
		},
		"rank_effects":       []interface{}{},
		"progression_phase":  "demigod",
	}
}

func applyTalentFields(dst map[string]interface{}, src map[string]interface{}) {
	if v := textOf(src, "display_name", "displayName"); v != "" {
		dst["display_name"] = v
	}
	if v := textOf(src, "description"); v != "" {
		dst["description"] = v
	}
	if v := textOf(src, "tree_id", "treeId"); v != "" {
		dst["tree_id"] = v
	}
	if v := textOf(src, "point_pool_id", "pointPoolId"); v != "" {
		dst["point_pool_id"] = v
	}
	if v := textOf(src, "trr_key", "trrKey"); v != "" {
		dst["trr_key"] = v
	}
	if v := textOf(src, "progression_phase", "progressionPhase"); v != "" {
		dst["progression_phase"] = v
	}
	if src["class_ids"] != nil || src["classIds"] != nil {
		ids := intSlice(src["class_ids"])
		if len(ids) == 0 {
			ids = intSlice(src["classIds"])
		}
		arr := make([]interface{}, 0, len(ids))
		for _, n := range ids {
			arr = append(arr, n)
		}
		dst["class_ids"] = arr
	}
	if src["rank_cap"] != nil || src["rankCap"] != nil {
		n := asInt(src["rank_cap"])
		if n == 0 {
			n = asInt(src["rankCap"])
		}
		if n > 0 {
			dst["rank_cap"] = n
		}
	}
	if src["sort_order"] != nil || src["sortOrder"] != nil {
		n := asInt(src["sort_order"])
		if src["sortOrder"] != nil && asInt(src["sort_order"]) == 0 {
			n = asInt(src["sortOrder"])
		}
		dst["sort_order"] = n
	}
	cost := asObject(dst["cost_model"])
	if v := textOf(src, "cost_type", "costType"); v != "" {
		cost["type"] = v
	}
	if src["cost_amount"] != nil || src["costAmount"] != nil || src["amount"] != nil {
		n := asInt(src["cost_amount"])
		if n == 0 {
			n = asInt(src["costAmount"])
		}
		if n == 0 {
			n = asInt(src["amount"])
		}
		if n > 0 {
			cost["amount"] = n
		}
	}
	if cost["type"] == "" {
		cost["type"] = "flat"
	}
	dst["cost_model"] = cost
	prereq := asObject(dst["prerequisites"])
	if src["min_level"] != nil || src["minLevel"] != nil {
		n := asInt(src["min_level"])
		if n == 0 {
			n = asInt(src["minLevel"])
		}
		if n > 0 {
			prereq["min_level"] = n
		}
	}
	if src["points_spent_at_least"] != nil || src["pointsSpentAtLeast"] != nil {
		n := asInt(src["points_spent_at_least"])
		if src["pointsSpentAtLeast"] != nil {
			n = asInt(src["pointsSpentAtLeast"])
		}
		prereq["points_spent_at_least"] = n
	}
	if src["required_talents"] != nil || src["requiredTalents"] != nil {
		raw := src["required_talents"]
		if raw == nil {
			raw = src["requiredTalents"]
		}
		prereq["required_talents"] = normalizeRequiredTalents(raw)
	}
	dst["prerequisites"] = prereq
	ui := asObject(dst["ui"])
	if v := textOf(src, "say_token", "sayToken"); v != "" {
		ui["say_token"] = v
	}
	if src["task_offset"] != nil || src["taskOffset"] != nil {
		n := asInt(src["task_offset"])
		if src["taskOffset"] != nil {
			n = asInt(src["taskOffset"])
		}
		ui["task_offset"] = n
	}
	if src["hidden"] != nil {
		ui["hidden"] = src["hidden"] == true || asString(src["hidden"]) == "true"
	}
	dst["ui"] = ui
	if src["rank_effects"] != nil {
		dst["rank_effects"] = src["rank_effects"]
	}
}

func normalizeRequiredTalents(raw interface{}) []interface{} {
	out := []interface{}{}
	switch t := raw.(type) {
	case []interface{}:
		for _, item := range t {
			if s, ok := item.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					out = append(out, map[string]interface{}{"talent_id": s, "min_rank": 1})
				}
				continue
			}
			row := asObject(item)
			id := textOf(row, "talent_id", "talentId")
			if id == "" {
				continue
			}
			min := asInt(row["min_rank"])
			if min <= 0 {
				min = asInt(row["minRank"])
			}
			if min <= 0 {
				min = 1
			}
			out = append(out, map[string]interface{}{"talent_id": id, "min_rank": min})
		}
	case string:
		for _, part := range strings.Split(t, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, map[string]interface{}{"talent_id": part, "min_rank": 1})
			}
		}
	}
	return out
}

func (s *Service) writeTree(plan *WritePlan, req WriteRequest) error {
	id := strings.TrimSpace(textOf(req.Entry, "tree_id", "treeId", "id"))
	if id == "" {
		return fmt.Errorf("tree_id is required")
	}
	if !slugOK(id) {
		return fmt.Errorf("tree_id must be lowercase letters, numbers, or underscore")
	}
	plan.Key = id
	cat, err := s.loadMap(talentRel)
	if err != nil {
		return err
	}
	trees := asArray(cat["trees"])
	idx := -1
	var current map[string]interface{}
	for i, raw := range trees {
		row := asObject(raw)
		if asString(row["tree_id"]) == id {
			idx = i
			current = row
			break
		}
	}
	if req.Action == "delete" {
		if idx < 0 {
			return fmt.Errorf("tree %s not found", id)
		}
		cat["trees"] = append(trees[:idx], trees[idx+1:]...)
		return s.commitJSON(plan, talentRel, cat)
	}
	if current == nil {
		current = map[string]interface{}{
			"tree_id":    id,
			"class_id":   1,
			"tier":       "demigod",
			"label":      id,
			"sort_order": 10,
		}
	}
	if v := textOf(req.Entry, "label"); v != "" {
		current["label"] = v
	}
	if v := textOf(req.Entry, "tier"); v != "" {
		current["tier"] = v
	}
	if req.Entry["class_id"] != nil || req.Entry["classId"] != nil {
		n := asInt(req.Entry["class_id"])
		if n == 0 {
			n = asInt(req.Entry["classId"])
		}
		if n > 0 {
			current["class_id"] = n
		}
	}
	if req.Entry["sort_order"] != nil || req.Entry["sortOrder"] != nil {
		n := asInt(req.Entry["sort_order"])
		if req.Entry["sortOrder"] != nil {
			n = asInt(req.Entry["sortOrder"])
		}
		current["sort_order"] = n
	}
	if idx >= 0 {
		trees[idx] = current
	} else {
		trees = append(trees, current)
	}
	cat["trees"] = trees
	return s.commitJSON(plan, talentRel, cat)
}

func (s *Service) writePool(plan *WritePlan, req WriteRequest) error {
	id := strings.TrimSpace(textOf(req.Entry, "pool_id", "poolId", "id"))
	if id == "" {
		return fmt.Errorf("pool_id is required")
	}
	if !slugOK(id) {
		return fmt.Errorf("pool_id must be lowercase letters, numbers, or underscore")
	}
	plan.Key = id
	cat, err := s.loadMap(talentRel)
	if err != nil {
		return err
	}
	pools := asArray(cat["point_pools"])
	idx := -1
	var current map[string]interface{}
	for i, raw := range pools {
		row := asObject(raw)
		if asString(row["pool_id"]) == id {
			idx = i
			current = row
			break
		}
	}
	if req.Action == "delete" {
		if idx < 0 {
			return fmt.Errorf("pool %s not found", id)
		}
		cat["point_pools"] = append(pools[:idx], pools[idx+1:]...)
		return s.commitJSON(plan, talentRel, cat)
	}
	if current == nil {
		current = map[string]interface{}{
			"pool_id":           id,
			"display_name":      id,
			"available_trr_key": id + "_available",
			"spent_trr_key":     id + "_spent",
			"max_points":        80,
			"currency_ids":      []interface{}{},
			"access": map[string]interface{}{
				"min_level":          1,
				"max_level":          85,
				"required_tss_flags": []interface{}{},
			},
		}
	}
	if v := textOf(req.Entry, "display_name", "displayName"); v != "" {
		current["display_name"] = v
	}
	if req.Entry["max_points"] != nil || req.Entry["maxPoints"] != nil {
		n := asInt(req.Entry["max_points"])
		if n == 0 {
			n = asInt(req.Entry["maxPoints"])
		}
		if n > 0 {
			current["max_points"] = n
		}
	}
	if idx >= 0 {
		pools[idx] = current
	} else {
		pools = append(pools, current)
	}
	cat["point_pools"] = pools
	return s.commitJSON(plan, talentRel, cat)
}

func (s *Service) writeRank(plan *WritePlan, req WriteRequest) error {
	key := strings.TrimSpace(textOf(req.Entry, "key", "id"))
	if key == "" {
		return fmt.Errorf("rank key is required")
	}
	if !slugOK(key) {
		return fmt.Errorf("rank key must be lowercase letters, numbers, or underscore")
	}
	plan.Key = key
	doc, err := s.loadMap(rankRel)
	if err != nil {
		return err
	}
	ranks := asObject(doc["talent_ranks"])
	if req.Action == "delete" {
		if _, ok := ranks[key]; !ok {
			return fmt.Errorf("rank %s not found", key)
		}
		delete(ranks, key)
		doc["talent_ranks"] = ranks
		return s.commitJSON(plan, rankRel, doc)
	}
	current := asObject(ranks[key])
	if len(current) == 0 {
		fam := strings.TrimSpace(textOf(req.Entry, "class_family", "classFamily"))
		if fam == "" {
			if i := strings.Index(key, "_"); i > 0 {
				fam = key[:i]
			}
		}
		current = map[string]interface{}{
			"state_type":        "talent_rank",
			"runtime_managed":   true,
			"bucket_suffix":     "_TalentRank_" + key,
			"legacy_global_key": key,
			"class_family":      fam,
			"value_type":        "int",
		}
	}
	if v := textOf(req.Entry, "class_family", "classFamily"); v != "" {
		current["class_family"] = v
	}
	if v := textOf(req.Entry, "bucket_suffix", "bucketSuffix"); v != "" {
		current["bucket_suffix"] = v
	}
	if v := textOf(req.Entry, "legacy_global_key", "legacyGlobalKey"); v != "" {
		current["legacy_global_key"] = v
	}
	if v := textOf(req.Entry, "value_type", "valueType"); v != "" {
		current["value_type"] = v
	}
	if v := textOf(req.Entry, "state_type", "stateType"); v != "" {
		current["state_type"] = v
	}
	ranks[key] = current
	doc["talent_ranks"] = ranks
	return s.commitJSON(plan, rankRel, doc)
}

func (s *Service) writeUnlock(plan *WritePlan, req WriteRequest) error {
	key := strings.TrimSpace(textOf(req.Entry, "key", "id"))
	if key == "" {
		return fmt.Errorf("unlock key is required")
	}
	plan.Key = key
	doc, err := s.loadMap(unlockRel)
	if err != nil {
		return err
	}
	unlocks := asObject(doc["unlocks"])
	if req.Action == "delete" {
		if _, ok := unlocks[key]; !ok {
			return fmt.Errorf("unlock %s not found", key)
		}
		delete(unlocks, key)
		doc["unlocks"] = unlocks
		return s.commitJSON(plan, unlockRel, doc)
	}
	current := asObject(unlocks[key])
	if len(current) == 0 {
		current = map[string]interface{}{
			"state_type":        "spell_unlock",
			"runtime_managed":   true,
			"bucket_suffix":     "_Unlock_" + key,
			"spell_id":          0,
			"spell_name":        "",
			"expected_value":    "1",
			"legacy_global_key": key,
			"sources":           []interface{}{"spire"},
			"notes":             nil,
		}
	}
	if req.Entry["spell_id"] != nil || req.Entry["spellId"] != nil {
		n := asInt(req.Entry["spell_id"])
		if n == 0 {
			n = asInt(req.Entry["spellId"])
		}
		current["spell_id"] = n
	}
	if v := textOf(req.Entry, "spell_name", "spellName"); v != "" {
		current["spell_name"] = v
	}
	if v := textOf(req.Entry, "expected_value", "expectedValue"); v != "" {
		current["expected_value"] = v
	}
	if v := textOf(req.Entry, "bucket_suffix", "bucketSuffix"); v != "" {
		current["bucket_suffix"] = v
	}
	if v := textOf(req.Entry, "legacy_global_key", "legacyGlobalKey"); v != "" {
		current["legacy_global_key"] = v
	}
	if req.Entry["notes"] != nil {
		current["notes"] = req.Entry["notes"]
	}
	unlocks[key] = current
	doc["unlocks"] = unlocks
	return s.commitJSON(plan, unlockRel, doc)
}

func (s *Service) writeSpec(plan *WritePlan, req WriteRequest) error {
	name := strings.TrimSpace(textOf(req.Entry, "name"))
	classID := asInt(req.Entry["class_id"])
	if classID == 0 {
		classID = asInt(req.Entry["classId"])
	}
	if name == "" || classID < 1 || classID > 16 {
		return fmt.Errorf("class_id (1-16) and name are required")
	}
	plan.Key = fmt.Sprintf("%d/%s", classID, name)
	doc, err := s.loadMap(specRel)
	if err != nil {
		return err
	}
	classKey := fmt.Sprintf("%d", classID)
	classNode := asObject(doc[classKey])
	if req.Action == "delete" {
		if _, ok := classNode[name]; !ok {
			return fmt.Errorf("specialization %s not found", plan.Key)
		}
		delete(classNode, name)
		doc[classKey] = classNode
		return s.commitJSON(plan, specRel, doc)
	}
	current := asObject(classNode[name])
	if len(current) == 0 {
		current = map[string]interface{}{
			"datakey":     "_Has_" + strings.ReplaceAll(name, " ", "_"),
			"description": name,
			"scrollid":    -1,
			"spellid":     -1,
			"spelltype":   -1,
			"cost":        1,
			"rankcount":   1,
			"ranks":       []interface{}{map[string]interface{}{"id": 1, "name": name + " I"}},
		}
	}
	if v := textOf(req.Entry, "description"); v != "" {
		current["description"] = v
	}
	if v := textOf(req.Entry, "datakey", "dataKey"); v != "" {
		current["datakey"] = v
	}
	if req.Entry["cost"] != nil {
		current["cost"] = asInt(req.Entry["cost"])
	}
	if req.Entry["rankcount"] != nil || req.Entry["rankCount"] != nil {
		n := asInt(req.Entry["rankcount"])
		if n == 0 {
			n = asInt(req.Entry["rankCount"])
		}
		if n > 0 {
			current["rankcount"] = n
			ranks := make([]interface{}, 0, n)
			for i := 1; i <= n; i++ {
				ranks = append(ranks, map[string]interface{}{
					"id":   i,
					"name": fmt.Sprintf("%s %s", name, roman(i)),
				})
			}
			current["ranks"] = ranks
		}
	}
	if req.Entry["scrollid"] != nil || req.Entry["scrollId"] != nil {
		n := asInt(req.Entry["scrollid"])
		if req.Entry["scrollId"] != nil {
			n = asInt(req.Entry["scrollId"])
		}
		current["scrollid"] = n
	}
	if req.Entry["spellid"] != nil || req.Entry["spellId"] != nil {
		n := asInt(req.Entry["spellid"])
		if req.Entry["spellId"] != nil {
			n = asInt(req.Entry["spellId"])
		}
		current["spellid"] = n
	}
	classNode[name] = current
	doc[classKey] = classNode
	return s.commitJSON(plan, specRel, doc)
}

func roman(n int) string {
	vals := []int{10, 9, 5, 4, 1}
	syms := []string{"X", "IX", "V", "IV", "I"}
	out := ""
	for i, v := range vals {
		for n >= v {
			out += syms[i]
			n -= v
		}
	}
	if out == "" {
		return "I"
	}
	return out
}
