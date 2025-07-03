package session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tentackles/shield/modules/authn/internal/models"
)

// SessionManager handles session creation, validation, and cleanup
type SessionManager interface {
	CreateSession(ctx context.Context, userID uuid.UUID, clientInfo ClientInfo) (*models.Session, error)
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	ValidateSession(ctx context.Context, sessionID string) (*models.Session, error)
	InvalidateSession(ctx context.Context, sessionID string) error
	RefreshSession(ctx context.Context, sessionID string) (*models.Session, error)
	CleanupExpiredSessions(ctx context.Context) error
}

// ClientInfo contains information about the client making the request
type ClientInfo struct {
	IPAddress string
	UserAgent string
	DeviceID  string
}

// SessionRepository defines the interface for session persistence
type SessionRepository interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error)
	UpdateSession(ctx context.Context, session *models.Session) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteExpiredSessions(ctx context.Context) error
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
}

// DefaultSessionManager is the default implementation of SessionManager
type DefaultSessionManager struct {
	repository SessionRepository
	config     SessionConfig
}

// SessionConfig contains configuration for session management
type SessionConfig struct {
	SessionTTL    time.Duration
	RefreshTTL    time.Duration
	MaxSessions   int // Maximum sessions per user
	SecureCookies bool
}

// NewDefaultSessionManager creates a new session manager
func NewDefaultSessionManager(repo SessionRepository, config SessionConfig) *DefaultSessionManager {
	return &DefaultSessionManager{
		repository: repo,
		config:     config,
	}
}

// CreateSession creates a new session for a user
func (sm *DefaultSessionManager) CreateSession(ctx context.Context, userID uuid.UUID, clientInfo ClientInfo) (*models.Session, error) {
	sessionID := uuid.New().String()
	refreshToken := uuid.New().String()

	now := time.Now()
	session := &models.Session{
		ID:               sessionID,
		UserID:           userID,
		RefreshToken:     refreshToken,
		IPAddress:        clientInfo.IPAddress,
		UserAgent:        clientInfo.UserAgent,
		DeviceID:         clientInfo.DeviceID,
		CreatedAt:        now,
		ExpiresAt:        now.Add(sm.config.SessionTTL),
		RefreshExpiresAt: now.Add(sm.config.RefreshTTL),
		IsActive:         true,
	}

	// Check if we need to enforce max sessions per user
	if sm.config.MaxSessions > 0 {
		existingSessions, err := sm.repository.GetSessionsByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing sessions: %w", err)
		}

		if len(existingSessions) >= sm.config.MaxSessions {
			// Remove oldest session
			oldestSession := existingSessions[0]
			for _, s := range existingSessions {
				if s.CreatedAt.Before(oldestSession.CreatedAt) {
					oldestSession = s
				}
			}
			_ = sm.repository.DeleteSession(ctx, oldestSession.ID)
		}
	}

	if err := sm.repository.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *DefaultSessionManager) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := sm.repository.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// ValidateSession validates a session and returns it if valid
func (sm *DefaultSessionManager) ValidateSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if !session.IsActive {
		return nil, fmt.Errorf("session is inactive")
	}

	if time.Now().After(session.ExpiresAt) {
		// Session expired, mark as inactive
		session.IsActive = false
		_ = sm.repository.UpdateSession(ctx, session)
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// InvalidateSession marks a session as inactive
func (sm *DefaultSessionManager) InvalidateSession(ctx context.Context, sessionID string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.IsActive = false
	session.ExpiresAt = time.Now()

	if err := sm.repository.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	return nil
}

// RefreshSession extends the session lifetime using refresh token
func (sm *DefaultSessionManager) RefreshSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if !session.IsActive {
		return nil, fmt.Errorf("session is inactive")
	}

	if time.Now().After(session.RefreshExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	// Update session expiry
	now := time.Now()
	session.ExpiresAt = now.Add(sm.config.SessionTTL)
	session.RefreshExpiresAt = now.Add(sm.config.RefreshTTL)
	session.RefreshToken = uuid.New().String() // Rotate refresh token

	if err := sm.repository.UpdateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}

	return session, nil
}

// CleanupExpiredSessions removes expired sessions from storage
func (sm *DefaultSessionManager) CleanupExpiredSessions(ctx context.Context) error {
	if err := sm.repository.DeleteExpiredSessions(ctx); err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	return nil
}
