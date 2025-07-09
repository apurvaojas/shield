package auth

import (
	"context"
	"fmt" // For error wrapping

	appconfig "shield/cmd/app/config" // Updated import path
	"shield/modules/authn/internal/auth/nonce"
	authprovider "shield/modules/authn/internal/auth/provider" // Updated import path
	"shield/modules/authn/internal/auth/session"
	"shield/modules/authn/internal/models"
	"shield/modules/authn/internal/repository" // Add repository import

	"github.com/aws/aws-sdk-go-v2/aws" // Added for aws.String
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// AuthService provides methods for authentication.
type AuthService struct {
	provider       authprovider.AuthProvider
	config         *appconfig.Config
	userRepository repository.UserRepository
	sessionManager session.SessionManager
	nonceValidator nonce.NonceValidator
}

// NewAuthService creates a new AuthService.
func NewAuthService(provider authprovider.AuthProvider, cfg *appconfig.Config, userRepo repository.UserRepository, sessionMgr session.SessionManager, nonceVal nonce.NonceValidator) *AuthService {
	return &AuthService{
		provider:       provider,
		config:         cfg,
		userRepository: userRepo,
		sessionManager: sessionMgr,
		nonceValidator: nonceVal,
	}
}

// SignupUserRequest contains parameters for signing up a new user.
// These fields should align with the API contract defined in Signup_flow.md for /auth/signup
type SignupUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"` // Example: enforce min length via binding
	// Add other fields like name, phone_number if needed, and ensure they are part of UserAttributes
	// For example: GivenName, FamilyName, PhoneNumber etc.
}

// SignupUserResponse contains the result of a user signup.
type SignupUserResponse struct {
	UserID               string                                `json:"userID"` // Cognito User Sub
	RequiresConfirmation bool                                  `json:"requiresConfirmation"`
	CodeDeliveryDetails  *authprovider.CodeDeliveryDetailsData `json:"codeDeliveryDetails,omitempty"`
}

// SignupUser handles the registration of a new individual user.
func (s *AuthService) SignupUser(ctx context.Context, req SignupUserRequest) (*SignupUserResponse, error) {
	// Prepare user attributes for Cognito
	userAttributes := []types.AttributeType{
		{Name: aws.String("email"), Value: aws.String(req.Email)},
		{Name: aws.String("custom:user_type"), Value: aws.String(string(models.UserTypeIndividual))},
		// Add other attributes from req if necessary
	}

	providerReq := authprovider.SignUpRequestData{
		Username:       req.Email, // Using email as username for Cognito
		Password:       req.Password,
		Email:          req.Email,
		UserAttributes: userAttributes,
	}

	result, err := s.provider.SignUp(ctx, providerReq)
	if err != nil {
		return nil, fmt.Errorf("provider SignUp failed: %w", err)
	}

	// Create user record in local database
	user := &models.User{
		Email:      req.Email,
		CognitoSub: result.UserSub,
		UserType:   models.UserTypeIndividual,
		IsVerified: result.UserConfirmed, // Set verification status from Cognito response
		// OrgID will be uuid.Nil for individual users
	}

	if err := s.userRepository.CreateUser(ctx, user); err != nil {
		// Log the error but don't fail the signup - user is already created in Cognito
		// In production, you might want to implement compensation logic or queue for retry
		// For now, we'll log and continue
		fmt.Printf("Warning: Failed to create user in local database: %v\n", err)
	}

	return &SignupUserResponse{
		UserID:               result.UserSub,
		RequiresConfirmation: !result.UserConfirmed, // UserConfirmed is true if already confirmed (e.g. by admin)
		CodeDeliveryDetails:  result.CodeDeliveryDetails,
	}, nil
}

// ConfirmSignupRequest contains parameters for confirming a user's signup.
// Aligns with /auth/confirm
type ConfirmSignupRequest struct {
	Email            string `json:"email" binding:"required,email"`
	VerificationCode string `json:"verificationCode" binding:"required"`
}

// ConfirmSignupResponse contains the result of confirming a user's signup.
type ConfirmSignupResponse struct {
	Message string `json:"message"`
	// Typically, a successful confirmation might lead to session creation or MFA setup prompt.
}

// ConfirmUserSignup handles the confirmation of a user's account (e.g., email verification).
func (s *AuthService) ConfirmUserSignup(ctx context.Context, req ConfirmSignupRequest) (*ConfirmSignupResponse, error) {
	providerReq := authprovider.ConfirmSignUpRequestData{
		Username:         req.Email, // Assuming email was used as username
		ConfirmationCode: req.VerificationCode,
	}

	_, err := s.provider.ConfirmSignUp(ctx, providerReq)
	if err != nil {
		return nil, fmt.Errorf("provider ConfirmSignUp failed: %w", err)
	}

	// Update user status in local database
	user, err := s.userRepository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// User might not exist in local DB yet, log but don't fail
		fmt.Printf("Warning: User not found in local database during confirmation: %v\n", err)
	} else {
		// Update verification status
		user.IsVerified = true
		if err := s.userRepository.UpdateUser(ctx, user); err != nil {
			fmt.Printf("Warning: Failed to update user verification status in local database: %v\n", err)
		}
	}

	return &ConfirmSignupResponse{Message: "User confirmed successfully."}, nil
}

