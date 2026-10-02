package serverfiles

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EQEmu/spire/internal/auditlog"
	"github.com/EQEmu/spire/internal/eqemuserverconfig"
	"github.com/EQEmu/spire/internal/http/routes"
	"github.com/EQEmu/spire/internal/pathmgmt"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	svc      *Service
	auditLog *auditlog.UserEvent
}

func NewController(pathmgmt *pathmgmt.PathManagement, config *eqemuserverconfig.Config, auditLog *auditlog.UserEvent) *Controller {
	return &Controller{svc: NewService(pathmgmt, config), auditLog: auditLog}
}

func (e *Controller) Routes() []*routes.Route {
	return []*routes.Route{
		routes.RegisterRoute(http.MethodGet, "admin/server-files/status", e.status, nil),
		routes.RegisterRoute(http.MethodPost, "admin/server-files/connect", e.connect, nil),
		routes.RegisterRoute(http.MethodGet, "admin/server-files/list", e.list, nil),
		routes.RegisterRoute(http.MethodGet, "admin/server-files/file", e.getFile, nil),
		routes.RegisterRoute(http.MethodPut, "admin/server-files/file", e.saveFile, nil),
		routes.RegisterRoute(http.MethodPost, "admin/server-files/file", e.createFile, nil),
		routes.RegisterRoute(http.MethodGet, "admin/server-files/search", e.search, nil),
	}
}

func (e *Controller) status(c echo.Context) error {
	conn := e.svc.currentConnection()
	detected := e.svc.detected()
	ok := conn.Root != ""
	var errMsg string
	if ok {
		if _, err := e.svc.probe(conn.Source, conn.Root); err != nil {
			ok = false
			errMsg = err.Error()
		}
	}
	return c.JSON(http.StatusOK, echo.Map{
		"ok":        ok,
		"source":    conn.Source,
		"root":      conn.Root,
		"writable":  conn.Writable && ok,
		"error":     errMsg,
		"detected":  detected,
		"connected": ok,
	})
}

type connectRequest struct {
	Source string `json:"source"`
	Root   string `json:"root"`
	Save   *bool  `json:"save"`
}

func (e *Controller) connect(c echo.Context) error {
	req := new(connectRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}
	conn, err := e.svc.probe(req.Source, req.Root)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	shouldSave := req.Save == nil || *req.Save
	if shouldSave {
		if err := e.svc.saveConnection(conn.Source, conn.Root); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
	}
	return c.JSON(http.StatusOK, echo.Map{
		"ok":       true,
		"source":   conn.Source,
		"root":     conn.Root,
		"writable": conn.Writable,
		"saved":    shouldSave,
	})
}

func (e *Controller) list(c echo.Context) error {
	conn, err := e.live()
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	rel := c.QueryParam("path")
	var entries []Entry
	if conn.Source == SourceHTTP {
		entries, err = e.svc.listHTTP(conn.Root, rel)
	} else {
		entries, err = e.svc.listLocal(conn.Root, rel)
	}
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"ok":       true,
		"source":   conn.Source,
		"root":     conn.Root,
		"path":     rel,
		"writable": conn.Writable,
		"entries":  entries,
	})
}

func (e *Controller) getFile(c echo.Context) error {
	conn, err := e.live()
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	rel := c.QueryParam("path")
	var body string
	if conn.Source == SourceHTTP {
		body, err = e.svc.readHTTP(conn.Root, rel)
	} else {
		body, err = e.svc.readLocal(conn.Root, rel)
	}
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"ok":       true,
		"path":     rel,
		"writable": conn.Writable,
		"content":  body,
		"hash":     fileHash(body),
	})
}

type fileBody struct {
	Path         string `json:"path"`
	Content      string `json:"content"`
	ExpectedHash string `json:"expectedHash"`
}

func (e *Controller) saveFile(c echo.Context) error {
	return e.writeFile(c, false)
}

func (e *Controller) createFile(c echo.Context) error {
	return e.writeFile(c, true)
}

func (e *Controller) writeFile(c echo.Context, create bool) error {
	conn, err := e.live()
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	if !conn.Writable || conn.Source != SourceLocal {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "remote HTTP sources are read-only; mount the folder or run Spire on that host to save"})
	}
	req := new(fileBody)
	if err := c.Bind(req); err != nil || strings.TrimSpace(req.Path) == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "path is required"})
	}
	if err := e.svc.writeLocal(conn.Root, req.Path, req.Content, req.ExpectedHash, create); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errStaleFile) {
			status = http.StatusConflict
		}
		return c.JSON(status, echo.Map{"error": err.Error()})
	}
	if e.auditLog != nil {
		action := "UPDATE"
		if create {
			action = "CREATE"
		}
		e.auditLog.LogUserEvent(c, action, fmt.Sprintf("Server files %s [%s]", strings.ToLower(action), req.Path))
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true, "path": req.Path, "hash": fileHash(req.Content)})
}

func (e *Controller) search(c echo.Context) error {
	conn, err := e.live()
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	if conn.Source != SourceLocal {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "search is only available for local folders"})
	}
	entries, err := e.svc.searchLocal(conn.Root, c.QueryParam("path"), c.QueryParam("q"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true, "entries": entries})
}

func (e *Controller) live() (Connection, error) {
	conn := e.svc.currentConnection()
	if conn.Root == "" {
		return Connection{}, errors.New("no server files folder is connected")
	}
	return e.svc.probe(conn.Source, conn.Root)
}
