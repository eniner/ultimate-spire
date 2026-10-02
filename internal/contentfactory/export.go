package contentfactory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EQEmu/spire/internal/database"
	"github.com/EQEmu/spire/internal/models"
)

func (s *Service) ExportClient(includeDbStr bool) ExportInfo {
	out := ExportInfo{Warnings: []string{}}
	if s.eqemu() == nil {
		out.Note = "PEQ database is not connected"
		return out
	}
	spells, n, err := s.exportCaretTable("spells_new", &models.SpellsNew{}, false)
	if err != nil {
		out.Note = err.Error()
		return out
	}
	out.SpellRows = n
	paths := []string{}
	if p, err := s.writeExportFile("spells_us.txt", spells); err != nil {
		out.Warnings = append(out.Warnings, err.Error())
	} else {
		out.SpellsPath = p
		paths = append(paths, p)
	}
	if includeDbStr {
		dbstr, _, err := s.exportCaretTable("db_str", &models.DbStr{}, true)
		if err != nil {
			out.Warnings = append(out.Warnings, "dbstr: "+err.Error())
		} else if p, err := s.writeExportFile("dbstr_us.txt", dbstr); err != nil {
			out.Warnings = append(out.Warnings, err.Error())
		} else {
			out.DbStrPath = p
			paths = append(paths, p)
		}
	}
	if out.SpellsPath == "" {
		out.Note = "Could not write spells_us.txt"
		return out
	}
	out.OK = true
	out.Note = fmt.Sprintf("Wrote %d spells to %s. Copy into the client folder.", n, strings.Join(paths, " and "))
	return out
}

func (s *Service) exportCaretTable(table string, model interface{}, extraZero bool) (string, int, error) {
	var entries []map[string]interface{}
	if err := s.eqemu().Model(model).Find(&entries).Error; err != nil {
		return "", 0, err
	}
	columns, err := database.GetTableSchema(s.eqemu(), table)
	if err != nil {
		return "", 0, err
	}
	rows := make([]string, 0, len(entries))
	for _, entry := range entries {
		cols := make([]string, 0, len(columns)+1)
		for _, column := range columns {
			cols = append(cols, fmt.Sprintf("%v", entry[column.Column]))
		}
		if extraZero {
			cols = append(cols, "0")
		}
		rows = append(rows, strings.Join(cols, "^"))
	}
	return strings.Join(rows, "\n"), len(rows), nil
}

func (s *Service) writeExportFile(name, body string) (string, error) {
	targets := []string{}
	if s.pathmgmt != nil {
		if dir := strings.TrimSpace(s.pathmgmt.GetExportDir()); dir != "" {
			targets = append(targets, filepath.Join(dir, name))
		}
	}
	if s.questsDirPath() != "" {
		targets = append(targets, filepath.Join(s.questsDirPath(), filepath.FromSlash(exportRel), name))
	}
	if len(targets) == 0 {
		return "", fmt.Errorf("no export directory")
	}
	var last error
	written := ""
	for _, abs := range targets {
		if err := ensureDir(filepath.Dir(abs)); err != nil {
			last = err
			continue
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			last = err
			continue
		}
		written = abs
	}
	if written == "" {
		if last != nil {
			return "", last
		}
		return "", fmt.Errorf("could not write %s", name)
	}
	return written, nil
}

func (s *Service) ReloadZones(zoneIDs []int) ReloadInfo {
	out := ReloadInfo{Reloaded: []string{}, Warnings: []string{}}
	if s.world == nil {
		out.Note = "World telnet is not wired in this process. Use the reload checkbox so the UI can call Zone Controller apply."
		out.OK = true
		return out
	}
	ids := uniqueInts(zoneIDs)
	if len(ids) == 0 {
		out.Note = "No zones to reload"
		out.OK = true
		return out
	}
	out.WorldUsed = true
	if _, err := s.world.Reload("quests"); err != nil {
		out.Warnings = append(out.Warnings, "world quest reload: "+err.Error())
	} else {
		out.Reloaded = append(out.Reloaded, "world:quests")
	}
	for _, id := range ids {
		short, err := s.zoneShort(id)
		if err != nil {
			out.Warnings = append(out.Warnings, err.Error())
			continue
		}
		if err := s.world.ReloadQuestsForZone(short); err != nil {
			out.Warnings = append(out.Warnings, short+": "+err.Error())
			continue
		}
		out.Reloaded = append(out.Reloaded, short)
	}
	out.OK = true
	if len(out.Reloaded) == 0 {
		out.Note = "World did not reload any zone. Queue refresh/repop from Zone Controller."
	} else {
		out.Note = "Reloaded " + strings.Join(out.Reloaded, ", ") + ". Queue apply/repop if NPCs are already popped."
	}
	return out
}
