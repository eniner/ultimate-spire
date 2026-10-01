package peqraw

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/EQEmu/spire/internal/auditlog"
	"github.com/EQEmu/spire/internal/database"
	"github.com/EQEmu/spire/internal/http/routes"
	"github.com/EQEmu/spire/internal/models"
	"github.com/labstack/echo/v4"
)

var identRe = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type Controller struct {
	db       *database.Resolver
	auditLog *auditlog.UserEvent
}

func NewController(db *database.Resolver, auditLog *auditlog.UserEvent) *Controller {
	return &Controller{db: db, auditLog: auditLog}
}

func (e *Controller) Routes() []*routes.Route {
	return []*routes.Route{
		routes.RegisterRoute(http.MethodGet, "peq-raw/tables", e.listTables, nil),
		routes.RegisterRoute(http.MethodGet, "peq-raw/:table/count", e.countRows, nil),
		routes.RegisterRoute(http.MethodGet, "peq-raw/:table", e.listRows, nil),
		routes.RegisterRoute(http.MethodPut, "peq-raw/:table", e.createRow, nil),
		routes.RegisterRoute(http.MethodPatch, "peq-raw/:table", e.updateRow, nil),
		routes.RegisterRoute(http.MethodDelete, "peq-raw/:table", e.deleteRow, nil),
	}
}

type columnInfo struct {
	Name string
	Key  string
	Type string
}

func allowedTable(name string) bool {
	n := strings.ToLower(name)
	return identRe.MatchString(n) && (n == "mercs" || strings.HasPrefix(n, "merc_"))
}

func quoteIdent(name string) (string, error) {
	if !identRe.MatchString(name) {
		return "", fmt.Errorf("invalid identifier")
	}
	return "`" + name + "`", nil
}

func (e *Controller) listTables(c echo.Context) error {
	db := e.db.Get(models.CharacterDatum{}, c)
	var names []string
	err := db.Raw(`
		SELECT TABLE_NAME
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
		  AND (TABLE_NAME = 'mercs' OR TABLE_NAME LIKE 'merc_%')
		ORDER BY TABLE_NAME
	`).Scan(&names).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	filtered := make([]string, 0, len(names))
	for _, name := range names {
		if allowedTable(name) {
			filtered = append(filtered, name)
		}
	}
	return c.JSON(http.StatusOK, echo.Map{"tables": filtered})
}

type showColumn struct {
	Field string `gorm:"column:Field"`
	Type  string `gorm:"column:Type"`
	Key   string `gorm:"column:Key"`
}

func (e *Controller) columns(c echo.Context, table string) ([]columnInfo, error) {
	qTable, err := quoteIdent(table)
	if err != nil {
		return nil, err
	}
	db := e.db.Get(models.CharacterDatum{}, c)
	var raw []showColumn
	if err := db.Raw("SHOW COLUMNS FROM " + qTable).Scan(&raw).Error; err != nil {
		return nil, err
	}
	cols := make([]columnInfo, 0, len(raw))
	for _, row := range raw {
		col := columnInfo{
			Name: row.Field,
			Key:  strings.ToUpper(row.Key),
			Type: strings.ToLower(row.Type),
		}
		if identRe.MatchString(col.Name) {
			cols = append(cols, col)
		}
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("no columns")
	}
	return cols, nil
}

func primaryCols(cols []columnInfo) []string {
	var pks []string
	for _, col := range cols {
		if col.Key == "PRI" {
			pks = append(pks, col.Name)
		}
	}
	if len(pks) == 0 && len(cols) > 0 {
		pks = []string{cols[0].Name}
	}
	return pks
}

func stringifyRow(row map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(row))
	for k, v := range row {
		switch t := v.(type) {
		case []byte:
			out[k] = string(t)
		default:
			out[k] = v
		}
	}
	return out
}

func (e *Controller) listRows(c echo.Context) error {
	table := c.Param("table")
	if !allowedTable(table) {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "table is not allowed"})
	}
	cols, err := e.columns(c, table)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	qTable, _ := quoteIdent(table)
	limit := 100
	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	page := 1
	if v := c.QueryParam("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	offset := (page - 1) * limit
	search := strings.TrimSpace(c.QueryParam("search"))
	whereSQL := ""
	args := []interface{}{}
	if search != "" {
		parts := []string{}
		for _, col := range cols {
			qcol, err := quoteIdent(col.Name)
			if err != nil {
				continue
			}
			if strings.Contains(col.Type, "char") || strings.Contains(col.Type, "text") {
				parts = append(parts, qcol+" LIKE ?")
				args = append(args, "%"+search+"%")
			} else if _, err := strconv.Atoi(search); err == nil && (col.Key == "PRI" || strings.Contains(col.Type, "int")) {
				parts = append(parts, qcol+" = ?")
				args = append(args, search)
			}
		}
		if len(parts) > 0 {
			whereSQL = " WHERE " + strings.Join(parts, " OR ")
		}
	}
	sql := fmt.Sprintf("SELECT * FROM %s%s LIMIT ? OFFSET ?", qTable, whereSQL)
	args = append(args, limit, offset)
	db := e.db.Get(models.CharacterDatum{}, c)
	var rows []map[string]interface{}
	if err := db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, stringifyRow(row))
	}
	return c.JSON(http.StatusOK, echo.Map{
		"rows":       out,
		"primaryKey": primaryCols(cols),
		"columns":    colNames(cols),
	})
}

