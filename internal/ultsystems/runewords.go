package ultsystems

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type RunewordItem struct {
	SourceKey    string `json:"sourceKey"`
	ItemKind     string `json:"itemKind"`
	TargetItemID int    `json:"targetItemId"`
	Name         string `json:"name"`
}

type RunewordCombo struct {
	ComboKey      string `json:"comboKey"`
	FamilyKey     string `json:"familyKey"`
	QualityTier   string `json:"qualityTier"`
	OutputItemID  int    `json:"outputItemId"`
	OutputName    string `json:"outputName"`
	BaseItemID    int    `json:"baseItemId"`
	Rune1         int    `json:"rune1"`
	Rune2         int    `json:"rune2"`
	Rune3         int    `json:"rune3"`
	Rune4         int    `json:"rune4"`
	EffectProfile string `json:"effectProfile"`
}

type RunewordFamily struct {
	FamilyKey string `json:"familyKey"`
	Label     string `json:"label"`
	Domain    string `json:"domain"`
}

type RunewordCatalog struct {
	SchemaVersion string           `json:"schemaVersion"`
	Seeded        bool             `json:"seeded"`
	Exists        bool             `json:"exists"`
	Items         []RunewordItem   `json:"items"`
	Combos        []RunewordCombo  `json:"combos"`
	Families      []RunewordFamily `json:"families"`
}

func (s *Service) Runewords() RunewordCatalog {
	return s.runewordCatalog()
}

func (s *Service) runewordCatalog() RunewordCatalog {
	if doc, err := s.loadMap(runewordRel); err == nil {
		return parseRunewordDoc(doc, true, false)
	}
	return defaultRunewordCatalog()
}

func parseRunewordDoc(doc map[string]interface{}, exists, seeded bool) RunewordCatalog {
	out := RunewordCatalog{
		SchemaVersion: asString(doc["schema_version"]),
		Exists:        exists,
		Seeded:        seeded,
		Items:         []RunewordItem{},
		Combos:        []RunewordCombo{},
		Families:      []RunewordFamily{},
	}
	if out.SchemaVersion == "" {
		out.SchemaVersion = "1.0.0"
	}
	for _, raw := range asArray(doc["items"]) {
		row := asObject(raw)
		out.Items = append(out.Items, RunewordItem{
			SourceKey:    asString(row["source_key"]),
			ItemKind:     asString(row["item_kind"]),
			TargetItemID: asInt(row["target_item_id"]),
			Name:         asString(row["name"]),
		})
	}
	for _, raw := range asArray(doc["combos"]) {
		row := asObject(raw)
		runes := intSlice(row["runes"])
		for len(runes) < 4 {
			runes = append(runes, 0)
		}
		out.Combos = append(out.Combos, RunewordCombo{
			ComboKey:      asString(row["combo_key"]),
			FamilyKey:     asString(row["family_key"]),
			QualityTier:   asString(row["quality_tier"]),
			OutputItemID:  asInt(row["output_item_id"]),
			OutputName:    asString(row["output_name"]),
			BaseItemID:    asInt(row["base_item_id"]),
			Rune1:         pickRune(row, runes, 0, "rune_1_item_id"),
			Rune2:         pickRune(row, runes, 1, "rune_2_item_id"),
			Rune3:         pickRune(row, runes, 2, "rune_3_item_id"),
			Rune4:         pickRune(row, runes, 3, "rune_4_item_id"),
			EffectProfile: asString(row["effect_profile"]),
		})
	}
	for _, raw := range asArray(doc["families"]) {
		row := asObject(raw)
		out.Families = append(out.Families, RunewordFamily{
			FamilyKey: asString(row["family_key"]),
			Label:     asString(row["label"]),
			Domain:    asString(row["domain"]),
		})
	}
	return out
}

func pickRune(row map[string]interface{}, runes []int, idx int, key string) int {
	if n := asInt(row[key]); n > 0 {
		return n
	}
	if idx < len(runes) {
		return runes[idx]
	}
	return 0
}

func defaultRunewordCatalog() RunewordCatalog {
	doc := defaultRunewordDoc()
	return parseRunewordDoc(doc, false, true)
}

