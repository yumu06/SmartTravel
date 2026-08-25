package auth

import (
	"context"
	"errors"
	"testing"
	"time"
	"travel/TravelModel"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestServiceResolveUserCachesSuccessfulDatabaseLoad(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	service := NewService(nil, nil, WithUserCache(NewRedisUserCache(client), 5*time.Minute))
	loads := 0
	load := func() (TravelModel.TraUser, error) {
		loads++
		return TravelModel.TraUser{ID: 7, OpenID: "benchmark-user-7", SessionKey: "session-7"}, nil
	}

	first, err := service.ResolveUser(context.Background(), 7, load)
	if err != nil {
		t.Fatalf("first ResolveUser() error = %v", err)
	}
	second, err := service.ResolveUser(context.Background(), 7, load)
	if err != nil {
		t.Fatalf("second ResolveUser() error = %v", err)
	}
	if loads != 1 {
		t.Fatalf("database loads = %d, want 1", loads)
	}
	if first.ID != 7 || second.OpenID != "benchmark-user-7" || second.SessionKey != "session-7" {
		t.Fatalf("resolved users = (%+v, %+v)", first, second)
	}
}

func TestServiceResolveUserReloadsAfterCacheTTL(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	service := NewService(nil, nil, WithUserCache(NewRedisUserCache(client), time.Minute))
	loads := 0
	load := func() (TravelModel.TraUser, error) {
		loads++
		return TravelModel.TraUser{ID: 7, OpenID: "benchmark-user-7"}, nil
	}

	if _, err := service.ResolveUser(context.Background(), 7, load); err != nil {
		t.Fatal(err)
	}
	server.FastForward(61 * time.Second)
	if _, err := service.ResolveUser(context.Background(), 7, load); err != nil {
		t.Fatal(err)
	}
	if loads != 2 {
		t.Fatalf("database loads = %d after cache expiry, want 2", loads)
	}
}

func TestServiceResolveUserFallsBackWhenCacheFails(t *testing.T) {
	service := NewService(nil, nil, WithUserCache(failingUserCache{}, time.Minute))
	want := TravelModel.TraUser{ID: 9, OpenID: "database-user-9"}

	got, err := service.ResolveUser(context.Background(), 9, func() (TravelModel.TraUser, error) {
		return want, nil
	})
	if err != nil {
		t.Fatalf("ResolveUser() error = %v", err)
	}
	if got.ID != want.ID || got.OpenID != want.OpenID {
		t.Fatalf("ResolveUser() = %+v, want %+v", got, want)
	}
}

type failingUserCache struct{}

func (failingUserCache) Get(context.Context, uint64) (TravelModel.TraUser, bool, error) {
	return TravelModel.TraUser{}, false, errors.New("redis unavailable")
}
func (failingUserCache) Set(context.Context, TravelModel.TraUser, time.Duration) error {
	return errors.New("redis unavailable")
}
func (failingUserCache) Delete(context.Context, uint64) error {
	return errors.New("redis unavailable")
}
