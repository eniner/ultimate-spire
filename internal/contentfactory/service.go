package contentfactory

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/EQEmu/spire/internal/database"
	"github.com/EQEmu/spire/internal/eqemuserver"
	"github.com/EQEmu/spire/internal/pathmgmt"
	"gorm.io/gorm"
)

const (
	reservedFloor     = 800000
	npcCastSpellFloor = 50000
	recipeRel         = "global/ultimatedata/_spire_recipes/recipes.json"
	runRel            = "global/ultimatedata/_spire_runs"
	exportRel         = "global/ultimatedata/_spire_export"
	draftRel          = "global/ultimatedata/_spire_drafts/content-factory.json"
	maxRuns           = 40
	npcSpellIDCap     = 65535
	censusNpcLimit    = 400
	catalogLimit      = 80
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

type Status struct {
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	QuestsDir string `json:"questsDir"`
	RecipeRel string `json:"recipeRel"`
	RunRel    string `json:"runRel"`
	Reserved  int    `json:"reserved"`
}

func (s *Service) Status() Status {
	out := Status{RecipeRel: recipeRel, RunRel: runRel, Reserved: reservedFloor}
	dir := s.questsDirPath()
	out.QuestsDir = dir
	if dir == "" {
		out.Error = "quests directory is not set"
		return out
	}
	out.OK = true
	return out
}

type DBWrite struct {
	Table  string         `json:"table"`
	Action string         `json:"action"`
	ID     int            `json:"id"`
	Note   string         `json:"note,omitempty"`
	Extra  map[string]int `json:"extra,omitempty"`
}

type FileWrite struct {
	Kind    string `json:"kind"`
	RelPath string `json:"relPath"`
	Action  string `json:"action"`
}

type Clone struct {
	Kind     string `json:"kind"`
	SourceID int    `json:"sourceId"`
	NewID    int    `json:"newId"`
	Name     string `json:"name"`
	Step     int    `json:"step"`
	Slot     string      `json:"slot,omitempty"`
	Class    string      `json:"class,omitempty"`
	Slots    int         `json:"slots,omitempty"`
	Classes  int         `json:"classes,omitempty"`
	Diff     []DiffField `json:"diff,omitempty"`
}

type DiffField struct {
	Field string `json:"field"`
	From  string `json:"from"`
	To    string `json:"to"`
}

type ImpactRow struct {
	Kind   string `json:"kind"`
	ID     int    `json:"id"`
	Name   string `json:"name,omitempty"`
	Field  string `json:"field"`
	From   int    `json:"from"`
	To     int    `json:"to"`
	ZoneID int    `json:"zoneId,omitempty"`
	Note   string `json:"note,omitempty"`
	Role   string `json:"role,omitempty"`
	Key    string `json:"key,omitempty"`
}

type ExportInfo struct {
	OK         bool     `json:"ok"`
	SpellsPath string   `json:"spellsPath,omitempty"`
	DbStrPath  string   `json:"dbstrPath,omitempty"`
	SpellRows  int      `json:"spellRows,omitempty"`
	Note       string   `json:"note,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
}

type ReloadInfo struct {
	OK        bool     `json:"ok"`
	Note      string   `json:"note,omitempty"`
	Reloaded  []string `json:"reloaded,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
	WorldUsed bool     `json:"worldUsed"`
}

type Plan struct {
	OK        bool        `json:"ok"`
	DryRun    bool        `json:"dryRun"`
	Error     string      `json:"error,omitempty"`
	Kind      string      `json:"kind"`
	RunID     string      `json:"runId,omitempty"`
	RecipeID  string      `json:"recipeId,omitempty"`
	Writes    []FileWrite `json:"writes"`
	DBWrites  []DBWrite   `json:"dbWrites"`
	Clones    []Clone     `json:"clones"`
	Warnings  []string    `json:"warnings"`
	Impact    []ImpactRow `json:"impact,omitempty"`
	Attached  int         `json:"attached,omitempty"`
	Tables    []LootMade  `json:"tables,omitempty"`
	Merchants []int       `json:"merchants,omitempty"`
	Export    *ExportInfo `json:"export,omitempty"`
	Reload    *ReloadInfo `json:"reload,omitempty"`
	Probe     *ProbeResult `json:"probe,omitempty"`
	Lint      []LintItem  `json:"lint,omitempty"`
	Synced    int         `json:"synced,omitempty"`
}

func emptyPlan(kind string, dry bool) Plan {
	return Plan{Kind: kind, DryRun: dry, Writes: []FileWrite{}, DBWrites: []DBWrite{}, Clones: []Clone{}, Warnings: []string{}, Impact: []ImpactRow{}}
}

func mergePlan(dst *Plan, src Plan) {
	if dst == nil {
		return
	}
	dst.Writes = append(dst.Writes, src.Writes...)
	dst.DBWrites = append(dst.DBWrites, src.DBWrites...)
	dst.Clones = append(dst.Clones, src.Clones...)
	dst.Warnings = append(dst.Warnings, src.Warnings...)
	dst.Impact = append(dst.Impact, src.Impact...)
	dst.Tables = append(dst.Tables, src.Tables...)
	dst.Merchants = uniqueInts(append(dst.Merchants, src.Merchants...))
	dst.Attached += src.Attached
	dst.Synced += src.Synced
	if src.RecipeID != "" {
		dst.RecipeID = src.RecipeID
	}
	if src.Export != nil {
		dst.Export = src.Export
	}
	if src.Reload != nil {
		dst.Reload = src.Reload
	}
	if src.Probe != nil {
		dst.Probe = src.Probe
	}
	dst.Lint = append(dst.Lint, src.Lint...)
	if src.Error != "" && dst.Error == "" {
		dst.Error = src.Error
		dst.OK = false
	}
}

func (p *Plan) fail(err string) Plan {
	p.Error = err
	p.OK = false
	return *p
}

type idSpec struct {
	Key    string
	Table  string
	Column string
	Floor  int
	Note   string
}

func idSpecs() []idSpec {
	return []idSpec{
		{"items", "items", "id", reservedFloor, "Custom gear and clickies. Tier Factory also mints from here."},
		{"spells_new", "spells_new", "id", reservedFloor, "Item click/proc/worn/focus. Keep NPC-castable clones at or below 65535."},
		{"npc_cast_spells", "spells_new", "id", npcCastSpellFloor, "NPC-castable spell IDs. Next free at or below 65535."},
		{"npc_types", "npc_types", "id", reservedFloor, "Custom NPCs."},
		{"npc_spells", "npc_spells", "id", reservedFloor, "NPC spell set headers."},
		{"loottable", "loottable", "id", reservedFloor, "Loot table headers."},
		{"lootdrop", "lootdrop", "id", reservedFloor, "Loot drop groups inside a table."},
		{"merchantlist", "merchantlist", "merchantid", reservedFloor, "Merchant IDs (not slot)."},
	}
}

type IDRow struct {
	Key      string `json:"key"`
	Table    string `json:"table"`
	Column   string `json:"column"`
	Floor    int    `json:"floor"`
	MaxID    int    `json:"maxId"`
	NextID   int    `json:"nextId"`
	Reserved int    `json:"reservedUsed"`
	Note     string `json:"note"`
}

type IDBoard struct {
	OK       bool    `json:"ok"`
	Error    string  `json:"error,omitempty"`
	Reserved int     `json:"reserved"`
	Rows     []IDRow `json:"rows"`
}

func (s *Service) IDBoard() IDBoard {
	out := IDBoard{Reserved: reservedFloor, Rows: []IDRow{}}
	db := s.eqemu()
	if db == nil {
		out.Error = "PEQ database is not connected"
		for _, spec := range idSpecs() {
			out.Rows = append(out.Rows, IDRow{
				Key: spec.Key, Table: spec.Table, Column: spec.Column, Floor: spec.Floor,
				MaxID: 0, NextID: spec.Floor, Note: spec.Note,
			})
		}
		return out
	}
	out.OK = true
	for _, spec := range idSpecs() {
		row := IDRow{Key: spec.Key, Table: spec.Table, Column: spec.Column, Floor: spec.Floor, Note: spec.Note}
		if spec.Key == "npc_cast_spells" {
			var max int
			if err := db.Table(spec.Table).Select("COALESCE(MAX("+spec.Column+"), 0)").Where(spec.Column+" <= ?", npcSpellIDCap).Scan(&max).Error; err != nil {
				row.Note = spec.Note + " (" + err.Error() + ")"
			}
			row.MaxID = max
			row.NextID = nextFromMax(max, spec.Floor)
			if row.NextID > npcSpellIDCap {
				row.NextID = 0
				row.Note = spec.Note + " No free IDs at or below 65535."
			}
			var used int64
			_ = db.Table(spec.Table).Where(spec.Column+" BETWEEN ? AND ?", spec.Floor, npcSpellIDCap).Count(&used).Error
			row.Reserved = int(used)
			out.Rows = append(out.Rows, row)
			continue
		}
		var max int
		if err := db.Table(spec.Table).Select("COALESCE(MAX(" + spec.Column + "), 0)").Scan(&max).Error; err != nil {
			row.Note = spec.Note + " (" + err.Error() + ")"
		}
		row.MaxID = max
		row.NextID = nextFromMax(max, spec.Floor)
		var used int64
		_ = db.Table(spec.Table).Where(spec.Column+" >= ?", spec.Floor).Count(&used).Error
		row.Reserved = int(used)
		out.Rows = append(out.Rows, row)
	}
	return out
}

func nextFromMax(max, floor int) int {
	if max < floor {
		return floor
	}
	return max + 1
}

func (s *Service) nextID(table, column string, floor int) (int, error) {
	return s.nextIDAtMost(table, column, floor, 0)
}

func (s *Service) nextIDAtMost(table, column string, floor, cap int) (int, error) {
	if floor <= 0 {
		floor = reservedFloor
	}
	if s.eqemu() == nil {
		return floor, nil
	}
	var max int
	tx := s.eqemu().Table(table).Select("COALESCE(MAX(" + column + "), 0)")
	if cap > 0 {
		tx = tx.Where(column+" <= ?", cap)
	}
	if err := tx.Scan(&max).Error; err != nil {
		return 0, err
	}
	next := nextFromMax(max, floor)
	if cap > 0 && next > cap {
		return 0, fmt.Errorf("no free %s.%s at or below %d", table, column, cap)
	}
	return next, nil
}

func (s *Service) existingIDs(table, column string, start, end int) (map[int]bool, error) {
	out := map[int]bool{}
	if s.eqemu() == nil || end < start {
		return out, nil
	}
	var ids []int
	if err := s.eqemu().Table(table).Where(column+" BETWEEN ? AND ?", start, end).Pluck(column, &ids).Error; err != nil {
		return nil, err
	}
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

func (s *Service) takeIDs(table, column string, start, count, floor int, plan *Plan) ([]int, error) {
	return s.takeIDsCapped(table, column, start, count, floor, 0, plan)
}

func (s *Service) takeIDsCapped(table, column string, start, count, floor, cap int, plan *Plan) ([]int, error) {
	if count < 1 {
		count = 1
	}
	if start <= 0 {
		n, err := s.nextIDAtMost(table, column, floor, cap)
		if err != nil {
			return nil, err
		}
		start = n
	}
	if cap > 0 && start > cap {
		return nil, fmt.Errorf("start id %d is above cap %d for %s", start, cap, table)
	}
	probe := start + count + 64
	if cap > 0 && probe > cap {
		probe = cap
	}
	exists, err := s.existingIDs(table, column, start, probe)
	if err != nil {
		return nil, err
	}
	out := make([]int, 0, count)
	next := start
	for len(out) < count {
		if cap > 0 && next > cap {
			return out, fmt.Errorf("could not reserve %d free %s ids at or below %d", count, table, cap)
		}
		if exists[next] {
			if plan != nil {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("%s %s %d already exists; skipped", table, column, next))
			}
			next++
			if next > start+count+4096 {
				return out, fmt.Errorf("could not reserve %d free %s ids near %d", count, table, start)
			}
			continue
		}
		out = append(out, next)
		next++
	}
	return out, nil
}

type SearchRow struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Extra  string `json:"extra,omitempty"`
	Slots  int    `json:"slots,omitempty"`
	Class  int    `json:"classes,omitempty"`
	Level  int    `json:"level,omitempty"`
	Merchant int  `json:"merchantId,omitempty"`
}

