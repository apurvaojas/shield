package dto

// SignupRequest represents the request body for user signup
type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// SignupResponse represents the response for a successful signup
type SignupResponse struct {
	UserID               string               `json:"user_id"`
	RequiresConfirmation bool                 `json:"requires_confirmation,omitempty"`
	Message              string               `json:"message,omitempty"`
	CodeDeliveryDetails  *CodeDeliveryDetails `json:"code_delivery_details,omitempty"`
}

// CodeDeliveryDetails represents the delivery method for verification codes
type CodeDeliveryDetails struct {
	AttributeName  string `json:"attribute_name"`
	DeliveryMedium string `json:"delivery_medium"` // EMAIL or SMS
	Destination    string `json:"destination"`
}

// ConfirmSignupRequest represents the request body for confirming signup
type ConfirmSignupRequest struct {
	Email            string `json:"email" binding:"required,email"`
	VerificationCode string `json:"verification_code" binding:"required"`
}
