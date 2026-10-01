package website

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/EQEmu/spire/internal/auditlog"
	"github.com/EQEmu/spire/internal/database"
	"github.com/EQEmu/spire/internal/http/routes"
	"github.com/EQEmu/spire/internal/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var knownTables = []string{
	"web_users",
	"web_user_favorites",
	"web_eq_account_links",
	"web_eq_account_link_tokens",
	"web_eq_character_links",
	"web_user_active_character",
	"web_admin_audit_log",
}

type Controller struct {
	db       *database.Resolver
	auditLog *auditlog.UserEvent
}

func NewController(db *database.Resolver, auditLog *auditlog.UserEvent) *Controller {
	return &Controller{db: db, auditLog: auditLog}
}

func (e *Controller) Routes() []*routes.Route {
	return []*routes.Route{
		routes.RegisterRoute(http.MethodGet, "admin/website/status", e.status, nil),
		routes.RegisterRoute(http.MethodGet, "admin/website/users", e.listUsers, nil),
		routes.RegisterRoute(http.MethodPatch, "admin/website/users/:id/role", e.updateUserRole, nil),
		routes.RegisterRoute(http.MethodGet, "admin/website/account-links", e.listAccountLinks, nil),
		routes.RegisterRoute(http.MethodPost, "admin/website/account-links", e.createAccountLink, nil),
		routes.RegisterRoute(http.MethodDelete, "admin/website/account-links/:id", e.deleteAccountLink, nil),
		routes.RegisterRoute(http.MethodGet, "admin/website/character-links", e.listCharacterLinks, nil),
		routes.RegisterRoute(http.MethodPost, "admin/website/character-links", e.createCharacterLink, nil),
		routes.RegisterRoute(http.MethodDelete, "admin/website/character-links/:id", e.deleteCharacterLink, nil),
		routes.RegisterRoute(http.MethodGet, "admin/website/eq-accounts", e.searchAccounts, nil),
		routes.RegisterRoute(http.MethodGet, "admin/website/eq-characters", e.searchCharacters, nil),
	}
}

func (e *Controller) conn(c echo.Context) *gorm.DB {
	return e.db.Get(models.CharacterDatum{}, c)
}

func (e *Controller) present(c echo.Context, name string) bool {
	var n int
	err := e.conn(c).Raw(`
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_name = ?
	`, name).Scan(&n).Error
	return err == nil && n > 0
}

func (e *Controller) count(c echo.Context, name string) int64 {
	var n int64
	if !e.present(c, name) {
		return 0
	}
	_ = e.conn(c).Raw("SELECT COUNT(*) FROM `" + name + "`").Scan(&n).Error
	return n
}

func (e *Controller) status(c echo.Context) error {
	tables := make([]echo.Map, 0, len(knownTables)+8)
	for _, name := range knownTables {
		ok := e.present(c, name)
		item := echo.Map{"name": name, "present": ok, "count": int64(0)}
		if ok {
			item["count"] = e.count(c, name)
		}
		tables = append(tables, item)
	}
	var extras []string
	_ = e.conn(c).Raw(`
		SELECT TABLE_NAME
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
		  AND (TABLE_NAME LIKE 'market_%' OR TABLE_NAME LIKE 'achievement_%')
		ORDER BY TABLE_NAME
	`).Scan(&extras).Error
	for _, name := range extras {
		tables = append(tables, echo.Map{"name": name, "present": true, "count": e.count(c, name)})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"ok":           true,
		"linked":       e.present(c, "web_users") && e.present(c, "web_eq_account_links"),
		"website_root": strings.TrimSpace(os.Getenv("SPIRE_WEBSITE_ROOT")),
		"quests_root":  strings.TrimSpace(os.Getenv("SPIRE_QUESTS_ROOT")),
		"tables":       tables,
	})
}

