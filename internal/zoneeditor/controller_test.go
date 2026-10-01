package zoneeditor

import (
	"testing"
)

func TestValidateSaveRequest(t *testing.T) {
	x := float32(1)
	validChange := PlacementChange{Table: "spawn2", ID: 10, X: &x}

	t.Run("requires zone", func(t *testing.T) {
		err := validateSaveRequest(&SavePlacementsRequest{
			Changes: []PlacementChange{validChange},
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("requires changes", func(t *testing.T) {
		err := validateSaveRequest(&SavePlacementsRequest{Zone: "poknowledge"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects unknown table", func(t *testing.T) {
		err := validateSaveRequest(&SavePlacementsRequest{
			Zone:    "poknowledge",
			Changes: []PlacementChange{{Table: "doors", ID: 1, X: &x}},
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("accepts spawn2", func(t *testing.T) {
		err := validateSaveRequest(&SavePlacementsRequest{
			Zone:    "poknowledge",
			Version: 0,
			Changes: []PlacementChange{validChange},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
