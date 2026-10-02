package ultsystems

import (
	"net/http"
	"strings"

	"github.com/EQEmu/spire/internal/database"
	"github.com/EQEmu/spire/internal/http/routes"
	"github.com/EQEmu/spire/internal/pathmgmt"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	svc *Service
}

func NewController(pathmgmt *pathmgmt.PathManagement, db *database.Resolver) *Controller {
	return &Controller{svc: NewService(pathmgmt, db)}
}

func (e *Controller) Routes() []*routes.Route {
	return []*routes.Route{
		routes.RegisterRoute(http.MethodGet, "admin/ultimate-systems/status", e.status, nil),
		routes.RegisterRoute(http.MethodGet, "admin/ultimate-systems/catalog", e.catalog, nil),
		routes.RegisterRoute(http.MethodGet, "admin/ultimate-systems/talents/:id", e.talent, nil),
		routes.RegisterRoute(http.MethodGet, "admin/ultimate-systems/traits", e.traits, nil),
		routes.RegisterRoute(http.MethodGet, "admin/ultimate-systems/runewords", e.runewords, nil),
		routes.RegisterRoute(http.MethodPost, "admin/ultimate-systems/write", e.write, nil),
	}
}

func (e *Controller) status(c echo.Context) error {
	return c.JSON(http.StatusOK, e.svc.Status())
}

func (e *Controller) catalog(c echo.Context) error {
	payload, err := e.svc.Catalog()
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, payload)
}

func (e *Controller) talent(c echo.Context) error {
	row, err := e.svc.Talent(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, row)
}

func (e *Controller) traits(c echo.Context) error {
	payload, err := e.svc.Traits()
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, payload)
}

func (e *Controller) runewords(c echo.Context) error {
	return c.JSON(http.StatusOK, e.svc.Runewords())
}

func (e *Controller) write(c echo.Context) error {
	req := WriteRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	if strings.TrimSpace(req.Kind) == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "kind is required"})
	}
	plan := e.svc.Write(req)
	code := http.StatusOK
	if plan.Error != "" {
		code = http.StatusBadRequest
	}
	return c.JSON(code, plan)
}
