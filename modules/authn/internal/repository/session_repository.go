package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tentackles/shield/modules/authn/internal/models"
	"gorm.io/gorm"
)

// SessionRepository handles session persistence operations
type SessionRepository interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error)
	UpdateSession(ctx context.Context, session *models.Session) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteExpiredSessions(ctx context.Context) error
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
}

// GormSessionRepository implements SessionRepository using GORM
type GormSessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &GormSessionRepository{db: db}
}

// CreateSession creates a new session record
func (r *GormSessionRepository) CreateSession(ctx context.Context, session *models.Session) error {
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(session).Error
}

// GetSessionByID retrieves a session by its ID
func (r *GormSessionRepository) GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSession updates an existing session
func (r *GormSessionRepository) UpdateSession(ctx context.Context, session *models.Session) error {
	session.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(session).Error
}

// DeleteSession deletes a session by ID
func (r *GormSessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).Where("id = ?", sessionID).Delete(&models.Session{}).Error
}

// DeleteExpiredSessions removes all expired sessions
func (r *GormSessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	now := time.Now()
	return r.db.WithContext(ctx).Where("expires_at < ? OR (is_active = false AND updated_at < ?)", now, now.Add(-24*time.Hour)).Delete(&models.Session{}).Error
}

// GetSessionsByUserID retrieves all sessions for a user
func (r *GormSessionRepository) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	var sessions []*models.Session
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = true", userID).Find(&sessions).Error
	return sessions, err
}
