package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"travel/auth"
	traveljwt "travel/pkg/jwt"
)

type middlewareSessionStore struct {
	blacklistChecked bool
}

func (s *middlewareSessionStore) Create(context.Context, string, auth.Session, time.Duration) error {
	return nil
}
func (s *middlewareSessionStore) Rotate(context.Context, string, string, string, time.Duration) error {
	return nil
}
func (s *middlewareSessionStore) Delete(context.Context, string) error { return nil }
func (s *middlewareSessionStore) Blacklist(context.Context, string, time.Duration) error {
	return nil
}
func (s *middlewareSessionStore) IsBlacklisted(context.Context, string) (bool, error) {
	s.blacklistChecked = true
	return true, nil
}

func TestAuthMiddlewareChecksAccessBlacklistBeforeDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager, err := traveljwt.NewManager(traveljwt.Config{
		AccessSecret:  []byte("access-secret-at-least-32-bytes-long"),
		RefreshSecret: []byte("refresh-secret-at-least-32-bytes-long"),
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
		Issuer:        "ongoing-trip",
	})
	if err != nil {
		t.Fatal(err)
	}
	store := &middlewareSessionStore{}
	service := auth.NewService(manager, store)
	auth.SetDefaultService(service)
	t.Cleanup(func() { auth.SetDefaultService(nil) })
	pair, err := service.StartSession(context.Background(), 10001)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/travel/authorization", nil)
	ctx.Request.Header.Set("Authorization", "Bearer "+pair.AccessToken)

	AuthMiddleware()(ctx)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
	if !store.blacklistChecked {
		t.Fatal("session store blacklist was not checked")
	}
}
