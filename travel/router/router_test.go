package router

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewRouterRegistersRefreshAndLogoutRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	routes := NewRouter(gin.New()).Routes()
	found := map[string]bool{}
	for _, route := range routes {
		found[route.Method+" "+route.Path] = true
	}
	for _, want := range []string{"POST /travel/auth/refresh", "POST /travel/auth/logout"} {
		if !found[want] {
			t.Fatalf("route %q is not registered", want)
		}
	}
}
