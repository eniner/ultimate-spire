package lantern

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/EQEmu/spire/internal/http/routes"
	"github.com/labstack/echo/v4"
)

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (e *Controller) Routes() []*routes.Route {
	return []*routes.Route{
		routes.RegisterRoute(http.MethodGet, "zone-editor/lantern/status", e.status, nil),
		routes.RegisterRoute(http.MethodGet, "zone-editor/lantern/zones", e.zones, nil),
		routes.RegisterRoute(http.MethodGet, "zone-editor/lantern/zones/:zone", e.zone, nil),
		routes.RegisterRoute(http.MethodGet, "zone-editor/lantern/zones/:zone/instances", e.instances, nil),
		routes.RegisterRoute(http.MethodGet, "zone-editor/lantern/file", e.file, nil),
	}
}

func (e *Controller) status(c echo.Context) error {
	root, err := resolveRoot()
	if err != nil {
		return c.JSON(http.StatusOK, echo.Map{
			"ok":              false,
			"root":            "",
			"totalZones":      0,
			"zonesWithModels": 0,
			"error":           err.Error(),
		})
	}
	zones, err := listZones(root)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"ok":              true,
		"root":            root,
		"totalZones":      len(zones),
		"zonesWithModels": len(zones),
	})
}

func (e *Controller) zones(c echo.Context) error {
	root, err := resolveRoot()
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	zones, err := listZones(root)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"ok":    true,
		"root":  root,
		"count": len(zones),
		"zones": zones,
	})
}

func (e *Controller) zone(c echo.Context) error {
	root, err := resolveRoot()
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	zone, err := sanitizeZone(c.Param("zone"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	rel := zoneModelRel(zone)
	abs := filepath.Join(root, filepath.FromSlash(rel))
	hasMesh := false
	if info, err := os.Stat(abs); err == nil && info.Mode().IsRegular() {
		hasMesh = true
	}
	instances, err := parseObjectInstances(root, zone)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"ok":            true,
		"root":          root,
		"zone":          zone,
		"hasMesh":       hasMesh,
		"modelRelPath":  rel,
		"instanceCount": len(instances),
	})
}

func (e *Controller) instances(c echo.Context) error {
	root, err := resolveRoot()
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	zone, err := sanitizeZone(c.Param("zone"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	instances, err := parseObjectInstances(root, zone)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"ok":        true,
		"zone":      zone,
		"count":     len(instances),
		"instances": instances,
	})
}

func (e *Controller) file(c echo.Context) error {
	root, err := resolveRoot()
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	full, err := resolveModelPath(root, c.QueryParam("rel"))
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		return c.JSON(status, echo.Map{"error": err.Error()})
	}
	c.Response().Header().Set(echo.HeaderContentType, "model/gltf-binary")
	return c.File(full)
}
