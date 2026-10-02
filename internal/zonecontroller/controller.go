package zonecontroller

import (
	"net/http"
	"strconv"

	"github.com/EQEmu/spire/internal/database"
	"github.com/EQEmu/spire/internal/eqemuserver"
	"github.com/EQEmu/spire/internal/http/routes"
	"github.com/EQEmu/spire/internal/pathmgmt"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	svc *Service
}

func NewController(pathmgmt *pathmgmt.PathManagement, db *database.Resolver, world *eqemuserver.Client) *Controller {
	return &Controller{svc: NewService(pathmgmt, db, world)}
}

func (e *Controller) Routes() []*routes.Route {
	return []*routes.Route{
		routes.RegisterRoute(http.MethodGet, "admin/zone-controller/status", e.status, nil),
		routes.RegisterRoute(http.MethodGet, "admin/zone-controller/zones", e.list, nil),
		routes.RegisterRoute(http.MethodGet, "admin/zone-controller/zones/:id", e.get, nil),
		routes.RegisterRoute(http.MethodGet, "admin/zone-controller/systems", e.systems, nil),
		routes.RegisterRoute(http.MethodPost, "admin/zone-controller/create", e.create, nil),
		routes.RegisterRoute(http.MethodPost, "admin/zone-controller/tier", e.tier, nil),
		routes.RegisterRoute(http.MethodPost, "admin/zone-controller/mobs", e.mobs, nil),
		routes.RegisterRoute(http.MethodGet, "admin/zone-controller/recipes", e.listRecipes, nil),
		routes.RegisterRoute(http.MethodPost, "admin/zone-controller/recipes", e.saveRecipe, nil),
		routes.RegisterRoute(http.MethodDelete, "admin/zone-controller/recipes/:id", e.deleteRecipe, nil),
		routes.RegisterRoute(http.MethodGet, "admin/zone-controller/recipes/from/:id", e.recipeFromZone, nil),
		routes.RegisterRoute(http.MethodPost, "admin/zone-controller/factory", e.factory, nil),
		routes.RegisterRoute(http.MethodPost, "admin/zone-controller/classify", e.classify, nil),
		routes.RegisterRoute(http.MethodGet, "admin/zone-controller/validate", e.validate, nil),
		routes.RegisterRoute(http.MethodGet, "admin/zone-controller/live", e.live, nil),
		routes.RegisterRoute(http.MethodPost, "admin/zone-controller/apply", e.apply, nil),
	}
}

func (e *Controller) status(c echo.Context) error {
	return c.JSON(http.StatusOK, e.svc.Status())
}

func (e *Controller) list(c echo.Context) error {
	zones, err := e.svc.ListZones()
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}
	st := e.svc.Status()
	return c.JSON(http.StatusOK, echo.Map{
		"ok":        st.OK,
		"error":     st.Error,
		"questsDir": st.QuestsDir,
		"dataDir":   st.DataDir,
		"dataRel":   st.DataRel,
		"scriptRel": st.ScriptRel,
		"zones":     zones,
	})
}

func (e *Controller) get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid zone id"})
	}
	detail, err := e.svc.GetZone(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, detail)
}

func (e *Controller) systems(c echo.Context) error {
	sys, err := e.svc.Systems()
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, sys)
}

func (e *Controller) create(c echo.Context) error {
	req := CreateRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	plan := e.svc.CreateZones(req)
	return writePlanResponse(c, plan)
}

func (e *Controller) tier(c echo.Context) error {
	req := TierRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	plan := e.svc.ApplyTier(req)
	return writePlanResponse(c, plan)
}

func (e *Controller) mobs(c echo.Context) error {
	req := MobsRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	plan := e.svc.AddMobs(req)
	return writePlanResponse(c, plan)
}

func writePlanResponse(c echo.Context, plan WritePlan) error {
	code := http.StatusOK
	if plan.Error != "" {
		code = http.StatusBadRequest
	}
	return c.JSON(code, plan)
}

func (e *Controller) listRecipes(c echo.Context) error {
	return c.JSON(http.StatusOK, e.svc.ListRecipes())
}

func (e *Controller) saveRecipe(c echo.Context) error {
	req := Recipe{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	plan := e.svc.SaveRecipe(req, false)
	return writePlanResponse(c, plan.WritePlan)
}

func (e *Controller) deleteRecipe(c echo.Context) error {
	plan := e.svc.DeleteRecipe(c.Param("id"), false)
	return writePlanResponse(c, plan.WritePlan)
}

func (e *Controller) recipeFromZone(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid zone id"})
	}
	recipe, err := e.svc.RecipeFromZone(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, recipe)
}

func (e *Controller) factory(c echo.Context) error {
	req := FactoryRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	plan := e.svc.RunFactory(req)
	code := http.StatusOK
	if plan.Error != "" {
		code = http.StatusBadRequest
	}
	return c.JSON(code, plan)
}

func (e *Controller) classify(c echo.Context) error {
	req := ClassifyRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	plan := e.svc.ClassifyZones(req)
	code := http.StatusOK
	if plan.Error != "" {
		code = http.StatusBadRequest
	}
	return c.JSON(code, plan)
}

func (e *Controller) validate(c echo.Context) error {
	return c.JSON(http.StatusOK, e.svc.ValidateZones(queryZoneIDs(c.QueryParam("zones"))))
}

func (e *Controller) live(c echo.Context) error {
	return c.JSON(http.StatusOK, e.svc.LiveStatus(queryZoneIDs(c.QueryParam("zones"))))
}

func (e *Controller) apply(c echo.Context) error {
	req := ApplyRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	plan := e.svc.ApplyLive(req)
	code := http.StatusOK
	if plan.Error != "" {
		code = http.StatusBadRequest
	}
	return c.JSON(code, plan)
}

func queryZoneIDs(raw string) []int {
	ids := []int{}
	for _, part := range splitIDs(raw) {
		if n, err := strconv.Atoi(part); err == nil && n > 0 {
			ids = append(ids, n)
		}
	}
	return ids
}

func splitIDs(raw string) []string {
	out := []string{}
	cur := ""
	for _, r := range raw {
		if r == ',' || r == ' ' || r == '\n' || r == '\t' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
