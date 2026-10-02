package permissions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestIsWriteRequest(t *testing.T) {
	s := &Service{}
	e := echo.New()
	cases := []struct {
		method string
		path   string
		write  bool
	}{
		{http.MethodGet, "/api/v1/items", false},
		{http.MethodHead, "/api/v1/items", false},
		{http.MethodPost, "/api/v1/items/bulk", false},
		{http.MethodPost, "/api/v1/inventories/bulk", false},
		{http.MethodPost, "/api/v1/items_evolving_details/synchronize", true},
		{http.MethodPatch, "/api/v1/inventory/1", true},
		{http.MethodPut, "/api/v1/inventory", true},
		{http.MethodDelete, "/api/v1/inventory/1", true},
	}
	for _, tc := range cases {
		c := e.NewContext(httptest.NewRequest(tc.method, tc.path, nil), httptest.NewRecorder())
		if got := s.IsWriteRequest(c); got != tc.write {
			t.Fatalf("%s %s: write=%v want %v", tc.method, tc.path, got, tc.write)
		}
	}
}

func TestServerFilesResourceRegistered(t *testing.T) {
	s := &Service{}
	paths, ok := s.RegisterManualResources()["Server Files"]
	if !ok || len(paths) < 3 || paths[0] != "admin/server-files" || paths[1] != "admin/zone-controller" || paths[2] != "admin/ultimate-systems" {
		t.Fatalf("Server Files resource missing: %#v", paths)
	}
}
