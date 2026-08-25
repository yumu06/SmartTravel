package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"travel/TravelModel"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	uuid "github.com/satori/go.uuid"
)

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"
)

var ErrManagerNotConfigured = errors.New("jwt manager is not configured")

type Config struct {
	AccessSecret  []byte
	RefreshSecret []byte
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	Issuer        string
}

type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
}

type Claims struct {
	UserID    uint64 `json:"uid"`
	SessionID string `json:"sid"`
	TokenType string `json:"typ"`
	jwtv5.RegisteredClaims
}

type Claim = Claims

type Pair struct {
	AccessToken      string `json:"access_token"`
	AccessExpiresIn  int    `json:"access_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
}

func NewManager(cfg Config) (*Manager, error) {
	if len(cfg.AccessSecret) == 0 || len(cfg.RefreshSecret) == 0 {
		return nil, errors.New("JWT access and refresh secrets are required")
	}
	if cfg.AccessTTL <= 0 || cfg.RefreshTTL <= 0 {
		return nil, errors.New("JWT access and refresh TTLs must be positive")
	}
	if cfg.Issuer == "" {
		return nil, errors.New("JWT issuer is required")
	}
	return &Manager{
		accessSecret:  append([]byte(nil), cfg.AccessSecret...),
		refreshSecret: append([]byte(nil), cfg.RefreshSecret...),
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
		issuer:        cfg.Issuer,
	}, nil
}

func (m *Manager) IssuePair(userID uint64) (Pair, error) {
	if userID == 0 {
		return Pair{}, errors.New("user ID is required")
	}
	return m.issuePairForSession(userID, uuid.NewV4().String())
}

func (m *Manager) RotatePair(userID uint64, sessionID string) (Pair, error) {
	if userID == 0 || sessionID == "" {
		return Pair{}, errors.New("user ID and session ID are required")
	}
	return m.issuePairForSession(userID, sessionID)
}

func (m *Manager) issuePairForSession(userID uint64, sessionID string) (Pair, error) {
	now := time.Now().UTC().Truncate(time.Second)
	access, err := m.issue(userID, sessionID, uuid.NewV4().String(), AccessTokenType, now, m.accessTTL, m.accessSecret)
	if err != nil {
		return Pair{}, err
	}
	refresh, err := m.issue(userID, sessionID, uuid.NewV4().String(), RefreshTokenType, now, m.refreshTTL, m.refreshSecret)
	if err != nil {
		return Pair{}, err
	}
	return Pair{
		AccessToken:      access,
		AccessExpiresIn:  int(m.accessTTL / time.Second),
		RefreshToken:     refresh,
		RefreshExpiresIn: int(m.refreshTTL / time.Second),
	}, nil
}

func (m *Manager) issue(userID uint64, sid, jti, tokenType string, now time.Time, ttl time.Duration, secret []byte) (string, error) {
	claims := Claims{
		UserID:    userID,
		SessionID: sid,
		TokenType: tokenType,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   strconv.FormatUint(userID, 10),
			ID:        jti,
			IssuedAt:  jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims).SignedString(secret)
}

func (m *Manager) ParseAccess(tokenString string) (*Claims, error) {
	_, claims, err := m.parse(tokenString, m.accessSecret, AccessTokenType)
	return claims, err
}

func (m *Manager) ParseRefresh(tokenString string) (*Claims, error) {
	_, claims, err := m.parse(tokenString, m.refreshSecret, RefreshTokenType)
	return claims, err
}

func (m *Manager) parse(tokenString string, secret []byte, expectedType string) (*jwtv5.Token, *Claims, error) {
	claims := &Claims{}
	token, err := jwtv5.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwtv5.Token) (interface{}, error) { return secret, nil },
		jwtv5.WithValidMethods([]string{jwtv5.SigningMethodHS256.Alg()}),
		jwtv5.WithIssuer(m.issuer),
		jwtv5.WithExpirationRequired(),
	)
	if err != nil {
		return token, claims, err
	}
	if !token.Valid || claims.TokenType != expectedType {
		return token, claims, fmt.Errorf("invalid %s token", expectedType)
	}
	if claims.UserID == 0 || claims.SessionID == "" || claims.ID == "" {
		return token, claims, errors.New("token is missing required claims")
	}
	return token, claims, nil
}

var defaultManager *Manager

func SetDefaultManager(manager *Manager) {
	defaultManager = manager
}

func ReleaseToken(user TravelModel.TraUser) (string, error) {
	if defaultManager == nil {
		return "", ErrManagerNotConfigured
	}
	pair, err := defaultManager.IssuePair(user.ID)
	return pair.AccessToken, err
}

func ParseToken(tokenString string) (*jwtv5.Token, *Claim, error) {
	if defaultManager == nil {
		return nil, nil, ErrManagerNotConfigured
	}
	return defaultManager.parse(tokenString, defaultManager.accessSecret, AccessTokenType)
}