type SearchResult struct {
	OK    bool        `json:"ok"`
	Error string      `json:"error,omitempty"`
	Kind  string      `json:"kind"`
	Query string      `json:"query"`
	Rows  []SearchRow `json:"rows"`
}

func (s *Service) Search(kind, q string, limit int) SearchResult {
	kind = strings.ToLower(strings.TrimSpace(kind))
	q = strings.TrimSpace(q)
	if limit <= 0 || limit > 80 {
		limit = 25
	}
	out := SearchResult{Kind: kind, Query: q, Rows: []SearchRow{}}
	if s.eqemu() == nil {
		out.Error = "PEQ database is not connected"
		return out
	}
	if q == "" {
		out.OK = true
		return out
	}
	like := "%" + q + "%"
	id, _ := strconv.Atoi(q)
	switch kind {
	case "spells", "spells_new":
		type row struct {
			ID   int
			Name string
		}
		var rows []row
		tx := s.eqemu().Table("spells_new").Select("id, name").Limit(limit)
		if id > 0 {
			tx = tx.Where("id = ? OR name LIKE ?", id, like)
		} else {
			tx = tx.Where("name LIKE ?", like)
		}
		if err := tx.Scan(&rows).Error; err != nil {
			out.Error = err.Error()
			return out
		}
		for _, r := range rows {
			out.Rows = append(out.Rows, SearchRow{ID: r.ID, Name: r.Name})
		}
	case "npcs", "npc_types":
		type row struct {
			ID         int
			Name       string
			Level      int
			MerchantID int `gorm:"column:merchant_id"`
		}
		var rows []row
		tx := s.eqemu().Table("npc_types").Select("id, name, level, merchant_id").Limit(limit)
		if id > 0 {
			tx = tx.Where("id = ? OR name LIKE ?", id, like)
		} else {
			tx = tx.Where("name LIKE ?", like)
		}
		if err := tx.Scan(&rows).Error; err != nil {
			out.Error = err.Error()
			return out
		}
		for _, r := range rows {
			out.Rows = append(out.Rows, SearchRow{ID: r.ID, Name: r.Name, Level: r.Level, Merchant: r.MerchantID})
		}
	case "npc_spells", "spellsets":
		type row struct {
			ID   int
			Name string
		}
		var rows []row
		tx := s.eqemu().Table("npc_spells").Select("id, name").Limit(limit)
		if id > 0 {
			tx = tx.Where("id = ? OR name LIKE ?", id, like)
		} else {
			tx = tx.Where("name LIKE ?", like)
		}
		if err := tx.Scan(&rows).Error; err != nil {
			out.Error = err.Error()
			return out
		}
		for _, r := range rows {
			out.Rows = append(out.Rows, SearchRow{ID: r.ID, Name: r.Name})
		}
	case "merchants":
		type row struct {
			ID         int
			Name       string
			MerchantID int `gorm:"column:merchant_id"`
		}
		var rows []row
		tx := s.eqemu().Table("npc_types").Select("id, name, merchant_id").Where("merchant_id > 0").Limit(limit)
		if id > 0 {
			tx = tx.Where("merchant_id = ? OR id = ? OR name LIKE ?", id, id, like)
		} else {
			tx = tx.Where("name LIKE ?", like)
		}
		if err := tx.Scan(&rows).Error; err != nil {
			out.Error = err.Error()
			return out
		}
		for _, r := range rows {
			out.Rows = append(out.Rows, SearchRow{ID: r.MerchantID, Name: r.Name, Extra: fmt.Sprintf("npc %d", r.ID), Merchant: r.MerchantID})
		}
	case "characters", "character", "character_data":
		type row struct {
			ID    int
			Name  string
			Level int
		}
		var rows []row
		tx := s.eqemu().Table("character_data").Select("id, name, level").Limit(limit)
		if id > 0 {
			tx = tx.Where("id = ? OR name LIKE ?", id, like)
		} else {
			tx = tx.Where("name LIKE ?", like)
		}
		if err := tx.Scan(&rows).Error; err != nil {
			out.Error = err.Error()
			return out
		}
		for _, r := range rows {
			out.Rows = append(out.Rows, SearchRow{ID: r.ID, Name: r.Name, Level: r.Level})
		}
	default:
		type row struct {
			ID      int
			Name    string `gorm:"column:Name"`
			Slots   int
			Classes int
		}
		var rows []row
		tx := s.eqemu().Table("items").Select("id, Name, slots, classes").Limit(limit)
		if id > 0 {
			tx = tx.Where("id = ? OR Name LIKE ?", id, like)
		} else {
			tx = tx.Where("Name LIKE ?", like)
		}
		if err := tx.Scan(&rows).Error; err != nil {
			out.Error = err.Error()
			return out
		}
		for _, r := range rows {
			out.Rows = append(out.Rows, SearchRow{ID: r.ID, Name: r.Name, Slots: r.Slots, Class: r.Classes})
		}
	}
	out.OK = true
	return out
}