func defaultRunewordDoc() map[string]interface{} {
	return map[string]interface{}{
		"schema_version": "1.0.0",
		"generated_at_utc": time.Now().UTC().Format(time.RFC3339),
		"source": "spire_seed_from_ueq_runeword_inventory",
		"items": []interface{}{
			itemDoc("runeword.rune.el", "rune", 899300, "El Rune"),
			itemDoc("runeword.rune.ort", "rune", 899301, "Ort Rune"),
			itemDoc("runeword.rune.thul", "rune", 899302, "Thul Rune"),
			itemDoc("runeword.rune.um", "rune", 899303, "Um Rune"),
			itemDoc("runeword.rune.vex", "rune", 899304, "Vex Rune"),
			itemDoc("runeword.rune.ber", "rune", 899305, "Ber Rune"),
			itemDoc("runeword.rune.jah", "rune", 899306, "Jah Rune"),
			itemDoc("runeword.rune.zod", "rune", 899307, "Zod Rune"),
			itemDoc("runeword.base.apostle_mace.rusty", "base", 899320, "Rusty Apostle Mace (Ethereal)"),
			itemDoc("runeword.base.apostle_mace", "base", 899321, "Apostle Mace (Ethereal)"),
			itemDoc("runeword.base.apostle_mace.ultimate", "base", 899322, "Ultimate Apostle Mace (Ethereal)"),
			itemDoc("runeword.base.enforcer_sword.rusty", "base", 899323, "Rusty Enforcer Sword (Ethereal)"),
			itemDoc("runeword.base.enforcer_sword", "base", 899324, "Enforcer Sword (Ethereal)"),
			itemDoc("runeword.base.enforcer_sword.ultimate", "base", 899325, "Ultimate Enforcer Sword (Ethereal)"),
			itemDoc("runeword.base.enforcer_zweihander.rusty", "base", 899326, "Rusty Enforcer Zweihander (Ethereal)"),
			itemDoc("runeword.base.enforcer_zweihander", "base", 899327, "Enforcer Zweihander (Ethereal)"),
			itemDoc("runeword.base.enforcer_zweihander.ult", "base", 899328, "Ultimate Enforcer Zweihander (Ethereal)"),
			itemDoc("runeword.base.scout_dagger.rusty", "base", 899329, "Rusty Scout Dagger (Ethereal)"),
			itemDoc("runeword.base.scout_dagger", "base", 899330, "Scout Dagger (Ethereal)"),
			itemDoc("runeword.base.scout_dagger.ultimate", "base", 899331, "Ultimate Scout Dagger (Ethereal)"),
			itemDoc("runeword.base.disciple_fist.rusty", "base", 899332, "Rusty Disciple Fist (Ethereal)"),
			itemDoc("runeword.base.disciple_fist", "base", 899333, "Disciple Fist (Ethereal)"),
			itemDoc("runeword.base.disciple_fist.ultimate", "base", 899334, "Ultimate Disciple Fist (Ethereal)"),
			itemDoc("runeword.base.warden_bow.rusty", "base", 899335, "Rusty Warden Bow (Ethereal)"),
			itemDoc("runeword.base.warden_bow", "base", 899336, "Warden Bow (Ethereal)"),
			itemDoc("runeword.base.warden_bow.ultimate", "base", 899337, "Ultimate Warden Bow (Ethereal)"),
			itemDoc("runeword.base.channeler_staff.rusty", "base", 899338, "Rusty Channeler Staff (Ethereal)"),
			itemDoc("runeword.base.channeler_staff", "base", 899339, "Channeler Staff (Ethereal)"),
			itemDoc("runeword.base.channeler_staff.ultimate", "base", 899340, "Ultimate Channeler Staff (Ethereal)"),
			itemDoc("runeword.base.champion_shield.rusty", "base", 899341, "Rusty Champion Shield (Ethereal)"),
			itemDoc("runeword.base.champion_shield", "base", 899342, "Champion Shield (Ethereal)"),
			itemDoc("runeword.base.champion_shield.ultimate", "base", 899343, "Ultimate Champion Shield (Ethereal)"),
		},
		"families": []interface{}{
			familyDoc("annihilus", "Annihilus, Blade of Destruction", "melee_amp"),
			familyDoc("berserk", "Berserk, Dragonslayer's Greatsword", "melee_amp"),
			familyDoc("dark_sister", "Dark Sister", "proc"),
			familyDoc("zangetsu", "Zangetsu", "proc"),
			familyDoc("thoridal", "Thoridal, Star's Fury", "proc"),
			familyDoc("brightroar", "Brightroar, Sword of Amplification", "proc"),
			familyDoc("ariandel", "Ariandel, Summoner's Greatstaff", "pet"),
			familyDoc("atiesh", "Atiesh, Tal Rasha's Greatstaff", "mana"),
			familyDoc("seraph", "Seraph, Mace of Adjudication", "click"),
			familyDoc("honor", "Honor, Hand of Justice", "worn"),
			familyDoc("lordaeron", "Royal Crest of Lordaeron", "click"),
			familyDoc("doom", "Doom", "proc"),
		},
		"combos": defaultComboDocs(),
	}
}

