package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSessionNotFound = errors.New("refresh session not found")
	ErrRefreshReplay   = errors.New("refresh token replay detected")
)

type Session struct {
	UserID     uint64
	RefreshJTI string
}

type SessionStore interface {
	Create(ctx context.Context, sessionID string, session Session, ttl time.Duration) error
	Rotate(ctx context.Context, sessionID, oldJTI, newJTI string, ttl time.Duration) error
	Delete(ctx context.Context, sessionID string) error
	Blacklist(ctx context.Context, accessJTI string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, accessJTI string) (bool, error)
}

func RefreshSessionKey(sessionID string) string {
	return "travel:auth:refresh:" + sessionID
}

func AccessBlacklistKey(jti string) string {
	return "travel:auth:blacklist:" + jti
}
