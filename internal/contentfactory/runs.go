package contentfactory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type RunRecord struct {
	ID       string      `json:"id"`
	Kind     string      `json:"kind"`
	At       string      `json:"at"`
	DryRun   bool        `json:"dryRun"`
	Summary  string      `json:"summary"`
	RecipeID string      `json:"recipeId,omitempty"`
	Writes   []FileWrite `json:"writes"`
	DBWrites []DBWrite   `json:"dbWrites"`
	Clones   []Clone     `json:"clones"`
	Tables   []LootMade  `json:"tables,omitempty"`
	Undone   bool        `json:"undone,omitempty"`
}

type RunList struct {
	OK    bool        `json:"ok"`
	Error string      `json:"error,omitempty"`
	Rel   string      `json:"relPath"`
	Runs  []RunRecord `json:"runs"`
}

func (s *Service) runDir() string {
	return filepath.Join(s.questsDirPath(), filepath.FromSlash(runRel))
}

func (s *Service) saveRun(plan *Plan) {
	if plan == nil || plan.DryRun || plan.Error != "" {
		return
	}
	if s.questsDirPath() == "" {
		plan.Warnings = append(plan.Warnings, "run history not saved: quests directory is not set")
		return
	}
	id := time.Now().UTC().Format("20060102-150405")
	rec := RunRecord{
		ID:       id,
		Kind:     plan.Kind,
		At:       time.Now().UTC().Format(time.RFC3339),
		Summary:  runSummary(*plan),
		RecipeID: plan.RecipeID,
		Writes:   plan.Writes,
		DBWrites: plan.DBWrites,
		Clones:   plan.Clones,
		Tables:   plan.Tables,
	}
	abs := filepath.Join(s.runDir(), id+".json")
	if err := writeJSON(abs, rec); err != nil {
		plan.Warnings = append(plan.Warnings, "run history: "+err.Error())
		return
	}
	plan.RunID = id
	s.pruneRuns()
}

func runSummary(plan Plan) string {
	return fmt.Sprintf("%s: %d db write(s), %d clone(s), %d file(s)", plan.Kind, len(plan.DBWrites), len(plan.Clones), len(plan.Writes))
}

func (s *Service) pruneRuns() {
	entries, err := os.ReadDir(s.runDir())
	if err != nil {
		return
	}
	names := []string{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	if len(names) <= maxRuns {
		return
	}
	for _, name := range names[:len(names)-maxRuns] {
		_ = os.Remove(filepath.Join(s.runDir(), name))
	}
}

func (s *Service) ListRuns() RunList {
	out := RunList{Rel: runRel, Runs: []RunRecord{}}
	if s.questsDirPath() == "" {
		out.Error = "quests directory is not set"
		return out
	}
	entries, err := os.ReadDir(s.runDir())
	if err != nil {
		if os.IsNotExist(err) {
			out.OK = true
			return out
		}
		out.Error = err.Error()
		return out
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(s.runDir(), e.Name()))
		if err != nil {
			continue
		}
		var rec RunRecord
		if err := json.Unmarshal(raw, &rec); err != nil {
			continue
		}
		out.Runs = append(out.Runs, rec)
	}
	sort.Slice(out.Runs, func(i, j int) bool { return out.Runs[i].ID > out.Runs[j].ID })
	out.OK = true
	return out
}

func (s *Service) loadRun(id string) (RunRecord, error) {
	id = filepath.Base(strings.TrimSpace(id))
	id = strings.TrimSuffix(id, ".json")
	if id == "" || strings.Contains(id, "..") {
		return RunRecord{}, fmt.Errorf("invalid run id")
	}
	raw, err := os.ReadFile(filepath.Join(s.runDir(), id+".json"))
	if err != nil {
		return RunRecord{}, err
	}
	var rec RunRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return RunRecord{}, err
	}
	return rec, nil
}

type UndoRequest struct {
	RunID string `json:"runId"`
}

func (s *Service) Undo(req UndoRequest) Plan {
	plan := emptyPlan("undo", false)
	rec, err := s.loadRun(req.RunID)
	if err != nil {
		return plan.fail(err.Error())
	}
	if rec.Undone {
		return plan.fail("run " + rec.ID + " was already undone")
	}
	if s.eqemu() == nil {
		return plan.fail("PEQ database is not connected")
	}
	for i := len(rec.DBWrites) - 1; i >= 0; i-- {
		w := rec.DBWrites[i]
		if err := s.undoWrite(w); err != nil {
			plan.Warnings = append(plan.Warnings, err.Error())
			continue
		}
		plan.DBWrites = append(plan.DBWrites, DBWrite{Table: w.Table, Action: "undo-" + w.Action, ID: w.ID, Note: w.Note, Extra: w.Extra})
	}
	rec.Undone = true
	if err := writeJSON(filepath.Join(s.runDir(), rec.ID+".json"), rec); err != nil {
		plan.Warnings = append(plan.Warnings, err.Error())
	}
	plan.RunID = rec.ID
	plan.OK = true
	return plan
}