func itemDoc(key, kind string, id int, name string) map[string]interface{} {
	return map[string]interface{}{
		"source_key":     key,
		"item_kind":      kind,
		"target_item_id": id,
		"name":           name,
	}
}

func familyDoc(key, label, domain string) map[string]interface{} {
	return map[string]interface{}{
		"family_key": key,
		"label":      label,
		"domain":     domain,
	}
}

func comboDoc(family, tier string, outputID int, outputName string, baseID int, r1, r2, r3, r4 int) map[string]interface{} {
	return map[string]interface{}{
		"combo_key":      family + "." + tier,
		"family_key":     family,
		"quality_tier":   tier,
		"output_item_id": outputID,
		"output_name":    outputName,
		"base_item_id":   baseID,
		"runes":          []interface{}{r1, r2, r3, r4},
		"effect_profile": family,
	}
}

func defaultComboDocs() []interface{} {
	el, ort, thul, um, vex, ber, jah := 899300, 899301, 899302, 899303, 899304, 899305, 899306
	return []interface{}{
		comboDoc("annihilus", "rusty", 899344, "Rusty Annihilus, Blade of Destruction", 899323, el, thul, vex, ber),
		comboDoc("annihilus", "normal", 899345, "Annihilus, Blade of Destruction", 899324, el, thul, vex, ber),
		comboDoc("annihilus", "ultimate", 899346, "Ultimate Annihilus, Blade of Destruction", 899325, el, thul, vex, ber),
		comboDoc("seraph", "rusty", 0, "Rusty Seraph, Mace of Adjudication", 899320, thul, um, vex, ber),
		comboDoc("seraph", "normal", 0, "Seraph, Mace of Adjudication", 899321, thul, um, vex, ber),
		comboDoc("seraph", "ultimate", 0, "Ultimate Seraph, Mace of Adjudication", 899322, thul, um, vex, ber),
		comboDoc("dark_sister", "rusty", 0, "Rusty Dark Sister", 899329, el, ort, vex, ber),
		comboDoc("dark_sister", "normal", 0, "Dark Sister", 899330, el, ort, vex, ber),
		comboDoc("dark_sister", "ultimate", 0, "Ultimate Dark Sister", 899331, el, ort, vex, ber),
		comboDoc("zangetsu", "rusty", 0, "Rusty Zangetsu", 899332, el, um, vex, ber),
		comboDoc("zangetsu", "normal", 0, "Zangetsu", 899333, el, um, vex, ber),
		comboDoc("zangetsu", "ultimate", 0, "Ultimate Zangetsu", 899334, el, um, vex, ber),
		comboDoc("ariandel", "rusty", 0, "Rusty Ariandel, Summoner's Greatstaff", 899338, ort, thul, um, ber),
		comboDoc("ariandel", "normal", 0, "Ariandel, Summoner's Greatstaff", 899339, ort, thul, um, ber),
		comboDoc("ariandel", "ultimate", 0, "Ultimate Ariandel, Summoner's Greatstaff", 899340, ort, thul, um, ber),
		comboDoc("thoridal", "rusty", 0, "Rusty Thoridal, Star's Fury", 899335, ort, thul, vex, ber),
		comboDoc("thoridal", "normal", 0, "Thoridal, Star's Fury", 899336, ort, thul, vex, ber),
		comboDoc("thoridal", "ultimate", 0, "Ultimate Thoridal, Star's Fury", 899337, ort, thul, vex, ber),
		comboDoc("brightroar", "rusty", 0, "Rusty Brightroar, Sword of Amplification", 899323, ort, um, vex, ber),
		comboDoc("brightroar", "normal", 0, "Brightroar, Sword of Amplification", 899324, ort, um, vex, ber),
		comboDoc("brightroar", "ultimate", 0, "Ultimate Brightroar, Sword of Amplification", 899325, ort, um, vex, ber),
		comboDoc("berserk", "rusty", 899365, "Rusty Berserk, Dragonslayer's Greatsword", 899326, ort, um, vex, ber),
		comboDoc("berserk", "normal", 899366, "Berserk, Dragonslayer's Greatsword", 899327, ort, um, vex, ber),
		comboDoc("berserk", "ultimate", 899367, "Ultimate Berserk, Dragonslayer's Greatsword", 899328, ort, um, vex, ber),
		comboDoc("atiesh", "rusty", 0, "Rusty Atiesh, Tal Rasha's Greatstaff", 899338, ort, um, vex, ber),
		comboDoc("atiesh", "normal", 0, "Atiesh, Tal Rasha's Greatstaff", 899339, ort, um, vex, ber),
		comboDoc("atiesh", "ultimate", 0, "Ultimate Atiesh, Tal Rasha's Greatstaff", 899340, ort, um, vex, ber),
		comboDoc("honor", "rusty", 0, "Rusty Honor, Hand of Justice", 899320, el, ort, vex, ber),
		comboDoc("honor", "normal", 0, "Honor, Hand of Justice", 899321, el, ort, vex, ber),
		comboDoc("honor", "ultimate", 0, "Ultimate Honor, Hand of Justice", 899322, el, ort, vex, ber),
		comboDoc("lordaeron", "rusty", 0, "Rusty Royal Crest of Lordaeron", 899341, el, um, ber, jah),
		comboDoc("lordaeron", "normal", 0, "Royal Crest of Lordaeron", 899342, el, um, ber, jah),
		comboDoc("lordaeron", "ultimate", 0, "Ultimate Royal Crest of Lordaeron", 899343, el, um, ber, jah),
	}
}

