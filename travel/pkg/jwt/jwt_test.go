package jwt

import (
	"testing"
	"time"
)

func testManager(t *testing.T) *Manager {
	t.Helper()
	manager, err := NewManager(Config{
		AccessSecret:  []byte("access-secret-at-least-32-bytes-long"),
		RefreshSecret: []byte("refresh-secret-at-least-32-bytes-long"),
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
		Issuer:        "ongoing-trip",
	})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	return manager
}

func TestNewManagerRejectsMissingSecrets(t *testing.T) {
	_, err := NewManager(Config{AccessTTL: time.Minute, RefreshTTL: time.Hour, Issuer: "ongoing-trip"})
	if err == nil {
		t.Fatal("NewManager() error = nil, want missing-secret error")
	}
}

func TestIssuePairCreatesTypedTokensWithConfiguredLifetimes(t *testing.T) {
	manager := testManager(t)
	before := time.Now()

	pair, err := manager.IssuePair(10001)
	if err != nil {
		t.Fatalf("IssuePair() error = %v", err)
	}
	access, err := manager.ParseAccess(pair.AccessToken)
	if err != nil {
		t.Fatalf("ParseAccess() error = %v", err)
	}
	refresh, err := manager.ParseRefresh(pair.RefreshToken)
	if err != nil {
		t.Fatalf("ParseRefresh() error = %v", err)
	}

	if access.UserID != 10001 || access.Subject != "10001" || access.TokenType != AccessTokenType {
		t.Fatalf("access claims = %+v, want user 10001 and access type", access)
	}
	if refresh.UserID != 10001 || refresh.TokenType != RefreshTokenType {
		t.Fatalf("refresh claims = %+v, want user 10001 and refresh type", refresh)
	}
	if access.SessionID == "" || access.SessionID != refresh.SessionID {
		t.Fatalf("session IDs = (%q, %q), want one shared non-empty SID", access.SessionID, refresh.SessionID)
	}
	if access.ID == "" || refresh.ID == "" || access.ID == refresh.ID {
		t.Fatalf("token JTIs = (%q, %q), want distinct non-empty values", access.ID, refresh.ID)
	}
	if got := access.ExpiresAt.Time.Sub(access.IssuedAt.Time); got != 15*time.Minute {
		t.Fatalf("access lifetime = %s, want 15m", got)
	}
	if got := refresh.ExpiresAt.Time.Sub(refresh.IssuedAt.Time); got != 7*24*time.Hour {
		t.Fatalf("refresh lifetime = %s, want 168h", got)
	}
	if access.IssuedAt.Time.Before(before.Add(-time.Second)) {
		t.Fatalf("issued at = %s, want current time", access.IssuedAt.Time)
	}
	if pair.AccessExpiresIn != 900 || pair.RefreshExpiresIn != 604800 {
		t.Fatalf("expires-in = (%d, %d), want (900, 604800)", pair.AccessExpiresIn, pair.RefreshExpiresIn)
	}
}

func TestParseRejectsWrongTokenType(t *testing.T) {
	manager := testManager(t)
	pair, err := manager.IssuePair(10001)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := manager.ParseAccess(pair.RefreshToken); err == nil {
		t.Fatal("ParseAccess(refresh) error = nil, want token-type error")
	}
	if _, err := manager.ParseRefresh(pair.AccessToken); err == nil {
		t.Fatal("ParseRefresh(access) error = nil, want token-type error")
	}
}

func TestParseRejectsTokenSignedWithAnotherSecret(t *testing.T) {
	manager := testManager(t)
	pair, err := manager.IssuePair(10001)
	if err != nil {
		t.Fatal(err)
	}
	other, err := NewManager(Config{
		AccessSecret:  []byte("another-access-secret-32-bytes-long"),
		RefreshSecret: []byte("another-refresh-secret-32-bytes-long"),
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
		Issuer:        "ongoing-trip",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := other.ParseAccess(pair.AccessToken); err == nil {
		t.Fatal("ParseAccess() error = nil, want signature error")
	}
}

func TestRotatePairPreservesSessionAndChangesRefreshJTI(t *testing.T) {
	manager := testManager(t)
	original, err := manager.IssuePair(10001)
	if err != nil {
		t.Fatal(err)
	}
	originalClaims, err := manager.ParseRefresh(original.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}

	rotated, err := manager.RotatePair(10001, originalClaims.SessionID)
	if err != nil {
		t.Fatalf("RotatePair() error = %v", err)
	}
	rotatedClaims, err := manager.ParseRefresh(rotated.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if rotatedClaims.SessionID != originalClaims.SessionID {
		t.Fatalf("rotated SID = %q, want %q", rotatedClaims.SessionID, originalClaims.SessionID)
	}
	if rotatedClaims.ID == originalClaims.ID {
		t.Fatalf("rotated JTI = %q, want a new value", rotatedClaims.ID)
	}
}
