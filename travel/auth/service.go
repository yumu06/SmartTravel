package auth

import (
	"context"
	"errors"
	"time"
	"travel/TravelModel"

	traveljwt "travel/pkg/jwt"
)

var ErrAccessRevoked = errors.New("access token has been revoked")

type TokenManager interface {
	IssuePair(userID uint64) (traveljwt.Pair, error)
	RotatePair(userID uint64, sessionID string) (traveljwt.Pair, error)
	ParseAccess(token string) (*traveljwt.Claims, error)
	ParseRefresh(token string) (*traveljwt.Claims, error)
}

type Service struct {
	tokens       TokenManager
	store        SessionStore
	userCache    UserCache
	userCacheTTL time.Duration
}

type ServiceOption func(*Service)

func WithUserCache(cache UserCache, ttl time.Duration) ServiceOption {
	return func(service *Service) {
		service.userCache = cache
		service.userCacheTTL = ttl
	}
}

func NewService(tokens TokenManager, store SessionStore, options ...ServiceOption) *Service {
	service := &Service{tokens: tokens, store: store}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *Service) ResolveUser(ctx context.Context, userID uint64, load func() (TravelModel.TraUser, error)) (TravelModel.TraUser, error) {
	if s.userCache != nil && s.userCacheTTL > 0 {
		user, found, err := s.userCache.Get(ctx, userID)
		if err == nil && found && user.ID == userID {
			return user, nil
		}
	}
	user, err := load()
	if err != nil {
		return TravelModel.TraUser{}, err
	}
	if s.userCache != nil && s.userCacheTTL > 0 {
		_ = s.userCache.Set(ctx, user, s.userCacheTTL)
	}
	return user, nil
}

func (s *Service) StartSession(ctx context.Context, userID uint64) (traveljwt.Pair, error) {
	pair, err := s.tokens.IssuePair(userID)
	if err != nil {
		return traveljwt.Pair{}, err
	}
	claims, err := s.tokens.ParseRefresh(pair.RefreshToken)
	if err != nil {
		return traveljwt.Pair{}, err
	}
	if err := s.store.Create(ctx, claims.SessionID, Session{UserID: userID, RefreshJTI: claims.ID}, remainingTTL(claims.ExpiresAt.Time)); err != nil {
		return traveljwt.Pair{}, err
	}
	return pair, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (traveljwt.Pair, error) {
	oldClaims, err := s.tokens.ParseRefresh(refreshToken)
	if err != nil {
		return traveljwt.Pair{}, err
	}
	pair, err := s.tokens.RotatePair(oldClaims.UserID, oldClaims.SessionID)
	if err != nil {
		return traveljwt.Pair{}, err
	}
	newClaims, err := s.tokens.ParseRefresh(pair.RefreshToken)
	if err != nil {
		return traveljwt.Pair{}, err
	}
	if err := s.store.Rotate(ctx, oldClaims.SessionID, oldClaims.ID, newClaims.ID, remainingTTL(newClaims.ExpiresAt.Time)); err != nil {
		return traveljwt.Pair{}, err
	}
	return pair, nil
}

func (s *Service) Logout(ctx context.Context, accessToken string) error {
	claims, err := s.tokens.ParseAccess(accessToken)
	if err != nil {
		return err
	}
	if err := s.store.Delete(ctx, claims.SessionID); err != nil {
		return err
	}
	ttl := remainingTTL(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}
	return s.store.Blacklist(ctx, claims.ID, ttl)
}

func (s *Service) ValidateAccess(ctx context.Context, accessToken string) (*traveljwt.Claims, error) {
	claims, err := s.tokens.ParseAccess(accessToken)
	if err != nil {
		return nil, err
	}
	blacklisted, err := s.store.IsBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, ErrAccessRevoked
	}
	return claims, nil
}

func remainingTTL(expiresAt time.Time) time.Duration {
	return time.Until(expiresAt)
}

var defaultService *Service

func SetDefaultService(service *Service) {
	defaultService = service
}

func DefaultService() (*Service, error) {
	if defaultService == nil {
		return nil, errors.New("authentication service is not configured")
	}
	return defaultService, nil
}
