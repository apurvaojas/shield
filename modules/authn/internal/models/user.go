package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserType defines the type of user
type UserType string

const (
	UserTypeIndividual   UserType = "individual"
	UserTypeOrganization UserType = "organization"
)

// MFAMethod defines the type of MFA method
type MFAMethod string

const (
	MFAMethodTOTP MFAMethod = "TOTP"
	MFAMethodSMS  MFAMethod = "SMS"
	// Add other methods as needed
)

type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email      string    `gorm:"uniqueIndex;not null" json:"email"`
	CognitoSub string    `gorm:"uniqueIndex" json:"cognito_sub,omitempty"`
	OrgID      uuid.UUID `gorm:"type:uuid" json:"org_id"`
	UserType   UserType  `gorm:"type:varchar(50)" json:"user_type"`
	IsVerified bool      `gorm:"default:false" json:"is_verified"` // Add email verification status
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relationships
	Organization *Organization `gorm:"foreignKey:OrgID" json:"organization,omitempty"` // This will now refer to Organization in organization.go
	Sessions     []Session     `gorm:"foreignKey:UserID" json:"sessions,omitempty"`
	UserAppRoles []UserAppRole `gorm:"foreignKey:UserID" json:"user_app_roles,omitempty"`
}

// Organization struct is now defined in organization.go

type Session struct {
	ID               string    `gorm:"type:varchar(255);primary_key" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	RefreshToken     string    `gorm:"type:varchar(255);not null" json:"refresh_token"`
	IPAddress        string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent        string    `gorm:"type:text" json:"user_agent"`
	DeviceID         string    `gorm:"type:varchar(255)" json:"device_id"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	IsActive         bool      `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

type Application struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	APIKey      string    `gorm:"uniqueIndex" json:"api_key"`
	OPAEndpoint string    `gorm:"not null" json:"opa_endpoint"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Roles       []ApplicationRole `gorm:"foreignKey:AppID" json:"roles,omitempty"`
	UserRoles   []UserAppRole     `gorm:"foreignKey:AppID" json:"user_roles,omitempty"`
	OPAPolicies []OPAPolicy       `gorm:"foreignKey:AppID" json:"opa_policies,omitempty"`
}

type ApplicationRole struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppID       uuid.UUID `gorm:"type:uuid;not null" json:"app_id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`

	// Relationships
	Application *Application `gorm:"foreignKey:AppID" json:"application,omitempty"`
}

type UserAppRole struct {
	UserID    uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"user_id"`
	AppID     uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"app_id"`
	RoleName  string    `gorm:"not null;primaryKey" json:"role_name"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User        *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Application *Application `gorm:"foreignKey:AppID" json:"application,omitempty"`
}

type OPAPolicy struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppID      uuid.UUID `gorm:"type:uuid;not null" json:"app_id"`
	Name       string    `gorm:"not null" json:"name"`
	RegoPolicy string    `gorm:"type:text;not null" json:"rego_policy"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`

	// Relationships
	Application *Application `gorm:"foreignKey:AppID" json:"application,omitempty"`
}

type PolicySyncStatus struct {
	AppID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"app_id"`
	Version  int       `json:"version"`
	SyncedAt time.Time `json:"synced_at"`

	// Relationships
	Application *Application `gorm:"foreignKey:AppID" json:"application,omitempty"`
}

// GORM hooks
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// BeforeCreate for Organization is now in organization.go
