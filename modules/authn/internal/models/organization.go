package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IDPType represents the type of Identity Provider.
type IDPType string

const (
	IDPTypeSAML  IDPType = "SAML"
	IDPTypeOIDC  IDPType = "OIDC"
	IDPTypeOther IDPType = "Other"
)

// Organization represents an organization entity in the system.
type Organization struct {
	ID                  uuid.UUID         `gorm:"type:uuid;primary_key;" json:"id"`
	Name                string            `gorm:"type:varchar(255);not null" json:"name"`
	SSOProviderName     string            `gorm:"type:varchar(255)" json:"ssoProviderName,omitempty"` // Name of the IdP registered in Cognito (e.g., "AcmeOktaSAML")
	IDPType             IDPType           `gorm:"type:varchar(50)" json:"idpType,omitempty"`          // Type of IdP (e.g., "SAML", "OIDC")
	CallbackURL         string            `gorm:"type:varchar(255)" json:"callbackUrl,omitempty"`     // Callback URL for this org's SSO, if specific
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
	DeletedAt           gorm.DeletedAt    `gorm:"index" json:"-"`                                                    // For soft deletes
	Users               []User            `gorm:"foreignKey:OrgID" json:"-"`                                         // Users belonging to this organization
	SSOProviderConfigID uuid.UUID         `gorm:"type:uuid;" json:"ssoProviderConfigId,omitempty"`                   // Foreign key to SSOProviderConfig
	SSOProviderConfig   SSOProviderConfig `gorm:"foreignKey:SSOProviderConfigID" json:"ssoProviderConfig,omitempty"` // Associated SSOProviderConfig
}

// BeforeCreate will set a UUID rather than relying on default database UUID generation.
func (org *Organization) BeforeCreate(tx *gorm.DB) (err error) {
	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}
	return
}

// SSOProviderConfig represents the SSO configurations for an organization.
type SSOProviderConfig struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	MetadataURL  string    `gorm:"type:varchar(255)" json:"metadataUrl,omitempty"`  // Metadata URL for SAML
	ClientID     string    `gorm:"type:varchar(255)" json:"clientId,omitempty"`     // Client ID for OIDC
	ClientSecret string    `gorm:"type:varchar(255)" json:"clientSecret,omitempty"` // Client Secret for OIDC
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