// SetupMFARequest contains parameters for initiating MFA setup.
type SetupMFARequest struct {
	UserID string           `json:"userID" binding:"required"` // Internal or Cognito User ID (usually Cognito Sub)
	Method models.MFAMethod `json:"method" binding:"required"` // e.g., "TOTP", "SMS"
}

// SetupMFAResponse contains data needed for the user to complete MFA setup.
type SetupMFAResponse struct {
	QRCodeURI string `json:"qrCodeUri,omitempty"` // For TOTP
	Secret    string `json:"secret,omitempty"`    // For TOTP, to display to the user as an alternative
	// For SMS, might include delivery details or just a success message
}

// SetupMFA initiates the MFA setup process for a user.
func (s *AuthService) SetupMFA(ctx context.Context, req SetupMFARequest) (*SetupMFAResponse, error) {
	// TODO:
	// 1. Call the identity provider (e.g., Cognito `AssociateSoftwareToken` or `SetUserMFAPreference` for SMS).
	//    - For TOTP (AssociateSoftwareToken), Cognito returns a SecretCode.
	//    - Construct QRCodeURI: fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", issuer, username, secret, issuer)
	// 2. Return the necessary information.

	// Placeholder implementation:
	if req.Method == models.MFAMethodTOTP {
		// This is a dummy response. Real values come from Cognito.
		return &SetupMFAResponse{
			QRCodeURI: "otpauth://totp/YourApp:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=YourApp",
			Secret:    "JBSWY3DPEHPK3PXP",
		}, nil
	} else if req.Method == models.MFAMethodSMS {
		// SMS MFA setup might involve verifying phone number first if not already done.
		// Cognito's SetUserMFAPreference handles enabling SMS MFA if phone_number_verified is true.
		return &SetupMFAResponse{}, nil // Or a message indicating SMS MFA setup initiated
	}
	return nil, fmt.Errorf("unsupported MFA method: %s", req.Method)
}

// VerifyMFARequest contains parameters for verifying an MFA code.
type VerifyMFARequest struct {
	UserID     string `json:"userID" binding:"required"` // Cognito User Sub
	MFACode    string `json:"mfaCode" binding:"required"`
	DeviceName string `json:"deviceName,omitempty"` // Optional, friendly name for the MFA device (especially for TOTP)
}

// VerifyMFAResponse indicates if MFA verification was successful.
type VerifyMFAResponse struct {
	Status string `json:"status"`
	// Could include session tokens if login is completed upon MFA verification.
}

// VerifyMFA verifies an MFA code provided by the user.
func (s *AuthService) VerifyMFA(ctx context.Context, req VerifyMFARequest) (*VerifyMFAResponse, error) {
	// TODO:
	// 1. Call the identity provider (e.g., Cognito `VerifySoftwareToken` for TOTP or handle SMS challenge response).
	// 2. If successful, Cognito usually marks the device/method as verified.
	// 3. If this is part of login, proceed to create a session.

	// Placeholder implementation:
	// For VerifySoftwareToken, if successful, status is SUCCESS.
	// If it's a new device, it might prompt to remember it.
	return &VerifyMFAResponse{Status: "verified"}, nil
}

// --- Organization Signup Flow ---

// OrgSignupRequest contains parameters for registering a new organization.
type OrgSignupRequest struct {
	OrgName       string `json:"orgName" binding:"required"`
	AdminEmail    string `json:"adminEmail" binding:"required,email"` // Email for the initial admin user
	AdminPassword string `json:"adminPassword,omitempty"`             // Optional: if creating a password-based admin
	// SSO details might be part of a separate step/request or included here if simple
	// e.g., IDPType (SAML/OIDC), IDPMetadataURL (for SAML)
}

// OrgSignupResponse contains the result of an organization signup.
type OrgSignupResponse struct {
	OrgID       string `json:"orgID"`
	AdminUserID string `json:"adminUserID"`
	SSOLoginURL string `json:"ssoLoginUrl,omitempty"` // e.g., /auth/sso/start?org=<orgID_or_name>
	Message     string `json:"message"`
}

