package models

type RoleEnum string

const (
	OrgAdmin     RoleEnum = "ORG_ADMIN"
	ProjectAdmin RoleEnum = "PROJECT_ADMIN"
	NormalUser   RoleEnum = "NORMAL_USER"
)

type AccessType string

const (
	RWAccess       RoleEnum = "RW_ACCESS"
	ReadOnlyAccess RoleEnum = "READ_ONLY_ACCESS"
)

type SSOType string

const (
	SAML SSOType = "SAML"
	OIDC SSOType = "OIDC"
)

type UserType string

const (
	Individual UserType = "INDIVIDUAL"
	Org        UserType = "ORGANIZATION"
)

type Organization struct {
	ID          string    `json:"id" gorm:"primary_key; type:uuid;default:uuid_generate_v4();"`
	Name        string    `json:"name"`
	EmailDomain string    `json:"emailDomain"`
	SSOType     SSOType   `json:"ssoType" gorm:"default:null"`
	SSOConfigs  SSOConfig `gorm:"foreignKey:OrganizationID"`
	Users       []User    `gorm:"foreignKey:OrganizationID"`
	Projects    []Project `gorm:"foreignKey:OrganizationID"`
}

type SSOConfig struct {
	ID             string `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	OrganizationID string `json:"organizationId"`
	ClientID       string `json:"clientId"`
	ClientSecret   string `json:"clientSecret"`
	RedirectURL    string `json:"redirectURL"`
}

type User struct {
	ID             string       `json:"id" gorm:"primary_key;type:uuid"`
	Email          string       `json:"email" gorm:"unique"`
	Name           string       `json:"name"`
	UserType       UserType     `json:"userType"`
	Role           RoleEnum     `json:"role"`
	EmailVerified  bool         `json:"emailVerified"`
	OrganizationID string       `json:"organizationId"`
	Organization   Organization `gorm:"foreignKey:OrganizationID;references:ID"`
	Accesses       []Access     `gorm:"foreignKey:UserID"`
}

//

type Project struct {
	ID             string        `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name           string        `json:"name"`
	OrganizationID string        `json:"organizationId"`
	Organization   Organization  `gorm:"foreignKey:OrganizationID;references:ID"`
	Environments   []Environment `gorm:"foreignKey:ProjectID"`
	Variants       []Variant     `gorm:"foreignKey:ProjectID"`
	Accesses       []Access      `gorm:"foreignKey:ProjectID"`
}

type Environment struct {
	ID          string       `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name        string       `json:"name"`
	ProjectID   string       `json:"projectId"`
	Domain      string       `json:"domain"`
	Project     Project      `gorm:"foreignKey:ProjectID;references:ID"`
	JsonConfigs []JsonConfig `gorm:"foreignKey:EnvironmentID"`
	Accesses    []Access     `gorm:"foreignKey:EnvironmentID"`
}

type Variant struct {
	ID          string       `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	ProjectID   string       `json:"projectId"`
	Project     Project      `gorm:"foreignKey:ProjectID;references:ID"`
	JsonConfigs []JsonConfig `gorm:"foreignKey:VariantID"`
}

type JsonConfig struct {
	ID            string      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name          string      `json:"name"`
	Content       string      `json:"content"`
	EnvironmentID string      `json:"environmentId"`
	VariantID     string      `json:"variantId"`
	Environment   Environment `gorm:"foreignKey:EnvironmentID;references:ID"`
	Variant       Variant     `gorm:"foreignKey:VariantID;references:ID"`
}

type Access struct {
	ID            string      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	UserID        string      `json:"userId"`
	ProjectID     string      `json:"projectId"`
	EnvironmentID string      `json:"environmentId"`
	Access        AccessType  `json:"access"`
	User          User        `gorm:"foreignKey:UserID;references:ID"`
	Project       Project     `gorm:"foreignKey:ProjectID;references:ID"`
	Environment   Environment `gorm:"foreignKey:EnvironmentID;references:ID"`
}
