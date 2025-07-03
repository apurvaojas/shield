package dto

// OrgSignupRequest represents the request body for organization signup
type OrgSignupRequest struct {
	OrgName       string `json:"org_name" binding:"required"`
	AdminEmail    string `json:"admin_email" binding:"required,email"`
	AdminPassword string `json:"admin_password,omitempty"`
}

// OrgSignupResponse represents the response for organization signup
type OrgSignupResponse struct {
	OrgID       string `json:"org_id"`
	AdminUserID string `json:"admin_user_id"`
	Message     string `json:"message"`
}

// OrgDetails represents organization details
type OrgDetails struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SSOProvider string `json:"sso_provider,omitempty"`
	IDPType     string `json:"idp_type,omitempty"`
	CallbackURL string `json:"callback_url,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// UpdateOrgRequest represents the request body for updating organization
type UpdateOrgRequest struct {
	Name        string `json:"name,omitempty"`
	SSOProvider string `json:"sso_provider,omitempty"`
	IDPType     string `json:"idp_type,omitempty"`
	CallbackURL string `json:"callback_url,omitempty"`
}
