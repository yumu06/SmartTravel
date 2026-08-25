package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	traveljwt "travel/pkg/jwt"
)

type recordingSessionStore struct {
	createdSID     string
	createdSession Session
	rotatedSID     string
	oldJTI         string
	newJTI         string
	deletedSID     string
	blacklistedJTI string
	rotateErr      error
	blacklisted    bool
}

func (s *recordingSessionStore) Create(_ context.Context, sid string, session Session, _ time.Duration) error {
	s.createdSID, s.createdSession = sid, session
	return nil
}
func (s *recordingSessionStore) Rotate(_ context.Context, sid, oldJTI, newJTI string, _ time.Duration) error {
	s.rotatedSID, s.oldJTI, s.newJTI = sid, oldJTI, newJTI
	return s.rotateErr
}
func (s *recordingSessionStore) Delete(_ context.Context, sid string) error {
	s.deletedSID = sid
	return nil
}
func (s *recordingSessionStore) Blacklist(_ context.Context, jti string, _ time.Duration) error {
	s.blacklistedJTI = jti
	return nil
}
func (s *recordingSessionStore) IsBlacklisted(_ context.Context, _ string) (bool, error) {
	return s.blacklisted, nil
}

func newServiceForTest(t *testing.T, store SessionStore) (*Service, *traveljwt.Manager) {
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
	return NewService(manager, store), manager
}

func TestServiceStartSessionPersistsRefreshIdentity(t *testing.T) {
	store := &recordingSessionStore{}
	service, manager := newServiceForTest(t, store)

	pair, err := service.StartSession(context.Background(), 10001)
	if err != nil {
		t.Fatalf("StartSession() error = %v", err)
	}
	claims, err := manager.ParseRefresh(pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if store.createdSID != claims.SessionID || store.createdSession.RefreshJTI != claims.ID || store.createdSession.UserID != 10001 {
		t.Fatalf("stored session = (%q, %+v), want refresh token identity", store.createdSID, store.createdSession)
	}
}

func TestServiceRefreshRotatesWithinExistingSession(t *testing.T) {
	store := &recordingSessionStore{}
	service, manager := newServiceForTest(t, store)
	original, err := service.StartSession(context.Background(), 10001)
	if err != nil {
		t.Fatal(err)
	}
	originalClaims, _ := manager.ParseRefresh(original.RefreshToken)

	rotated, err := service.Refresh(context.Background(), original.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	rotatedClaims, _ := manager.ParseRefresh(rotated.RefreshToken)
	if store.rotatedSID != originalClaims.SessionID || store.oldJTI != originalClaims.ID || store.newJTI != rotatedClaims.ID {
		t.Fatalf("rotation = (%q, %q, %q), want old and new token identities", store.rotatedSID, store.oldJTI, store.newJTI)
	}
	if rotatedClaims.SessionID != originalClaims.SessionID {
		t.Fatal("refresh rotation changed the session ID")
	}
}

func TestServiceRefreshPropagatesReplayDetection(t *testing.T) {
	store := &recordingSessionStore{rotateErr: ErrRefreshReplay}
	service, _ := newServiceForTest(t, store)
	pair, err := service.StartSession(context.Background(), 10001)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Refresh(context.Background(), pair.RefreshToken)
	if !errors.Is(err, ErrRefreshReplay) {
		t.Fatalf("Refresh() error = %v, want ErrRefreshReplay", err)
	}
}

func TestServiceLogoutDeletesSessionAndBlacklistsAccess(t *testing.T) {
	store := &recordingSessionStore{}
	service, manager := newServiceForTest(t, store)
	pair, err := service.StartSession(context.Background(), 10001)
	if err != nil {
		t.Fatal(err)
	}
	claims, _ := manager.ParseAccess(pair.AccessToken)

	if err := service.Logout(context.Background(), pair.AccessToken); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if store.deletedSID != claims.SessionID || store.blacklistedJTI != claims.ID {
		t.Fatalf("logout effects = (%q, %q), want session and access identities", store.deletedSID, store.blacklistedJTI)
	}
}

func TestServiceValidateAccessRejectsBlacklistedToken(t *testing.T) {
	store := &recordingSessionStore{blacklisted: true}
	service, _ := newServiceForTest(t, store)
	pair, err := service.StartSession(context.Background(), 10001)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.ValidateAccess(context.Background(), pair.AccessToken); !errors.Is(err, ErrAccessRevoked) {
		t.Fatalf("ValidateAccess() error = %v, want ErrAccessRevoked", err)
	}
}
