package zonecontroller

import (
	"os"
	"testing"
)

func TestDryRunCreateAgainstLocalUlt(t *testing.T) {
	quests := os.Getenv("SPIRE_QUESTS_ROOT")
	if quests == "" {
		if server := os.Getenv("EQ_SERVER"); server != "" {
			quests = server + `\quests`
		}
	}
	if quests == "" {
		t.Skip("set SPIRE_QUESTS_ROOT or EQ_SERVER to run this live dry-run")
	}
	if _, err := os.Stat(quests + `\global\ultimatedata\templates\_mob.json`); err != nil {
		t.Skip("local ultimatedata templates not present")
	}
	s := &Service{questsDir: quests}
	dry := true
	plan := s.CreateZones(CreateRequest{
		ZoneIDs:   []int{99991, 99992},
		Source:    "clone",
		CloneFrom: 17,
		Name:      "Test Clone",
		Overwrite: false,
		DryRun:    &dry,
	})
	if !plan.OK || plan.Error != "" {
		t.Fatalf("%#v", plan)
	}
	if !plan.DryRun {
		t.Fatal("expected dry run")
	}
	if len(plan.Writes) < 6 {
		t.Fatalf("writes %#v", plan.Writes)
	}
	plan2 := s.ApplyTier(TierRequest{
		ZoneIDs:  []int{17},
		FromZone: 31,
		CopyLoot: true,
		DryRun:   &dry,
	})
	if !plan2.OK {
		t.Fatalf("tier %#v", plan2)
	}
}
