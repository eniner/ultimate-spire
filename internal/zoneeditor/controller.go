package zoneeditor

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EQEmu/spire/internal/auditlog"
	"github.com/EQEmu/spire/internal/database"
	"github.com/EQEmu/spire/internal/http/routes"
	"github.com/EQEmu/spire/internal/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

const maxPlacementChanges = 250

type Controller struct {
	db       *database.Resolver
	auditLog *auditlog.UserEvent
}

func NewController(db *database.Resolver, auditLog *auditlog.UserEvent) *Controller {
	return &Controller{
		db:       db,
		auditLog: auditLog,
	}
}

func (e *Controller) Routes() []*routes.Route {
	return []*routes.Route{
		routes.RegisterRoute(http.MethodPost, "zone-editor/placements", e.savePlacements, nil),
	}
}

type PlacementChange struct {
	Table   string   `json:"table"`
	ID      int      `json:"id"`
	X       *float32 `json:"x"`
	Y       *float32 `json:"y"`
	Z       *float32 `json:"z"`
	Heading *float32 `json:"heading"`
}

type SavePlacementsRequest struct {
	Zone    string            `json:"zone"`
	Version int16             `json:"version"`
	Changes []PlacementChange `json:"changes"`
}

func validateSaveRequest(req *SavePlacementsRequest) error {
	if req == nil {
		return fmt.Errorf("missing request body")
	}
	zone := strings.TrimSpace(req.Zone)
	if zone == "" {
		return fmt.Errorf("zone is required")
	}
	if len(req.Changes) == 0 {
		return fmt.Errorf("changes is required")
	}
	if len(req.Changes) > maxPlacementChanges {
		return fmt.Errorf("too many changes (max %d)", maxPlacementChanges)
	}
	for i, change := range req.Changes {
		if change.ID <= 0 {
			return fmt.Errorf("changes[%d]: id is required", i)
		}
		table := strings.ToLower(strings.TrimSpace(change.Table))
		if table != "spawn2" {
			return fmt.Errorf("changes[%d]: unsupported table %q (phase 1 is spawn2 only)", i, change.Table)
		}
		if change.X == nil && change.Y == nil && change.Z == nil && change.Heading == nil {
			return fmt.Errorf("changes[%d]: no placement fields to update", i)
		}
	}
	return nil
}

func (e *Controller) savePlacements(c echo.Context) error {
	req := new(SavePlacementsRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := validateSaveRequest(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	req.Zone = strings.TrimSpace(req.Zone)
	db := e.db.Get(models.Spawn2{}, c)
	updatedIDs := make([]int, 0, len(req.Changes))

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, change := range req.Changes {
			var row models.Spawn2
			err := tx.Where("id = ? AND zone = ? AND version = ?", change.ID, req.Zone, req.Version).
				First(&row).Error
			if err != nil {
				return fmt.Errorf("spawn2 id %d not found in zone %s version %d", change.ID, req.Zone, req.Version)
			}

			updates := map[string]interface{}{}
			if change.X != nil {
				updates["x"] = *change.X
			}
			if change.Y != nil {
				updates["y"] = *change.Y
			}
			if change.Z != nil {
				updates["z"] = *change.Z
			}
			if change.Heading != nil {
				updates["heading"] = *change.Heading
			}

			if err := tx.Model(&row).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed updating spawn2 id %d: %v", change.ID, err)
			}
			updatedIDs = append(updatedIDs, change.ID)
		}
		return nil
	})
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		return c.JSON(status, echo.Map{"error": err.Error()})
	}

	if e.auditLog != nil && len(updatedIDs) > 0 {
		event := fmt.Sprintf(
			"Zone editor spawn2 placements zone=%s version=%d count=%d ids=%v",
			req.Zone,
			req.Version,
			len(updatedIDs),
			updatedIDs,
		)
		e.auditLog.LogUserEvent(c, "UPDATE", event)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": echo.Map{
			"updated": len(updatedIDs),
			"ids":     updatedIDs,
		},
	})
}