func (s *Service) loadOrSeedRunewordDoc() (map[string]interface{}, error) {
	if _, err := os.Stat(s.abs(runewordRel)); err == nil {
		return s.loadMap(runewordRel)
	}
	return defaultRunewordDoc(), nil
}

func (s *Service) writeRunewordItem(plan *WritePlan, req WriteRequest) error {
	key := strings.TrimSpace(textOf(req.Entry, "source_key", "sourceKey", "key"))
	if key == "" {
		return fmt.Errorf("source_key is required")
	}
	plan.Key = key
	doc, err := s.loadOrSeedRunewordDoc()
	if err != nil {
		return err
	}
	items := asArray(doc["items"])
	idx := -1
	var current map[string]interface{}
	for i, raw := range items {
		row := asObject(raw)
		if asString(row["source_key"]) == key {
			idx = i
			current = row
			break
		}
	}
	if req.Action == "delete" {
		if idx < 0 {
			return fmt.Errorf("runeword item %s not found", key)
		}
		doc["items"] = append(items[:idx], items[idx+1:]...)
		return s.commitJSON(plan, runewordRel, doc)
	}
	if current == nil {
		current = map[string]interface{}{
			"source_key":     key,
			"item_kind":      "rune",
			"target_item_id": 0,
			"name":           key,
		}
	}
	if v := textOf(req.Entry, "item_kind", "itemKind"); v != "" {
		current["item_kind"] = v
	}
	if v := textOf(req.Entry, "name"); v != "" {
		current["name"] = v
	}
	if req.Entry["target_item_id"] != nil || req.Entry["targetItemId"] != nil {
		n := asInt(req.Entry["target_item_id"])
		if n == 0 {
			n = asInt(req.Entry["targetItemId"])
		}
		current["target_item_id"] = n
	}
	if idx >= 0 {
		items[idx] = current
	} else {
		items = append(items, current)
	}
	doc["items"] = items
	return s.commitJSON(plan, runewordRel, doc)
}