type webUserRow struct {
	ID          uint64 `json:"id"`
	DiscordID   string `json:"discord_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	LastLoginAt string `json:"last_login_at"`
}

func (e *Controller) listUsers(c echo.Context) error {
	if !e.present(c, "web_users") {
		return c.JSON(http.StatusOK, echo.Map{"rows": []webUserRow{}, "missing": true})
	}
	q := strings.TrimSpace(c.QueryParam("q"))
	sql := `
		SELECT id, discord_id, username, display_name, role,
		       IFNULL(DATE_FORMAT(last_login_at, '%Y-%m-%d %H:%i'), '') AS last_login_at
		FROM web_users
	`
	args := []interface{}{}
	if q != "" {
		like := "%" + q + "%"
		sql += " WHERE username LIKE ? OR display_name LIKE ? OR discord_id LIKE ? OR CAST(id AS CHAR) = ?"
		args = append(args, like, like, like, q)
	}
	sql += " ORDER BY updated_at DESC LIMIT 200"
	var rows []webUserRow
	if err := e.conn(c).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"rows": rows})
}

type updateUserRoleReq struct {
	Role string `json:"role"`
}

func (e *Controller) updateUserRole(c echo.Context) error {
	if !e.present(c, "web_users") {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "web_users is not on this database"})
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	req := new(updateUserRoleReq)
	if err := c.Bind(req); err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "id and role are required"})
	}
	role := strings.ToLower(strings.TrimSpace(req.Role))
	switch role {
	case "user", "admin":
	default:
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "role must be user or admin"})
	}
	var current struct {
		ID        uint64
		Role      string
		DiscordID string
		Username  string
	}
	if err := e.conn(c).Raw("SELECT id, role, discord_id, username FROM web_users WHERE id = ?", id).Scan(&current).Error; err != nil || current.ID == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "web user not found"})
	}
	if current.Role == role {
		return c.JSON(http.StatusOK, echo.Map{"ok": true, "id": id, "role": role})
	}
	if current.Role == "admin" && role != "admin" {
		var admins int64
		_ = e.conn(c).Raw("SELECT COUNT(*) FROM web_users WHERE role = 'admin'").Scan(&admins).Error
		if admins <= 1 {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "cannot demote the last admin"})
		}
	}
	if err := e.conn(c).Exec("UPDATE web_users SET role = ? WHERE id = ?", role, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	if e.present(c, "web_admin_audit_log") {
		_ = e.conn(c).Exec(`
			INSERT INTO web_admin_audit_log
				(actor_discord_id, actor_username, actor_display_name, action, target_discord_id, result, details_json)
			VALUES ('spire', 'spire', 'Spire', 'role_update', ?, 'success', ?)
		`, current.DiscordID, fmt.Sprintf("from=%s to=%s user=%s", current.Role, role, current.Username)).Error
	}
	if e.auditLog != nil {
		e.auditLog.LogUserEvent(c, "Website", fmt.Sprintf("Set web user %d (%s) role %s -> %s", id, current.Username, current.Role, role))
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true, "id": id, "role": role})
}

type accountLinkRow struct {
	ID                   uint64 `json:"id"`
	WebUserID            uint64 `json:"web_user_id"`
	Username             string `json:"username"`
	DisplayName          string `json:"display_name"`
	DiscordID            string `json:"discord_id"`
	EqAccountID          uint   `json:"eq_account_id"`
	EqAccountName        string `json:"eq_account_name"`
	LinkedCharacterID    *uint  `json:"linked_character_id"`
	LinkedCharacterName  string `json:"linked_character_name"`
	LinkedBy             string `json:"linked_by"`
	CreatedAt            string `json:"created_at"`
}

func (e *Controller) listAccountLinks(c echo.Context) error {
	if !e.present(c, "web_eq_account_links") {
		return c.JSON(http.StatusOK, echo.Map{"rows": []accountLinkRow{}, "missing": true})
	}
	q := strings.TrimSpace(c.QueryParam("q"))
	sql := `
		SELECT l.id, l.web_user_id, IFNULL(u.username,'') AS username,
		       IFNULL(u.display_name,'') AS display_name, IFNULL(u.discord_id,'') AS discord_id,
		       l.eq_account_id, l.eq_account_name, l.linked_character_id,
		       IFNULL(l.linked_character_name,'') AS linked_character_name,
		       l.linked_by, DATE_FORMAT(l.created_at, '%Y-%m-%d %H:%i') AS created_at
		FROM web_eq_account_links l
		LEFT JOIN web_users u ON u.id = l.web_user_id
	`
	args := []interface{}{}
	if q != "" {
		like := "%" + q + "%"
		sql += ` WHERE l.eq_account_name LIKE ? OR IFNULL(l.linked_character_name,'') LIKE ?
		         OR IFNULL(u.username,'') LIKE ? OR IFNULL(u.display_name,'') LIKE ?
		         OR CAST(l.eq_account_id AS CHAR) = ? OR CAST(l.web_user_id AS CHAR) = ?`
		args = append(args, like, like, like, like, q, q)
	}
	sql += " ORDER BY l.id DESC LIMIT 300"
	var rows []accountLinkRow
	if err := e.conn(c).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"rows": rows})
}

type createAccountLinkReq struct {
	WebUserID   uint64 `json:"web_user_id"`
	EqAccountID uint   `json:"eq_account_id"`
}

func (e *Controller) createAccountLink(c echo.Context) error {
	if !e.present(c, "web_eq_account_links") || !e.present(c, "web_users") {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "website link tables are not on this database yet"})
	}
	req := new(createAccountLinkReq)
	if err := c.Bind(req); err != nil || req.WebUserID == 0 || req.EqAccountID == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "web_user_id and eq_account_id are required"})
	}
	var userCount int64
	if err := e.conn(c).Raw("SELECT COUNT(*) FROM web_users WHERE id = ?", req.WebUserID).Scan(&userCount).Error; err != nil || userCount == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "web user not found"})
	}
	var acc struct {
		ID   uint
		Name string
	}
	if err := e.conn(c).Raw("SELECT id, name FROM account WHERE id = ?", req.EqAccountID).Scan(&acc).Error; err != nil || acc.ID == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "EQ account not found"})
	}
	if err := e.conn(c).Exec(`
		INSERT INTO web_eq_account_links (web_user_id, eq_account_id, eq_account_name, linked_by)
		VALUES (?, ?, ?, 'admin')
	`, req.WebUserID, acc.ID, acc.Name).Error; err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	if e.auditLog != nil {
		e.auditLog.LogUserEvent(c, "Website", fmt.Sprintf("Linked web user %d to account %s (%d)", req.WebUserID, acc.Name, acc.ID))
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}

func (e *Controller) deleteAccountLink(c echo.Context) error {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 || !e.present(c, "web_eq_account_links") {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid link"})
	}
	if err := e.conn(c).Exec("DELETE FROM web_eq_account_links WHERE id = ?", id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	if e.auditLog != nil {
		e.auditLog.LogUserEvent(c, "Website", fmt.Sprintf("Unlinked account link %d", id))
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}

type characterLinkRow struct {
	ID            uint64 `json:"id"`
	WebUserID     uint64 `json:"web_user_id"`
	Username      string `json:"username"`
	DisplayName   string `json:"display_name"`
	EqAccountID   uint   `json:"eq_account_id"`
	CharacterID   uint   `json:"character_id"`
	CharacterName string `json:"character_name"`
	Source        string `json:"source"`
	CreatedAt     string `json:"created_at"`
}

func (e *Controller) listCharacterLinks(c echo.Context) error {
	if !e.present(c, "web_eq_character_links") {
		return c.JSON(http.StatusOK, echo.Map{"rows": []characterLinkRow{}, "missing": true})
	}
	q := strings.TrimSpace(c.QueryParam("q"))
	sql := `
		SELECT l.id, l.web_user_id, IFNULL(u.username,'') AS username,
		       IFNULL(u.display_name,'') AS display_name,
		       l.eq_account_id, l.character_id, l.character_name, l.source,
		       DATE_FORMAT(l.created_at, '%Y-%m-%d %H:%i') AS created_at
		FROM web_eq_character_links l
		LEFT JOIN web_users u ON u.id = l.web_user_id
	`
	args := []interface{}{}
	if q != "" {
		like := "%" + q + "%"
		sql += ` WHERE l.character_name LIKE ? OR IFNULL(u.username,'') LIKE ?
		         OR CAST(l.character_id AS CHAR) = ? OR CAST(l.eq_account_id AS CHAR) = ?`
		args = append(args, like, like, q, q)
	}
	sql += " ORDER BY l.id DESC LIMIT 300"
	var rows []characterLinkRow
	if err := e.conn(c).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"rows": rows})
}

type createCharacterLinkReq struct {
	WebUserID   uint64 `json:"web_user_id"`
	CharacterID uint   `json:"character_id"`
}

func (e *Controller) createCharacterLink(c echo.Context) error {
	if !e.present(c, "web_eq_character_links") || !e.present(c, "web_users") {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "character link tables are not on this database yet"})
	}
	req := new(createCharacterLinkReq)
	if err := c.Bind(req); err != nil || req.WebUserID == 0 || req.CharacterID == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "web_user_id and character_id are required"})
	}
	var userCount int64
	if err := e.conn(c).Raw("SELECT COUNT(*) FROM web_users WHERE id = ?", req.WebUserID).Scan(&userCount).Error; err != nil || userCount == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "web user not found"})
	}
	var ch struct {
		ID        uint
		AccountID int
		Name      string
	}
	if err := e.conn(c).Raw("SELECT id, account_id, name FROM character_data WHERE id = ?", req.CharacterID).Scan(&ch).Error; err != nil || ch.ID == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "character not found"})
	}
	if err := e.conn(c).Exec(`
		INSERT INTO web_eq_character_links (web_user_id, eq_account_id, character_id, character_name, source)
		VALUES (?, ?, ?, ?, 'admin')
	`, req.WebUserID, ch.AccountID, ch.ID, ch.Name).Error; err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	if e.auditLog != nil {
		e.auditLog.LogUserEvent(c, "Website", fmt.Sprintf("Linked web user %d to character %s (%d)", req.WebUserID, ch.Name, ch.ID))
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}

func (e *Controller) deleteCharacterLink(c echo.Context) error {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 || !e.present(c, "web_eq_character_links") {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid link"})
	}
	if err := e.conn(c).Exec("DELETE FROM web_eq_character_links WHERE id = ?", id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}

type eqAccountHit struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (e *Controller) searchAccounts(c echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	if q == "" {
		return c.JSON(http.StatusOK, echo.Map{"rows": []eqAccountHit{}})
	}
	like := "%" + q + "%"
	var rows []eqAccountHit
	err := e.conn(c).Raw(`
		SELECT id, name FROM account
		WHERE name LIKE ? OR CAST(id AS CHAR) = ?
		ORDER BY name LIMIT 40
	`, like, q).Scan(&rows).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"rows": rows})
}

type eqCharHit struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	AccountID int    `json:"account_id"`
	Level     uint   `json:"level"`
}

func (e *Controller) searchCharacters(c echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	if q == "" {
		return c.JSON(http.StatusOK, echo.Map{"rows": []eqCharHit{}})
	}
	like := "%" + q + "%"
	var rows []eqCharHit
	err := e.conn(c).Raw(`
		SELECT id, name, account_id, level FROM character_data
		WHERE name LIKE ? OR CAST(id AS CHAR) = ?
		ORDER BY name LIMIT 40
	`, like, q).Scan(&rows).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"rows": rows})
}
