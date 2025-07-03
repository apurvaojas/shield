package provider

import (
	"context"

	cognitoTypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/tentackles/shield/modules/authn/internal/models"
)

// --- Request & Response Structs for AuthProvider interface ---
// These structs are based on Cognito's needs but can be adapted if other providers are added.

// SignUpRequestData holds data for signing up a user.
type SignUpRequestData struct {
	Username       string
	Password       string
	Email          string
	UserAttributes []cognitoTypes.AttributeType // Using Cognito's type for now
}

// SignUpOutputData holds data returned after a successful signup.
// This can be expanded based on what Cognito's SignUpOutput returns that is useful.
type SignUpOutputData struct {
	UserSub             string // The unique identifier for the user (e.g., Cognito Sub)
	CodeDeliveryDetails *CodeDeliveryDetailsData
	UserConfirmed       bool
}

// CodeDeliveryDetailsData mirrors Cognito's CodeDeliveryDetailsType.
type CodeDeliveryDetailsData struct {
	AttributeName  string
	DeliveryMedium string // EMAIL or SMS
	Destination    string
}

// ConfirmSignUpRequestData holds data for confirming a user's signup.
type ConfirmSignUpRequestData struct {
	Username         string
	ConfirmationCode string
}

// ConfirmSignUpOutputData holds data after confirming signup.
type ConfirmSignUpOutputData struct {
	// Typically empty for Cognito, but can be defined for consistency
}

// AdminCreateUserRequestData holds data for an admin creating a user.
type AdminCreateUserRequestData struct {
	Username          string
	Email             string
	TemporaryPassword string
	UserAttributes    []cognitoTypes.AttributeType
}

// AdminCreateUserOutputData holds data after an admin creates a user.
type AdminCreateUserOutputData struct {
	User *models.User // Or a simplified user representation
}

// GetUserOutputData holds data for a retrieved user.
type GetUserOutputData struct {
	User *models.User
}

// CreateIdentityProviderRequestData holds data for creating an IdP.
type CreateIdentityProviderRequestData struct {
	ProviderName     string
	ProviderType     string // e.g., "SAML", "OIDC"
	ProviderDetails  map[string]string
	AttributeMapping map[string]string
	IdpIdentifiers   []string
}

// CreateIdentityProviderOutputData holds data after creating an IdP.
type CreateIdentityProviderOutputData struct {
	IdentityProvider *cognitoTypes.IdentityProviderType // Using Cognito's type for now
}

// AuthenticateRequestData holds data for user authentication
type AuthenticateRequestData struct {
	Username string
	Password string
}

// AuthenticateOutputData holds data returned after successful authentication
type AuthenticateOutputData struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	UserSub      string
}

// RefreshTokenRequestData holds data for token refresh
type RefreshTokenRequestData struct {
	RefreshToken string
}

// RefreshTokenOutputData holds data returned after token refresh
type RefreshTokenOutputData struct {
	AccessToken string
	ExpiresIn   int64
}

// AuthProvider defines the interface for authentication operations.
type AuthProvider interface {
	SignUp(ctx context.Context, req SignUpRequestData) (*SignUpOutputData, error)
	ConfirmSignUp(ctx context.Context, req ConfirmSignUpRequestData) (*ConfirmSignUpOutputData, error)
	AdminCreateUser(ctx context.Context, req AdminCreateUserRequestData) (*AdminCreateUserOutputData, error)
	GetUser(ctx context.Context, accessToken string) (*GetUserOutputData, error) // Or by other means like user ID/sub

	// Authentication methods
	Authenticate(ctx context.Context, req AuthenticateRequestData) (*AuthenticateOutputData, error)
	RefreshToken(ctx context.Context, req RefreshTokenRequestData) (*RefreshTokenOutputData, error)

	// Methods for Organization Signup Flow (SSO)
	CreateIdentityProvider(ctx context.Context, req CreateIdentityProviderRequestData) (*CreateIdentityProviderOutputData, error)
	// TODO: Add other methods as needed:
	// UpdateUserPool(...)
	// CreateUserPoolDomain(...)
	// UpdateUserPoolClient(...)

	// Methods for Login Flow
	// InitiateAuth(ctx context.Context, req InitiateAuthRequestData) (*InitiateAuthOutputData, error)
	// RespondToAuthChallenge(ctx context.Context, req RespondToAuthChallengeRequestData) (*RespondToAuthChallengeOutputData, error)
}
