package dto

// MFASetupRequest represents the request body for MFA setup
type MFASetupRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Method string `json:"method" binding:"required"` // e.g., "TOTP", "SMS"
}

// MFASetupResponse represents the response for MFA setup
type MFASetupResponse struct {
	QRCodeURI string `json:"qr_code_uri,omitempty"` // For TOTP
	Secret    string `json:"secret,omitempty"`      // For TOTP
	// For SMS, might just be a confirmation message
}

// MFAVerifyRequest represents the request body for MFA verification
type MFAVerifyRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Code   string `json:"code" binding:"required"`
}
