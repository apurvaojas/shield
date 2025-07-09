package authn

import (
	"shield/cmd/app/config"
	"shield/modules/authn/internal/api"
	"shield/modules/authn/internal/auth"
	"shield/modules/authn/internal/auth/nonce"
	"shield/modules/authn/internal/auth/provider/cognito"
	"shield/modules/authn/internal/auth/session"
	"shield/modules/authn/internal/models"
	"shield/modules/authn/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
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
	cfg := config.GetConfig()

	// Initialize Cognito provider
	provider, err := cognito.NewProvider(cfg.Cognito)
	if err != nil {
		// Log error and use a mock provider for development
		// In production, this should fail gracefully or use fallback
		provider = nil // This will need to be handled in the service
	}

	userRepo := NewUserRepository(db)

	// Initialize SessionManager with database backend
	sessionRepo := repository.NewSessionRepository(db)
	sessionConfig := session.SessionConfig{
		SessionTTL:    24 * time.Hour,     // 24 hours
		RefreshTTL:    7 * 24 * time.Hour, // 7 days
		MaxSessions:   5,                  // Max 5 sessions per user
		SecureCookies: cfg.Server.Environment == "production",
	}
	sessionManager := session.NewDefaultSessionManager(sessionRepo, sessionConfig)

	// Initialize NonceValidator with 5 minute TTL
	nonceValidator := nonce.NewInMemoryNonceValidator(5 * time.Minute)

	return auth.NewAuthService(provider, cfg, userRepo, sessionManager, nonceValidator)
}

// RegisterAuthRoutes exposes the route registration for AuthN
func RegisterAuthRoutes(rg *gin.RouterGroup, svc *auth.AuthService) {
	api.RegisterAuthRoutes(rg, svc)
}
