package zonecontroller

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeCommands(t *testing.T) {
	got, err := normalizeCommands([]string{"!initdata", "refreshzonedata", "repop", "nope"})
	if err == nil {
		t.Fatal("expected unsupported command")
	}
	got, err = normalizeCommands([]string{"!zc refreshzonedata", "repop", "REPOP"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "refreshzonedata" || got[1] != "repopallstaticbosses" {
		t.Fatalf("got %#v", got)
	}
}

func TestQueueAndApplyWithoutWorld(t *testing.T) {
	root := t.TempDir()
	s := &Service{questsDir: root}
	off := false
	plan := s.ApplyLive(ApplyRequest{
		ZoneIDs:       []int{17, 31},
		QueueCommands: []string{"refreshzonedata", "rebuffzone"},
		RebootEmpty:   &off,
	})
	if !plan.OK || plan.Error != "" {
		t.Fatalf("apply %#v", plan)
	}
	if len(plan.Queued) != 2 {
		t.Fatalf("queued %#v", plan.Queued)
	}
	if len(plan.Commands) != 2 || plan.Commands[0] != "!initdata 17" {
		t.Fatalf("fallback %#v", plan.Commands)
	}
	raw, err := os.ReadFile(filepath.Join(root, "global", "ultimatedata", "_spire_commands", "17.json"))
	if err != nil {
		t.Fatal(err)
	}
	var file commandFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	if file.ZoneID != 17 || len(file.Commands) != 2 || file.Commands[1] != "rebuffzone" {
		t.Fatalf("file %#v", file)
	}
	live := s.LiveStatus([]int{17})
	if live.WorldOK || len(live.Zones) != 1 || live.Zones[0].Popped {
		t.Fatalf("live %#v", live)
	}
}