func scaleInt(v *int, mult float64) {
	if v == nil || *v == 0 || math.Abs(mult-1) <= 0.0001 {
		return
	}
	n := int(math.Round(float64(*v) * mult))
	if n == 0 {
		if *v < 0 {
			n = -1
		} else {
			n = 1
		}
	}
	*v = n
}

func scaleInt16(v *int16, mult float64) {
	if v == nil || *v <= 0 || math.Abs(mult-1) <= 0.0001 {
		return
	}
	n := int(math.Round(float64(*v) * mult))
	if n < 1 {
		n = 1
	}
	if n > 32767 {
		n = 32767
	}
	*v = int16(n)
}

func scalePositive(v *int, mult float64) {
	if v == nil || *v <= 0 || math.Abs(mult-1) <= 0.0001 {
		return
	}
	n := int(math.Round(float64(*v) * mult))
	if n < 1 {
		n = 1
	}
	*v = n
}

func ladderMultiplier(step float64, index int) float64 {
	if step <= 0 {
		step = 1
	}
	if index <= 0 {
		return 1
	}
	return math.Pow(step, float64(index))
}

func prefixedName(prefix, name string) string {
	name = strings.TrimSpace(name)
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || name == "" {
		return name
	}
	if strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
		return name
	}
	return strings.TrimSpace(prefix + " " + name)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func parseIDs(raw string) []int {
	out := []int{}
	seen := map[int]bool{}
	cur := ""
	flush := func() {
		cur = strings.TrimSpace(cur)
		if cur == "" {
			return
		}
		n, err := strconv.Atoi(cur)
		if err == nil && n > 0 && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
		cur = ""
	}
	for _, r := range raw {
		if r == ',' || r == ' ' || r == '\n' || r == '\t' || r == ';' {
			flush()
			continue
		}
		cur += string(r)
	}
	flush()
	return out
}

