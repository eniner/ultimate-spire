package contentfactory

import (
	"io"
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
		routes.RegisterRoute(http.MethodGet, "admin/content-factory/status", e.status, nil),
		routes.RegisterRoute(http.MethodGet, "admin/content-factory/ids", e.ids, nil),
		routes.RegisterRoute(http.MethodGet, "admin/content-factory/meta", e.meta, nil),
		routes.RegisterRoute(http.MethodGet, "admin/content-factory/search", e.search, nil),
		routes.RegisterRoute(http.MethodGet, "admin/content-factory/runs", e.runs, nil),
		routes.RegisterRoute(http.MethodGet, "admin/content-factory/census", e.census, nil),
		routes.RegisterRoute(http.MethodGet, "admin/content-factory/catalog", e.catalog, nil),
		routes.RegisterRoute(http.MethodGet, "admin/content-factory/merchant", e.merchant, nil),
		routes.RegisterRoute(http.MethodGet, "admin/content-factory/import-zone", e.importZone, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/spells", e.spells, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/items", e.items, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/loot", e.loot, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/attach", e.attach, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/npc-spells", e.npcSpells, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/pipeline", e.pipeline, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/export", e.exportClient, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/reload", e.reload, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/sync-zone", e.syncZone, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/undo", e.undo, nil),
		routes.RegisterRoute(http.MethodGet, "admin/content-factory/draft", e.getDraft, nil),
		routes.RegisterRoute(http.MethodPut, "admin/content-factory/draft", e.putDraft, nil),
		routes.RegisterRoute(http.MethodGet, "admin/content-factory/spell-set", e.getSpellSet, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/spell-set", e.saveSpellSet, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/npcs", e.npcs, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/pawn", e.pawn, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/give", e.give, nil),
		routes.RegisterRoute(http.MethodPost, "admin/content-factory/probe", e.probe, nil),
	}
}

func (e *Controller) status(c echo.Context) error {
	return c.JSON(http.StatusOK, e.svc.Status())
}

func (e *Controller) ids(c echo.Context) error {
	return c.JSON(http.StatusOK, e.svc.IDBoard())
}

func (e *Controller) meta(c echo.Context) error {
	return c.JSON(http.StatusOK, FactoryMeta())
}

func (e *Controller) search(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	return c.JSON(http.StatusOK, e.svc.Search(c.QueryParam("kind"), c.QueryParam("q"), limit))
}

func (e *Controller) runs(c echo.Context) error {
	return c.JSON(http.StatusOK, e.svc.ListRuns())
}

func (e *Controller) census(c echo.Context) error {
	return c.JSON(http.StatusOK, e.svc.Census(firstNonEmpty(c.QueryParam("zone"), c.QueryParam("id"))))
}

func (e *Controller) catalog(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	return c.JSON(http.StatusOK, e.svc.Catalog(c.QueryParam("kind"), limit))
}

func (e *Controller) merchant(c echo.Context) error {
	id, _ := strconv.Atoi(firstNonEmpty(c.QueryParam("id"), c.QueryParam("merchantId")))
	return c.JSON(http.StatusOK, e.svc.Merchant(id))
}

func (e *Controller) importZone(c echo.Context) error {
	return c.JSON(http.StatusOK, e.svc.ImportZone(firstNonEmpty(c.QueryParam("zone"), c.QueryParam("id"))))
}

func (e *Controller) spells(c echo.Context) error {
	req := SpellRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	return writePlan(c, e.svc.RunSpells(req))
}

func (e *Controller) items(c echo.Context) error {
	req := ItemRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	return writePlan(c, e.svc.RunItems(req))
}

func (e *Controller) loot(c echo.Context) error {
	req := LootRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	return writePlan(c, e.svc.RunLoot(req))
}

func (e *Controller) attach(c echo.Context) error {
	req := AttachRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	return writePlan(c, e.svc.RunAttach(req))
}

func (e *Controller) npcSpells(c echo.Context) error {
	req := NpcSpellRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	return writePlan(c, e.svc.RunNpcSpells(req))
}

func (e *Controller) pipeline(c echo.Context) error {
	req := PipelineRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	return writePlan(c, e.svc.RunPipeline(req))
}

func (e *Controller) exportClient(c echo.Context) error {
	req := struct {
		DbStr bool `json:"dbstr"`
	}{}
	_ = c.Bind(&req)
	return c.JSON(http.StatusOK, e.svc.ExportClient(req.DbStr))
}

func (e *Controller) reload(c echo.Context) error {
	req := struct {
		ZoneIDs []int `json:"zoneIds"`
	}{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	return c.JSON(http.StatusOK, e.svc.ReloadZones(req.ZoneIDs))
}

func (e *Controller) syncZone(c echo.Context) error {
	req := struct {
		DryRun  bool  `json:"dryRun"`
		ZoneIDs []int `json:"zoneIds"`
	}{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	plan := emptyPlan("sync", req.DryRun)
	e.svc.SyncZoneItemSpells(req.ZoneIDs, req.DryRun, &plan)
	if plan.Error == "" {
		plan.OK = true
	}
	return writePlan(c, plan)
}

func (e *Controller) getDraft(c echo.Context) error {
	return c.JSON(http.StatusOK, e.svc.LoadDraft())
}

func (e *Controller) putDraft(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid draft"})
	}
	out := e.svc.SaveDraft(body)
	if out.Error != "" {
		return c.JSON(http.StatusBadRequest, out)
	}
	return c.JSON(http.StatusOK, out)
}

func (e *Controller) getSpellSet(c echo.Context) error {
	id, _ := strconv.Atoi(firstNonEmpty(c.QueryParam("id"), c.QueryParam("sourceId")))
	return c.JSON(http.StatusOK, e.svc.GetSpellSet(id))
}

func (e *Controller) saveSpellSet(c echo.Context) error {
	req := SpellSetSaveRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	return writePlan(c, e.svc.SaveSpellSet(req))
}

func (e *Controller) npcs(c echo.Context) error {
	req := NpcCloneRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	return writePlan(c, e.svc.RunNpcClones(req))
}

func (e *Controller) pawn(c echo.Context) error {
	req := PawnRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	return writePlan(c, e.svc.RunPawn(req))
}

func (e *Controller) give(c echo.Context) error {
	req := GiveRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	return writePlan(c, e.svc.RunGive(req))
}

func (e *Controller) probe(c echo.Context) error {
	req := ProbeRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	return c.JSON(http.StatusOK, e.svc.Probe(req))
}

func (e *Controller) undo(c echo.Context) error {
	req := UndoRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	return writePlan(c, e.svc.Undo(req))
}

func writePlan(c echo.Context, plan Plan) error {
	code := http.StatusOK
	if plan.Error != "" {
		code = http.StatusBadRequest
	}
	return c.JSON(code, plan)
}
