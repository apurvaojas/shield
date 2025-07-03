package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Organization represents an organization entity in the system.
type Organization struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	SSOProviderName string `gorm:"type:varchar(255)" json:"ssoProviderName,omitempty"` // Name of the IdP registered in Cognito (e.g., "AcmeOktaSAML")
	IDPType     string    `gorm:"type:varchar(50)" json:"idpType,omitempty"`         // Type of IdP (e.g., "SAML", "OIDC")
	CallbackURL string    `gorm:"type:varchar(255)" json:"callbackUrl,omitempty"`    // Callback URL for this org's SSO, if specific
	// Add other IdP specific metadata if needed, e.g., EntityID, MetadataURL for SAML,
	// or store them in a separate table linked to Organization if complex.
	// The Signup_flow.md suggests a JSON blob for IdP mappings, which could be stored here
	// or in a related table. For simplicity, keeping core fields here.
	// Example: IDPMetadata jsonb `gorm:"type:jsonb" json:"idpMetadata,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // For soft deletes

	Users []User `gorm:"foreignKey:OrgID" json:"-"` // Users belonging to this organization
}

// BeforeCreate will set a UUID rather than relying on default database UUID generation.
func (org *Organization) BeforeCreate(tx *gorm.DB) (err error) {
	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}
	return
}
