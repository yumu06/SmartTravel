package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"travel/auth"
	traveljwt "travel/pkg/jwt"
)

type controllerSessionStore struct {
	refreshJTI string
	deletedSID string
	blacklist  string
}

func (s *controllerSessionStore) Create(_ context.Context, _ string, session auth.Session, _ time.Duration) error {
	s.refreshJTI = session.RefreshJTI
	return nil
}
func (s *controllerSessionStore) Rotate(_ context.Context, _ string, oldJTI, newJTI string, _ time.Duration) error {
	if oldJTI != s.refreshJTI {
		return auth.ErrRefreshReplay
	}
	s.refreshJTI = newJTI
	return nil
}
func (s *controllerSessionStore) Delete(_ context.Context, sid string) error {
	s.deletedSID = sid
	return nil
}
func (s *controllerSessionStore) Blacklist(_ context.Context, jti string, _ time.Duration) error {
	s.blacklist = jti
	return nil
}
func (s *controllerSessionStore) IsBlacklisted(context.Context, string) (bool, error) {
	return false, nil
}

func setupControllerAuth(t *testing.T) (*auth.Service, *traveljwt.Manager, *controllerSessionStore) {
	t.Helper()
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
	store := &controllerSessionStore{}
	service := auth.NewService(manager, store)
	auth.SetDefaultService(service)
	t.Cleanup(func() { auth.SetDefaultService(nil) })
	return service, manager, store
}

func TestAuthRefreshReturnsRotatedPair(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service, manager, _ := setupControllerAuth(t)
	pair, err := service.StartSession(context.Background(), 10001)
	if err != nil {
		t.Fatal(err)
	}
	originalClaims, _ := manager.ParseRefresh(pair.RefreshToken)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/travel/auth/refresh", strings.NewReader(`{"refresh_token":"`+pair.RefreshToken+`"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	AuthRefresh(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Code int            `json:"code"`
		Data traveljwt.Pair `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	rotatedClaims, err := manager.ParseRefresh(response.Data.RefreshToken)
	if err != nil {
		t.Fatalf("response refresh token is invalid: %v", err)
	}
	if response.Code != 200 || response.Data.AccessToken == "" || rotatedClaims.SessionID != originalClaims.SessionID || rotatedClaims.ID == originalClaims.ID {
		t.Fatalf("response = %+v, want rotated pair in existing session", response)
	}
}

func TestAuthRefreshRejectsMalformedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupControllerAuth(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/travel/auth/refresh", strings.NewReader(`{}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	AuthRefresh(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
}

func TestAuthLogoutRevokesCurrentSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service, manager, store := setupControllerAuth(t)
	pair, err := service.StartSession(context.Background(), 10001)
	if err != nil {
		t.Fatal(err)
	}
	claims, _ := manager.ParseAccess(pair.AccessToken)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/travel/auth/logout", nil)
	ctx.Request.Header.Set("Authorization", "Bearer "+pair.AccessToken)

	AuthLogout(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if store.deletedSID != claims.SessionID || store.blacklist != claims.ID {
		t.Fatalf("revocation = (%q, %q), want (%q, %q)", store.deletedSID, store.blacklist, claims.SessionID, claims.ID)
	}
}

func TestLoginSuccessKeepsLegacyTokenAndAddsPair(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	pair := traveljwt.Pair{AccessToken: "access", AccessExpiresIn: 900, RefreshToken: "refresh", RefreshExpiresIn: 604800}

	writeLoginSuccess(ctx, pair, "session-key")

	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response["token"] != "access" || response["data"] == nil {
		t.Fatalf("response = %s, want legacy token and structured pair", recorder.Body.String())
	}
}