func colNames(cols []columnInfo) []string {
	names := make([]string, 0, len(cols))
	for _, col := range cols {
		names = append(names, col.Name)
	}
	return names
}

func (e *Controller) countRows(c echo.Context) error {
	table := c.Param("table")
	if !allowedTable(table) {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "table is not allowed"})
	}
	qTable, err := quoteIdent(table)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	db := e.db.Get(models.CharacterDatum{}, c)
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM " + qTable).Scan(&count).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"count": count})
}

func (e *Controller) createRow(c echo.Context) error {
	table := c.Param("table")
	if !allowedTable(table) {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "table is not allowed"})
	}
	cols, err := e.columns(c, table)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	body := map[string]interface{}{}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	qTable, _ := quoteIdent(table)
	allowed := map[string]bool{}
	for _, col := range cols {
		allowed[col.Name] = true
	}
	var fields []string
	var placeholders []string
	var args []interface{}
	for key, val := range body {
		if !allowed[key] {
			continue
		}
		qcol, err := quoteIdent(key)
		if err != nil {
			continue
		}
		fields = append(fields, qcol)
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}
	if len(fields) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "no valid columns"})
	}
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", qTable, strings.Join(fields, ","), strings.Join(placeholders, ","))
	db := e.db.Get(models.CharacterDatum{}, c)
	if err := db.Exec(sql, args...).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	if e.db.GetSpireDb() != nil {
		e.auditLog.LogUserEvent(c, "CREATE", fmt.Sprintf("Created peq-raw [%s]", table))
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}

func (e *Controller) updateRow(c echo.Context) error {
	return e.mutateRow(c, false)
}

func (e *Controller) deleteRow(c echo.Context) error {
	return e.mutateRow(c, true)
}

func (e *Controller) mutateRow(c echo.Context, del bool) error {
	table := c.Param("table")
	if !allowedTable(table) {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "table is not allowed"})
	}
	cols, err := e.columns(c, table)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	pks := primaryCols(cols)
	qTable, _ := quoteIdent(table)
	body := map[string]interface{}{}
	if err := c.Bind(&body); err != nil && !del {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	allowed := map[string]bool{}
	for _, col := range cols {
		allowed[col.Name] = true
	}
	var where []string
	var args []interface{}
	for _, pk := range pks {
		val := c.QueryParam(pk)
		if val == "" {
			if bodyVal, ok := body[pk]; ok {
				val = fmt.Sprintf("%v", bodyVal)
			}
		}
		if val == "" {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "missing primary key " + pk})
		}
		qcol, err := quoteIdent(pk)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		where = append(where, qcol+" = ?")
		args = append(args, val)
	}
	db := e.db.Get(models.CharacterDatum{}, c)
	if del {
		sql := fmt.Sprintf("DELETE FROM %s WHERE %s", qTable, strings.Join(where, " AND "))
		if err := db.Exec(sql, args...).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
		if e.db.GetSpireDb() != nil {
			e.auditLog.LogUserEvent(c, "DELETE", fmt.Sprintf("Deleted peq-raw [%s]", table))
		}
		return c.JSON(http.StatusOK, echo.Map{"ok": true})
	}
	var sets []string
	var setArgs []interface{}
	for key, val := range body {
		skip := false
		for _, pk := range pks {
			if key == pk {
				skip = true
				break
			}
		}
		if skip || !allowed[key] {
			continue
		}
		qcol, err := quoteIdent(key)
		if err != nil {
			continue
		}
		sets = append(sets, qcol+" = ?")
		setArgs = append(setArgs, val)
	}
	if len(sets) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "no fields to update"})
	}
	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s", qTable, strings.Join(sets, ","), strings.Join(where, " AND "))
	setArgs = append(setArgs, args...)
	if err := db.Exec(sql, setArgs...).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	if e.db.GetSpireDb() != nil {
		e.auditLog.LogUserEvent(c, "UPDATE", fmt.Sprintf("Updated peq-raw [%s]", table))
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}