// OrgSignup handles the registration of a new organization and its initial admin.
func (s *AuthService) OrgSignup(ctx context.Context, req OrgSignupRequest) (*OrgSignupResponse, error) {
	// 1. Create organization record in local DB first
	org := &models.Organization{
		Name: req.OrgName,
		// Set other fields as needed
	}

	if err := s.userRepository.CreateOrganization(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to create organization in database: %w", err)
	}

	// 2. Create admin user in Cognito with organization reference
	adminUserAttrs := []types.AttributeType{
		{Name: awsString("email"), Value: awsString(req.AdminEmail)},
		{Name: awsString("email_verified"), Value: awsString("true")}, // Or handle verification separately
		{Name: awsString("custom:org_id"), Value: awsString(org.ID.String())},
		{Name: awsString("custom:user_type"), Value: awsString("organization_admin")},
	}
	adminReq := authprovider.AdminCreateUserRequestData{
		Username:          req.AdminEmail,
		Email:             req.AdminEmail,
		TemporaryPassword: req.AdminPassword, // If empty, Cognito might require a different flow or not set password
		UserAttributes:    adminUserAttrs,
	}
	if req.AdminPassword == "" {
		// Handle case where admin password is not provided - e.g. set a temporary one or use a different flow
		adminReq.TemporaryPassword = "TempPass123!" // Example temporary password
	}

	adminCognitoUser, err := s.provider.AdminCreateUser(ctx, adminReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user in Cognito: %w", err)
	}

	adminUserID := ""
	if adminCognitoUser != nil && adminCognitoUser.User != nil {
		adminUserID = adminCognitoUser.User.CognitoSub // Use actual Cognito Sub

		// 3. Create admin user record in local database
		adminUser := &models.User{
			Email:      req.AdminEmail,
			CognitoSub: adminCognitoUser.User.CognitoSub,
			OrgID:      org.ID,
			UserType:   models.UserTypeOrganization, // or create a specific admin type if needed
			IsVerified: true,                        // Admin users are typically verified immediately
		}

		if err := s.userRepository.CreateUser(ctx, adminUser); err != nil {
			// Log the error but don't fail the signup - user is already created in Cognito
			fmt.Printf("Warning: Failed to create admin user in local database: %v\n", err)
		}
	}

	return &OrgSignupResponse{
		OrgID:       org.ID.String(),
		AdminUserID: adminUserID,
		Message:     "Organization and admin user created successfully. SSO setup may be required.",
		// SSOLoginURL can be constructed later when SSO is configured
	}, nil
}

// Ensure models.UserTypeIndividual is accessible
var _ = models.UserTypeIndividual

// Helper to get aws.String without importing aws package directly in all service files
func awsString(s string) *string {
	if s == "" {
		return nil // Or handle as per AWS SDK's expectation for empty strings
	}
	return &s
}

// LoginRequest contains parameters for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse contains the result of a user login
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	SessionID    string `json:"sessionId"`
	UserID       string `json:"userId"`
}

// RefreshTokenRequest contains parameters for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RefreshTokenResponse contains the result of token refresh
type RefreshTokenResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"`
}

// Login authenticates a user and creates a session
func (s *AuthService) Login(ctx context.Context, req LoginRequest, clientInfo session.ClientInfo) (*LoginResponse, error) {
	// Authenticate with provider (Cognito)
	authReq := authprovider.AuthenticateRequestData{
		Username: req.Email,
		Password: req.Password,
	}

	authResult, err := s.provider.Authenticate(ctx, authReq)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Get user from database
	user, err := s.userRepository.GetUserByCognitoSub(ctx, authResult.UserSub)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Create session
	sessionData, err := s.sessionManager.CreateSession(ctx, user.ID, clientInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResponse{
		AccessToken:  authResult.AccessToken,
		RefreshToken: sessionData.RefreshToken,
		ExpiresIn:    int(authResult.ExpiresIn),
		SessionID:    sessionData.ID,
		UserID:       user.ID.String(),
	}, nil
}

// RefreshToken refreshes an access token using refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	// Find session by refresh token
	// Note: This is a simplified approach. In production, you might want to hash refresh tokens
	// or use a more secure method to link refresh tokens to sessions

	// For now, we'll need to add a method to find session by refresh token
	// This would require updating the session repository interface

	refreshReq := authprovider.RefreshTokenRequestData{
		RefreshToken: req.RefreshToken,
	}

	refreshResult, err := s.provider.RefreshToken(ctx, refreshReq)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	return &RefreshTokenResponse{
		AccessToken: refreshResult.AccessToken,
		ExpiresIn:   int(refreshResult.ExpiresIn),
	}, nil
}

// Logout invalidates a user session
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	// Invalidate session
	if err := s.sessionManager.InvalidateSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	// Optionally, you could also revoke the token from the provider
	// This would require implementing a revoke method in the provider interface

	return nil
}