func (s *Service) undoWrite(w DBWrite) error {
	db := s.eqemu()
	switch w.Table + "/" + w.Action {
	case "items/create":
		return db.Exec("DELETE FROM items WHERE id = ?", w.ID).Error
	case "spells_new/create":
		return db.Exec("DELETE FROM spells_new WHERE id = ?", w.ID).Error
	case "loottable/create":
		return db.Exec("DELETE FROM loottable WHERE id = ?", w.ID).Error
	case "lootdrop/create":
		return db.Exec("DELETE FROM lootdrop WHERE id = ?", w.ID).Error
	case "loottable_entries/create":
		return db.Exec("DELETE FROM loottable_entries WHERE loottable_id = ? AND lootdrop_id = ?", extra(w, "loottableId"), extra(w, "lootdropId")).Error
	case "lootdrop_entries/create":
		return db.Exec("DELETE FROM lootdrop_entries WHERE lootdrop_id = ? AND item_id = ?", extra(w, "lootdropId"), extraOr(w, "itemId", w.ID)).Error
	case "merchantlist/create":
		return db.Exec("DELETE FROM merchantlist WHERE merchantid = ? AND slot = ?", extra(w, "merchantId"), extra(w, "slot")).Error
	case "merchantlist/delete":
		return fmt.Errorf("merchant slot restore is not stored; skipped item %d", w.ID)
	case "npc_types/update":
		values := map[string]interface{}{}
		if v, ok := w.Extra["loottableId"]; ok {
			values["loottable_id"] = v
		}
		if v, ok := w.Extra["npcSpellsId"]; ok {
			values["npc_spells_id"] = v
		}
		if len(values) == 0 {
			return nil
		}
		return db.Table("npc_types").Where("id = ?", w.ID).Updates(values).Error
	case "items/update":
		return nil
	case "npc_spells/create":
		_ = db.Exec("DELETE FROM npc_spells_entries WHERE npc_spells_id = ?", w.ID).Error
		return db.Exec("DELETE FROM npc_spells WHERE id = ?", w.ID).Error
	case "npc_spells_entries/create":
		return db.Exec("DELETE FROM npc_spells_entries WHERE npc_spells_id = ? AND spellid = ?", extra(w, "npcSpellsId"), extraOr(w, "spellId", w.ID)).Error
	case "npc_spells_entries/delete":
		return nil
	case "npc_types/create":
		return db.Exec("DELETE FROM npc_types WHERE id = ?", w.ID).Error
	case "spawngroup/create":
		return db.Exec("DELETE FROM spawngroup WHERE id = ?", w.ID).Error
	case "spawn2/create":
		return db.Exec("DELETE FROM spawn2 WHERE id = ?", w.ID).Error
	case "spawnentry/create":
		return db.Exec("DELETE FROM spawnentry WHERE spawngroupID = ? AND npcID = ?", extra(w, "spawngroupId"), extraOr(w, "npcId", w.ID)).Error
	case "spawnentry/update":
		oldID := extra(w, "oldNpcId")
		newID := extraOr(w, "newNpcId", w.ID)
		if oldID <= 0 || newID <= 0 {
			return nil
		}
		if zoneID := extra(w, "zoneId"); zoneID > 0 {
			short, err := s.zoneShort(zoneID)
			if err != nil {
				return db.Exec("UPDATE spawnentry SET npcID = ? WHERE npcID = ?", oldID, newID).Error
			}
			return db.Exec("UPDATE spawnentry se JOIN spawn2 s2 ON s2.spawngroupID = se.spawngroupID SET se.npcID = ? WHERE s2.zone = ? AND se.npcID = ?", oldID, short, newID).Error
		}
		return db.Exec("UPDATE spawnentry SET npcID = ? WHERE npcID = ?", oldID, newID).Error
	case "inventory/create":
		return db.Exec("DELETE FROM inventory WHERE character_id = ? AND slot_id = ?", extra(w, "characterId"), extra(w, "slot")).Error
	default:
		return fmt.Errorf("cannot undo %s %s %d", w.Table, w.Action, w.ID)
	}
}

func extra(w DBWrite, key string) int {
	if w.Extra == nil {
		return 0
	}
	return w.Extra[key]
}

func extraOr(w DBWrite, key string, fallback int) int {
	if v := extra(w, key); v > 0 {
		return v
	}
	return fallback
}
