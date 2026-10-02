package middleware

import (
	"fmt"
	"github.com/EQEmu/spire/internal/database"
	"github.com/EQEmu/spire/internal/env"
	"github.com/EQEmu/spire/internal/http/request"
	"github.com/EQEmu/spire/internal/logger"
	"github.com/EQEmu/spire/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

type ReadOnlyMiddleware struct {
	db     *database.Resolver
	logger *logger.AppLogger
}

func NewReadOnlyMiddleware(db *database.Resolver, logger *logger.AppLogger) *ReadOnlyMiddleware {
	return &ReadOnlyMiddleware{
		db:     db,
		logger: logger,
	}
}

func (r ReadOnlyMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			// continue request if this mode is not even enabled
			if !env.IsHostedReadOnlyModeEnabled() {
				return next(c)
			}

			path := c.Request().URL.Path
			method := c.Request().Method
			isRead := method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
			isBulkFetch := method == http.MethodPost && strings.Contains(path, "/bulk")
			isConnection := strings.Contains(path, "/api/v1/connection")
			if isRead || isBulkFetch || isConnection {
				return next(c)
			}

			// anything else we assume is something attempting to write
			user := request.GetUser(c)
			if user.ID == 0 {
				return c.JSON(
					http.StatusForbidden,
					echo.Map{"error": "Not logged in"},
				)
			}

			// if we have a user, lets check to see if our user database is equal to the default
			if user.ID > 0 {
				userDb, err := r.db.ResolveUserEqemuConnection(&models.Zone{}, user).DB()
				if err != nil {
					return c.JSON(
						http.StatusForbidden,
						echo.Map{"error": fmt.Sprintf("[read_only_middleware] Can't get user database %v", err.Error())},
					)
				}
				eqemuDb, err := r.db.GetEqemuDb().DB()
				if err != nil {
					return c.JSON(
						http.StatusForbidden,
						echo.Map{"error": fmt.Sprintf("[read_only_middleware] Can't get eqemu database %v", err.Error())},
					)
				}

				if userDb == eqemuDb {
					r.logger.Debug().
						Any("user", user).
						Msg("User attempted to write to database in read only mode")

					return c.JSON(
						http.StatusForbidden,
						echo.Map{"error": "Database is in read only mode"},
					)
				}
			}

			return next(c)
		}
	}
}
