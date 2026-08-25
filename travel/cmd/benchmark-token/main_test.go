package main

import (
	"context"
	"testing"
	"time"
	"travel/auth"
	traveljwt "travel/pkg/jwt"
)

type tokenTestStore struct {
	created bool
}

func (s *tokenTestStore) Create(context.Context, string, auth.Session, time.Duration) error {
	s.created = true
	return nil
}
func (*tokenTestStore) Rotate(context.Context, string, string, string, time.Duration) error {
	return nil
}
func (*tokenTestStore) Delete(context.Context, string) error                   { return nil }
func (*tokenTestStore) Blacklist(context.Context, string, time.Duration) error { return nil }
func (*tokenTestStore) IsBlacklisted(context.Context, string) (bool, error)    { return false, nil }

func TestIssueBenchmarkTokenCreatesAccessTokenAndSession(t *testing.T) {
	config := tokenConfig{
		accessSecret:  []byte("benchmark-access-secret-at-least-32-bytes"),
		refreshSecret: []byte("benchmark-refresh-secret-at-least-32-bytes"),
		issuer:        "ongoing-trip-benchmark",
		userID:        1,
	}
	store := &tokenTestStore{}

	token, err := issueBenchmarkToken(context.Background(), config, store)
	if err != nil {
		t.Fatalf("issueBenchmarkToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("issueBenchmarkToken() returned an empty token")
	}
	if !store.created {
		t.Fatal("issueBenchmarkToken() did not persist a refresh session")
	}
	manager, err := traveljwt.NewManager(traveljwt.Config{
		AccessSecret: config.accessSecret, RefreshSecret: config.refreshSecret,
		AccessTTL: 15 * time.Minute, RefreshTTL: 7 * 24 * time.Hour, Issuer: config.issuer,
	})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := manager.ParseAccess(token)
	if err != nil {
		t.Fatalf("ParseAccess() error = %v", err)
	}
	if claims.UserID != 1 || claims.Issuer != "ongoing-trip-benchmark" {
		t.Fatalf("claims = %+v, want benchmark user and issuer", claims)
	}
}
