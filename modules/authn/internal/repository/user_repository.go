package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/tentackles/shield/modules/authn/internal/models"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user database operations
type UserRepository interface {
	// User operations
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByCognitoSub(ctx context.Context, cognitoSub string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// Organization operations
	CreateOrganization(ctx context.Context, org *models.Organization) error
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error)
	GetOrganizationByName(ctx context.Context, name string) (*models.Organization, error)
	UpdateOrganization(ctx context.Context, org *models.Organization) error

	// Session operations
	CreateSession(ctx context.Context, session *models.Session) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error)
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error)
	UpdateSession(ctx context.Context, session *models.Session) error
	DeleteSession(ctx context.Context, id uuid.UUID) error
	DeleteExpiredSessions(ctx context.Context) error
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// User operations
func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Organization").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Organization").First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByCognitoSub(ctx context.Context, cognitoSub string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Organization").First(&user, "cognito_sub = ?", cognitoSub).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}

// Organization operations
func (r *userRepository) CreateOrganization(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

func (r *userRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).First(&org, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *userRepository) GetOrganizationByName(ctx context.Context, name string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.WithContext(ctx).First(&org, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *userRepository) UpdateOrganization(ctx context.Context, org *models.Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

// Session operations
func (r *userRepository) CreateSession(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *userRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).Preload("User").First(&session, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *userRepository) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	var sessions []models.Session
	err := r.db.WithContext(ctx).Find(&sessions, "user_id = ?", userID).Error
	return sessions, err
}

func (r *userRepository) UpdateSession(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *userRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Session{}, "id = ?", id).Error
}

func (r *userRepository) DeleteExpiredSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < NOW()").Delete(&models.Session{}).Error
}
