package auth

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestSessionStore(t *testing.T) (*RedisSessionStore, *miniredis.Miniredis) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return NewRedisSessionStore(client), server
}

func TestRedisSessionCreateStoresIdentityWithTTL(t *testing.T) {
	store, server := newTestSessionStore(t)
	ctx := context.Background()
	session := Session{UserID: 10001, RefreshJTI: "refresh-1"}

	if err := store.Create(ctx, "session-1", session, 7*24*time.Hour); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	key := RefreshSessionKey("session-1")
	if got := server.HGet(key, "user_id"); got != strconv.FormatUint(session.UserID, 10) {
		t.Fatalf("user_id = %q, want 10001", got)
	}
	if got := server.HGet(key, "refresh_jti"); got != "refresh-1" {
		t.Fatalf("refresh_jti = %q, want refresh-1", got)
	}
	if got := server.TTL(key); got != 7*24*time.Hour {
		t.Fatalf("TTL = %s, want 168h", got)
	}
}

func TestRedisSessionRotateAllowsOldRefreshOnlyOnce(t *testing.T) {
	store, server := newTestSessionStore(t)
	ctx := context.Background()
	if err := store.Create(ctx, "session-1", Session{UserID: 10001, RefreshJTI: "old-jti"}, time.Hour); err != nil {
		t.Fatal(err)
	}

	if err := store.Rotate(ctx, "session-1", "old-jti", "new-jti", time.Hour); err != nil {
		t.Fatalf("first Rotate() error = %v", err)
	}
	if got := server.HGet(RefreshSessionKey("session-1"), "refresh_jti"); got != "new-jti" {
		t.Fatalf("refresh_jti = %q, want new-jti", got)
	}

	err := store.Rotate(ctx, "session-1", "old-jti", "another-jti", time.Hour)
	if !errors.Is(err, ErrRefreshReplay) {
		t.Fatalf("second Rotate() error = %v, want ErrRefreshReplay", err)
	}
	if server.Exists(RefreshSessionKey("session-1")) {
		t.Fatal("session still exists after replay, want deletion")
	}
}

func TestRedisSessionDeleteAndBlacklist(t *testing.T) {
	store, server := newTestSessionStore(t)
	ctx := context.Background()
	if err := store.Create(ctx, "session-1", Session{UserID: 10001, RefreshJTI: "jti"}, time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(ctx, "session-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if server.Exists(RefreshSessionKey("session-1")) {
		t.Fatal("session still exists after Delete()")
	}

	if err := store.Blacklist(ctx, "access-jti", 10*time.Minute); err != nil {
		t.Fatalf("Blacklist() error = %v", err)
	}
	blacklisted, err := store.IsBlacklisted(ctx, "access-jti")
	if err != nil || !blacklisted {
		t.Fatalf("IsBlacklisted() = (%v, %v), want (true, nil)", blacklisted, err)
	}
	server.FastForward(11 * time.Minute)
	blacklisted, err = store.IsBlacklisted(ctx, "access-jti")
	if err != nil || blacklisted {
		t.Fatalf("IsBlacklisted() after expiry = (%v, %v), want (false, nil)", blacklisted, err)
	}
}

func TestRedisSessionPropagatesRedisFailure(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	store := NewRedisSessionStore(client)
	server.Close()
	t.Cleanup(func() { _ = client.Close() })

	err := store.Create(context.Background(), "session-1", Session{UserID: 10001, RefreshJTI: "jti"}, time.Hour)
	if err == nil {
		t.Fatal("Create() error = nil after Redis shutdown")
	}
}
