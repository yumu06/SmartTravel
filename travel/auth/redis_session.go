package auth

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var rotateRefreshScript = redis.NewScript(`
local current = redis.call('HGET', KEYS[1], 'refresh_jti')
if not current then
    return 0
end
if current ~= ARGV[1] then
    redis.call('DEL', KEYS[1])
    return -1
end
redis.call('HSET', KEYS[1], 'refresh_jti', ARGV[2])
redis.call('EXPIRE', KEYS[1], ARGV[3])
return 1
`)

type RedisSessionStore struct {
	client redis.UniversalClient
}

func NewRedisSessionStore(client redis.UniversalClient) *RedisSessionStore {
	return &RedisSessionStore{client: client}
}

func (s *RedisSessionStore) Create(ctx context.Context, sessionID string, session Session, ttl time.Duration) error {
	if sessionID == "" || session.UserID == 0 || session.RefreshJTI == "" || ttl <= 0 {
		return errors.New("invalid refresh session")
	}
	key := RefreshSessionKey(sessionID)
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, key, "user_id", strconv.FormatUint(session.UserID, 10), "refresh_jti", session.RefreshJTI)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisSessionStore) Rotate(ctx context.Context, sessionID, oldJTI, newJTI string, ttl time.Duration) error {
	if sessionID == "" || oldJTI == "" || newJTI == "" || ttl <= 0 {
		return errors.New("invalid refresh rotation")
	}
	result, err := rotateRefreshScript.Run(
		ctx,
		s.client,
		[]string{RefreshSessionKey(sessionID)},
		oldJTI,
		newJTI,
		int64(ttl/time.Second),
	).Int64()
	if err != nil {
		return err
	}
	switch result {
	case 1:
		return nil
	case -1:
		return ErrRefreshReplay
	default:
		return ErrSessionNotFound
	}
}

func (s *RedisSessionStore) Delete(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}
	return s.client.Del(ctx, RefreshSessionKey(sessionID)).Err()
}

func (s *RedisSessionStore) Blacklist(ctx context.Context, accessJTI string, ttl time.Duration) error {
	if accessJTI == "" || ttl <= 0 {
		return errors.New("invalid blacklist entry")
	}
	return s.client.Set(ctx, AccessBlacklistKey(accessJTI), "1", ttl).Err()
}

func (s *RedisSessionStore) IsBlacklisted(ctx context.Context, accessJTI string) (bool, error) {
	if accessJTI == "" {
		return false, errors.New("access JTI is required")
	}
	count, err := s.client.Exists(ctx, AccessBlacklistKey(accessJTI)).Result()
	return count > 0, err
}
