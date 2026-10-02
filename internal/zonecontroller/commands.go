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

type QueuedCommand struct {
	ZoneID   int      `json:"zoneId"`
	Short    string   `json:"short,omitempty"`
	ID       string   `json:"id"`
	Commands []string `json:"commands"`
	RelPath  string   `json:"relPath"`
}

type commandFile struct {
	ID        string   `json:"id"`
	ZoneID    int      `json:"zoneId"`
	Commands  []string `json:"commands"`
	QueuedAt  int64    `json:"queuedAt"`
}

var allowedZCCommands = map[string]string{
	"refreshzonedata":      "refreshzonedata",
	"initdata":             "refreshzonedata",
	"repopallstaticbosses": "repopallstaticbosses",
	"repop":                "repopallstaticbosses",
	"depopallstaticbosses": "depopallstaticbosses",
	"depop":                "depopallstaticbosses",
	"rebuffzone":           "rebuffzone",
	"reloadzone":           "rebuffzone",
	"applyallchanges":      "applyallchanges",
	"resetallchanges":      "resetallchanges",
	"saveallchanges":       "saveallchanges",
}

func normalizeCommands(raw []string) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	for _, item := range raw {
		key := strings.ToLower(strings.TrimSpace(item))
		key = strings.TrimPrefix(key, "!zc ")
		key = strings.TrimPrefix(key, "#zc ")
		key = strings.TrimPrefix(key, "!")
		key = strings.TrimPrefix(key, "#")
		if key == "" {
			continue
		}
		mapped, ok := allowedZCCommands[key]
		if !ok {
			return nil, fmt.Errorf("unsupported zone controller command %q", item)
		}
		if seen[mapped] {
			continue
		}
		seen[mapped] = true
		out = append(out, mapped)
	}
	return out, nil
}

func (s *Service) commandDir() string {
	return filepath.Join(s.questsDirPath(), filepath.FromSlash(commandRel))
}

func (s *Service) queueCommands(ids []int, commands []string) ([]QueuedCommand, []string) {
	out := []QueuedCommand{}
	warns := []string{}
	if len(ids) == 0 || len(commands) == 0 {
		return out, warns
	}
	if err := os.MkdirAll(s.commandDir(), 0o755); err != nil {
		return out, []string{"command queue: " + err.Error()}
	}
	now := time.Now().Unix()
	for _, id := range ids {
		short := ""
		if info, err := s.zoneInfo(id); err == nil {
			short = info.short
		}
		row := commandFile{
			ID:       fmt.Sprintf("%d-%d", id, now),
			ZoneID:   id,
			Commands: commands,
			QueuedAt: now,
		}
		name := strconv.Itoa(id) + ".json"
		rel := commandRel + "/" + name
		path := filepath.Join(s.commandDir(), name)
		body, err := json.MarshalIndent(row, "", "  ")
		if err != nil {
			warns = append(warns, fmt.Sprintf("zone %d: %v", id, err))
			continue
		}
		if err := os.WriteFile(path, body, 0o644); err != nil {
			warns = append(warns, fmt.Sprintf("zone %d: %v", id, err))
			continue
		}
		out = append(out, QueuedCommand{ZoneID: id, Short: short, ID: row.ID, Commands: append([]string{}, commands...), RelPath: rel})
	}
	return out, warns
}
