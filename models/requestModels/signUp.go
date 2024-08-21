package requestModels

import (
	models "org-forms-config-management/models"
)

type SignUp struct {
	Email        string           `json:"email" binding:"required,email"`
	Password     string           `json:"password" binding:"required"`
	Name         string           `json:"name" binding:"required"`
	UserType     models.UserType  `json:"userType" binding:"required"`
	Organization OrganizationInfo `json:"organizationInfo" validation:"required_if=UserType ORGANIZATION"`
}

type OrganizationInfo struct {
	HasSSO      bool       `json:"hasSSO" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	EmailDomain string     `json:"emailDomain" binding:"required"`
	SSOOptions  SSOOptions `json:"ssoOptions" binding:"required_if=HasSSO true"`
}

type SSOOptions struct {
	SSOType models.SSOType `json:"ssoType" binding:"required"`
}

type OIDCConfigs struct {
	ClientID              string   `json:"clientId" binding:"required"`
	ClientSecret          string   `json:"clientSecret" binding:"required"`
	Scopes                []string `json:"scopes" binding:"required"`
	IssuerUrl             string   `json:"issuerUrl" binding:"required"`
	AuthorizationEndpoint string   `json:"authorization" binding:"required"`
	TokenEndpoint         string   `json:"token" binding:"required"`
	UserInfoEndpoint      string   `json:"userInfo" binding:"required"`
	JwksUriEndpoints      string   `json:"jwksUriEndpoint" binding:"required"`
}

type SAMLConfigs struct {
	RequestSigningAlgorithm string `json:"requestSigningAlgorithm" binding:"required"`
	MetadataURL             string `json:"metadataUrl"`
	MetadataFile            string `json:"metadataFile"`
	EncryptedResponses      bool   `json:"encryptedResponses"`
}

type VerifyEmail struct {
	UserEmail        string `json:"email" binding:"required,email"`
	ConfirmationCode string `json:"confirmationCode" binding:"required"`
}

type ResendVerificationCode struct {
	UserEmail string `json:"email" binding:"required,email"`
}
