package authn

import (
	"github.com/gin-gonic/gin"
	"github.com/tentackles/shield/modules/authn/internal/api"
	"github.com/tentackles/shield/modules/authn/internal/auth"
	"github.com/tentackles/shield/modules/authn/internal/auth/provider/cognito"
	"github.com/tentackles/shield/modules/authn/internal/config"
	"github.com/tentackles/shield/modules/authn/internal/models"
	"github.com/tentackles/shield/modules/authn/internal/repository"
	"github.com/tentackles/shield/modules/authn/internal/auth/session"
	"github.com/tentackles/shield/modules/authn/internal/auth/nonce"
	"gorm.io/gorm"
)

// GetModelsForMigration returns all models that need to be migrated
func GetModelsForMigration() []interface{} {
	return []interface{}{
		&models.User{},
		&models.Organization{},
		&models.Session{},
		&models.Application{},
		&models.ApplicationRole{},
		&models.UserAppRole{},
		&models.OPAPolicy{},
		&models.PolicySyncStatus{},
	}
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return repository.NewUserRepository(db)
}

// NewAuthService is a public constructor for the AuthN service
func NewAuthService(db *gorm.DB) *auth.AuthService {
	// Load config
	cfg := config.AppConfig
	// Initialize Cognito provider
	provider, _ := cognito.NewProvider(cfg.Cognito) // In production, handle error
	userRepo := NewUserRepository(db)

	// Initialize SessionManager and NonceValidator (replace with actual constructors as needed)
	var sessionManager session.SessionManager // TODO: Replace with actual initialization
	var nonceValidator nonce.NonceValidator   // TODO: Replace with actual initialization

	return auth.NewAuthService(provider, cfg, userRepo, sessionManager, nonceValidator)
}

// RegisterAuthRoutes exposes the route registration for AuthN
func RegisterAuthRoutes(rg *gin.RouterGroup, svc *auth.AuthService) {
	api.RegisterAuthRoutes(rg, svc)
}