func uniqueInts(in []int) []int {
	seen := map[int]bool{}
	out := []int{}
	for _, n := range in {
		if n > 0 && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func (s *Service) zoneShort(id int) (string, error) {
	if s.eqemu() == nil {
		return "", fmt.Errorf("PEQ database is not connected")
	}
	var short string
	err := s.eqemu().Table("zone").Select("short_name").Where("zoneidnumber = ?", id).Order("version asc").Limit(1).Scan(&short).Error
	short = strings.TrimSpace(short)
	if err != nil || short == "" {
		return "", fmt.Errorf("no PEQ short_name for zone %d", id)
	}
	return short, nil
}

func (s *Service) npcIDsInZones(zoneIDs []int, nameContains string) ([]int, error) {
	out := []int{}
	if s.eqemu() == nil {
		return out, fmt.Errorf("PEQ database is not connected")
	}
	nameContains = strings.ToLower(strings.TrimSpace(nameContains))
	seen := map[int]bool{}
	for _, zoneID := range uniqueInts(zoneIDs) {
		short, err := s.zoneShort(zoneID)
		if err != nil {
			return out, err
		}
		tx := s.eqemu().Table("spawn2").
			Select("DISTINCT spawnentry.npcID").
			Joins("JOIN spawnentry ON spawnentry.spawngroupID = spawn2.spawngroupID").
			Joins("JOIN npc_types ON npc_types.id = spawnentry.npcID").
			Where("spawn2.zone = ?", short)
		if nameContains != "" {
			tx = tx.Where("LOWER(npc_types.name) LIKE ?", "%"+nameContains+"%")
		}
		var ids []int
		if err := tx.Pluck("spawnentry.npcID", &ids).Error; err != nil {
			return out, err
		}
		for _, id := range ids {
			if id > 0 && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out, nil
}

func writeJSON(abs string, v interface{}) error {
	if err := ensureDir(filepath.Dir(abs)); err != nil {
		return err
	}
	body, err := marshalPretty(v)
	if err != nil {
		return err
	}
	return os.WriteFile(abs, body, 0o644)
}
