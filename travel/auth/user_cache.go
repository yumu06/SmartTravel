package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"travel/TravelModel"

	"github.com/redis/go-redis/v9"
)

type UserCache interface {
	Get(ctx context.Context, userID uint64) (TravelModel.TraUser, bool, error)
	Set(ctx context.Context, user TravelModel.TraUser, ttl time.Duration) error
	Delete(ctx context.Context, userID uint64) error
}

type RedisUserCache struct {
	client redis.UniversalClient
}

func NewRedisUserCache(client redis.UniversalClient) *RedisUserCache {
	return &RedisUserCache{client: client}
}

func (c *RedisUserCache) Get(ctx context.Context, userID uint64) (TravelModel.TraUser, bool, error) {
	payload, err := c.client.Get(ctx, userCacheKey(userID)).Bytes()
	if err == redis.Nil {
		return TravelModel.TraUser{}, false, nil
	}
	if err != nil {
		return TravelModel.TraUser{}, false, err
	}
	var user TravelModel.TraUser
	if err := json.Unmarshal(payload, &user); err != nil {
		return TravelModel.TraUser{}, false, err
	}
	return user, true, nil
}

func (c *RedisUserCache) Set(ctx context.Context, user TravelModel.TraUser, ttl time.Duration) error {
	payload, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, userCacheKey(user.ID), payload, ttl).Err()
}

func (c *RedisUserCache) Delete(ctx context.Context, userID uint64) error {
	return c.client.Del(ctx, userCacheKey(userID)).Err()
}

func userCacheKey(userID uint64) string {
	return fmt.Sprintf("travel:auth:user:%d", userID)
}
