package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const JSONPatchKeysKey = "jsonPatchKeys"

func CaptureJSONKeys() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method
			if method != http.MethodPatch && method != http.MethodPut && method != http.MethodPost {
				return next(c)
			}
			if !strings.Contains(c.Request().Header.Get("Content-Type"), "json") {
				return next(c)
			}
			body, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return next(c)
			}
			c.Request().Body = io.NopCloser(bytes.NewReader(body))
			c.Set(JSONPatchKeysKey, jsonObjectKeys(body))
			return next(c)
		}
	}
}

func jsonObjectKeys(body []byte) map[string]bool {
	keys := map[string]bool{}
	var raw interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return keys
	}
	obj, ok := raw.(map[string]interface{})
	if !ok {
		return keys
	}
	collectKeys(obj, keys)
	return keys
}

func collectKeys(obj map[string]interface{}, keys map[string]bool) {
	for key, val := range obj {
		if child, ok := val.(map[string]interface{}); ok {
			collectKeys(child, keys)
			continue
		}
		keys[key] = true
		keys[strings.ToLower(key)] = true
	}
}