func (s *Service) writeRunewordCombo(plan *WritePlan, req WriteRequest) error {
	key := strings.TrimSpace(textOf(req.Entry, "combo_key", "comboKey", "key"))
	if key == "" {
		family := strings.TrimSpace(textOf(req.Entry, "family_key", "familyKey"))
		tier := strings.TrimSpace(textOf(req.Entry, "quality_tier", "qualityTier"))
		if family != "" && tier != "" {
			key = family + "." + tier
		}
	}
	if key == "" {
		return fmt.Errorf("combo_key or family_key + quality_tier is required")
	}
	plan.Key = key
	doc, err := s.loadOrSeedRunewordDoc()
	if err != nil {
		return err
	}
	combos := asArray(doc["combos"])
	idx := -1
	var current map[string]interface{}
	for i, raw := range combos {
		row := asObject(raw)
		if asString(row["combo_key"]) == key {
			idx = i
			current = row
			break
		}
	}
	if req.Action == "delete" {
		if idx < 0 {
			return fmt.Errorf("runeword combo %s not found", key)
		}
		doc["combos"] = append(combos[:idx], combos[idx+1:]...)
		return s.commitJSON(plan, runewordRel, doc)
	}
	if current == nil {
		current = map[string]interface{}{
			"combo_key":      key,
			"family_key":     "",
			"quality_tier":   "normal",
			"output_item_id": 0,
			"output_name":    "",
			"base_item_id":   0,
			"runes":          []interface{}{0, 0, 0, 0},
			"effect_profile": "",
		}
	}
	if v := textOf(req.Entry, "family_key", "familyKey"); v != "" {
		current["family_key"] = v
		if asString(current["effect_profile"]) == "" {
			current["effect_profile"] = v
		}
	}
	if v := textOf(req.Entry, "quality_tier", "qualityTier"); v != "" {
		current["quality_tier"] = v
	}
	if v := textOf(req.Entry, "output_name", "outputName"); v != "" {
		current["output_name"] = v
	}
	if v := textOf(req.Entry, "effect_profile", "effectProfile"); v != "" {
		current["effect_profile"] = v
	}
	if req.Entry["output_item_id"] != nil || req.Entry["outputItemId"] != nil {
		n := asInt(req.Entry["output_item_id"])
		if n == 0 {
			n = asInt(req.Entry["outputItemId"])
		}
		current["output_item_id"] = n
	}
	if req.Entry["base_item_id"] != nil || req.Entry["baseItemId"] != nil {
		n := asInt(req.Entry["base_item_id"])
		if n == 0 {
			n = asInt(req.Entry["baseItemId"])
		}
		current["base_item_id"] = n
	}
	runes := intSlice(current["runes"])
	for len(runes) < 4 {
		runes = append(runes, 0)
	}
	setRune := func(slot int, names ...string) {
		for _, name := range names {
			if req.Entry[name] != nil {
				runes[slot] = asInt(req.Entry[name])
				return
			}
		}
	}
	setRune(0, "rune1", "rune_1_item_id")
	setRune(1, "rune2", "rune_2_item_id")
	setRune(2, "rune3", "rune_3_item_id")
	setRune(3, "rune4", "rune_4_item_id")
	if req.Entry["runes"] != nil {
		got := intSlice(req.Entry["runes"])
		for i := 0; i < 4 && i < len(got); i++ {
			runes[i] = got[i]
		}
	}
	arr := make([]interface{}, 4)
	for i := 0; i < 4; i++ {
		arr[i] = runes[i]
	}
	current["runes"] = arr
	if idx >= 0 {
		combos[idx] = current
	} else {
		combos = append(combos, current)
	}
	doc["combos"] = combos
	return s.commitJSON(plan, runewordRel, doc)
}

func (s *Service) seedRunewords(plan *WritePlan, req WriteRequest) error {
	plan.Key = "seed"
	if req.Action != "seed" && req.Action != "upsert" {
		return fmt.Errorf("runeword seed only supports upsert/seed")
	}
	if _, err := os.Stat(s.abs(runewordRel)); err == nil && !truthy(req.Entry["overwrite"]) {
		return fmt.Errorf("runeword_catalog.json already exists; set overwrite to replace it")
	}
	return s.commitJSON(plan, runewordRel, defaultRunewordDoc())
}

func truthy(v interface{}) bool {
	return v == true || asString(v) == "true" || asString(v) == "1"
}

func (s *Service) probeRunewordTables() RunewordTables {
	out := RunewordTables{Counts: map[string]int{}}
	if s.db == nil {
		out.Error = "database resolver not attached"
		return out
	}
	gdb := s.db.GetEqemuDb()
	if gdb == nil {
		out.Error = "local peq connection is not ready"
		return out
	}
	names := []string{
		"custom_runeword_item_map",
		"custom_runeword_combo_map",
		"custom_runeword_effect_map",
		"custom_runeword_spell_map",
	}
	found := 0
	for _, name := range names {
		var n int
		err := gdb.Raw(`
			SELECT COUNT(*)
			FROM information_schema.tables
			WHERE table_schema = DATABASE() AND table_name = ?
		`, name).Scan(&n).Error
		if err != nil || n == 0 {
			out.Counts[name] = -1
			continue
		}
		var rows int
		_ = gdb.Raw("SELECT COUNT(*) FROM `" + name + "`").Scan(&rows).Error
		out.Counts[name] = rows
		found++
	}
	out.Available = found > 0
	return out
}
