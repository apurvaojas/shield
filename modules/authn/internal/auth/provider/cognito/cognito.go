package cognito

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"

	authprovider "github.com/tentackles/shield/modules/authn/internal/auth/provider" // Updated import path
	appConfig "github.com/tentackles/shield/modules/authn/internal/config"
	"github.com/tentackles/shield/modules/authn/internal/models" // Assuming user models are here
)

// Provider implements authentication logic using AWS Cognito.
type Provider struct {
	client *cognitoidentityprovider.Client
	config appConfig.CognitoConfig
}

// NewProvider creates a new Cognito authentication provider.
func NewProvider(cfg appConfig.CognitoConfig) (*Provider, error) {
	sdkConfig, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(cfg.Region))
	if err != nil {
		log.Printf("Failed to load AWS SDK config: %v", err)
		return nil, err
	}

	client := cognitoidentityprovider.NewFromConfig(sdkConfig)

	return &Provider{
		client: client,
		config: cfg,
	}, nil
}

// --- AWS Cognito API Reference: https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_Operations.html ---

// SignUp registers a new user with Cognito.
// Docs: https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_SignUp.html
func (p *Provider) SignUp(ctx context.Context, req authprovider.SignUpRequestData) (*authprovider.SignUpOutputData, error) {
	input := &cognitoidentityprovider.SignUpInput{
		ClientId:       aws.String(p.config.AppClientID),
		Username:       aws.String(req.Username),
		Password:       aws.String(req.Password),
		UserAttributes: req.UserAttributes,
	}

	if p.config.AppClientSecret != "" {
		// TODO: Calculate SecretHash if client secret is used
		log.Println("Warning: AppClientSecret is configured but SecretHash computation is not yet implemented for SignUp.")
	}

	result, err := p.client.SignUp(ctx, input)
	if err != nil {
		log.Printf("Cognito SignUp error: %v", err)
		return nil, err
	}

	output := &authprovider.SignUpOutputData{
		UserSub:       aws.ToString(result.UserSub),
		UserConfirmed: result.UserConfirmed,
	}
	if result.CodeDeliveryDetails != nil {
		output.CodeDeliveryDetails = &authprovider.CodeDeliveryDetailsData{
			AttributeName:  aws.ToString(result.CodeDeliveryDetails.AttributeName),
			DeliveryMedium: string(result.CodeDeliveryDetails.DeliveryMedium),
			Destination:    aws.ToString(result.CodeDeliveryDetails.Destination),
		}
	}
	return output, nil
}

// ConfirmSignUp confirms a user's registration using a confirmation code.
// Docs: https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_ConfirmSignUp.html
func (p *Provider) ConfirmSignUp(ctx context.Context, req authprovider.ConfirmSignUpRequestData) (*authprovider.ConfirmSignUpOutputData, error) {
	input := &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(p.config.AppClientID),
		Username:         aws.String(req.Username),
		ConfirmationCode: aws.String(req.ConfirmationCode),
	}

	if p.config.AppClientSecret != "" {
		// TODO: Calculate SecretHash
		log.Println("Warning: AppClientSecret is configured but SecretHash computation is not yet implemented for ConfirmSignUp.")
	}

	_, err := p.client.ConfirmSignUp(ctx, input)
	if err != nil {
		log.Printf("Cognito ConfirmSignUp error: %v", err)
		return nil, err
	}
	return &authprovider.ConfirmSignUpOutputData{}, nil
}

// AdminCreateUser creates a user as an administrator.
// Docs: https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_AdminCreateUser.html
func (p *Provider) AdminCreateUser(ctx context.Context, req authprovider.AdminCreateUserRequestData) (*authprovider.AdminCreateUserOutputData, error) {
	input := &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:        aws.String(p.config.UserPoolID),
		Username:          aws.String(req.Username),
		UserAttributes:    req.UserAttributes,
		TemporaryPassword: aws.String(req.TemporaryPassword),
		MessageAction:     types.MessageActionTypeSuppress,
	}

	result, err := p.client.AdminCreateUser(ctx, input)
	if err != nil {
		log.Printf("Cognito AdminCreateUser error: %v", err)
		return nil, err
	}

	// Map Cognito UserType to models.User
	var userModel *models.User
	if result.User != nil {
		userModel = &models.User{
			Email:      req.Email, // Assuming username is email for AdminCreateUser
			CognitoSub: aws.ToString(result.User.Username),
		}
		for _, attr := range result.User.Attributes {
			if aws.ToString(attr.Name) == "email" {
				userModel.Email = aws.ToString(attr.Value)
			}
		}
	}

	return &authprovider.AdminCreateUserOutputData{
		User: userModel,
	}, nil
}

// CreateIdentityProvider creates an identity provider in Cognito.
// Docs: https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateIdentityProvider.html
func (p *Provider) CreateIdentityProvider(ctx context.Context, req authprovider.CreateIdentityProviderRequestData) (*authprovider.CreateIdentityProviderOutputData, error) {
	input := &cognitoidentityprovider.CreateIdentityProviderInput{
		ProviderName:     aws.String(req.ProviderName),
		ProviderType:     types.IdentityProviderTypeType(req.ProviderType),
		UserPoolId:       aws.String(p.config.UserPoolID),
		ProviderDetails:  req.ProviderDetails,
		AttributeMapping: req.AttributeMapping,
		IdpIdentifiers:   req.IdpIdentifiers,
	}

	result, err := p.client.CreateIdentityProvider(ctx, input)
	if err != nil {
		log.Printf("Cognito CreateIdentityProvider error: %v", err)
		return nil, err
	}

	return &authprovider.CreateIdentityProviderOutputData{
		IdentityProvider: result.IdentityProvider,
	}, nil
}

// GetUser retrieves user information based on an access token.
// Docs: https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_GetUser.html
func (p *Provider) GetUser(ctx context.Context, accessToken string) (*authprovider.GetUserOutputData, error) {
	input := &cognitoidentityprovider.GetUserInput{AccessToken: aws.String(accessToken)}
	result, err := p.client.GetUser(ctx, input)
	if err != nil {
		log.Printf("Cognito GetUser error: %v", err)
		return nil, err
	}

	userModel := &models.User{
		CognitoSub: aws.ToString(result.Username),
	}
	for _, attr := range result.UserAttributes {
		if aws.ToString(attr.Name) == "email" {
			userModel.Email = aws.ToString(attr.Value)
		}
	}

	log.Println("GetUser partially implemented for Cognito provider. Mapping to models.User needs review.")
	return &authprovider.GetUserOutputData{User: userModel}, nil
}

// Authenticate authenticates a user with username and password
// Docs: https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_InitiateAuth.html
func (p *Provider) Authenticate(ctx context.Context, req authprovider.AuthenticateRequestData) (*authprovider.AuthenticateOutputData, error) {
	input := &cognitoidentityprovider.InitiateAuthInput{
		ClientId: aws.String(p.config.AppClientID),
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		AuthParameters: map[string]string{
			"USERNAME": req.Username,
			"PASSWORD": req.Password,
		},
	}

	// Add client secret if configured
	if p.config.AppClientSecret != "" {
		secretHash := computeSecretHash(req.Username, p.config.AppClientID, p.config.AppClientSecret)
		input.AuthParameters["SECRET_HASH"] = secretHash
	}

	result, err := p.client.InitiateAuth(ctx, input)
	if err != nil {
		return nil, err
	}

	// Handle potential challenges (MFA, etc.)
	if result.ChallengeName != "" {
		// For now, return an error if there are challenges
		// In a full implementation, you'd handle these challenges
		return nil, fmt.Errorf("authentication challenge required: %s", result.ChallengeName)
	}

	if result.AuthenticationResult == nil {
		return nil, fmt.Errorf("authentication failed: no result")
	}

	// Get user info to extract user sub
	userResult, err := p.GetUser(ctx, *result.AuthenticationResult.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &authprovider.AuthenticateOutputData{
		AccessToken:  *result.AuthenticationResult.AccessToken,
		RefreshToken: *result.AuthenticationResult.RefreshToken,
		ExpiresIn:    int64(result.AuthenticationResult.ExpiresIn),
		UserSub:      userResult.User.CognitoSub,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
// Docs: https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_InitiateAuth.html
func (p *Provider) RefreshToken(ctx context.Context, req authprovider.RefreshTokenRequestData) (*authprovider.RefreshTokenOutputData, error) {
	input := &cognitoidentityprovider.InitiateAuthInput{
		ClientId: aws.String(p.config.AppClientID),
		AuthFlow: types.AuthFlowTypeRefreshTokenAuth,
		AuthParameters: map[string]string{
			"REFRESH_TOKEN": req.RefreshToken,
		},
	}

	// Add client secret if configured
	if p.config.AppClientSecret != "" {
		// For refresh token flow, we don't have the username, so we can't compute SECRET_HASH
		// This is a limitation when using client secret with refresh tokens
		// Consider using a different approach or storing username with refresh token
	}

	result, err := p.client.InitiateAuth(ctx, input)
	if err != nil {
		return nil, err
	}

	if result.AuthenticationResult == nil {
		return nil, fmt.Errorf("token refresh failed: no result")
	}

	return &authprovider.RefreshTokenOutputData{
		AccessToken: *result.AuthenticationResult.AccessToken,
		ExpiresIn:   int64(result.AuthenticationResult.ExpiresIn),
	}, nil
}

// computeSecretHash computes the secret hash for Cognito client authentication
func computeSecretHash(username, clientID, clientSecret string) string {
	message := username + clientID
	key := []byte(clientSecret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// --- Placeholders for additional Cognito API operations (see AWS docs) ---
// ListUsers, AdminUpdateUserAttributes, AdminDeleteUser, AdminGetUser, etc.
// Example:
// // ListUsers lists users in the user pool.
// // Docs: https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_ListUsers.html
// func (p *Provider) ListUsers(ctx context.Context, filter string) (*cognitoidentityprovider.ListUsersOutput, error) {
// 	input := &cognitoidentityprovider.ListUsersInput{
// 		UserPoolId: aws.String(p.config.UserPoolID),
// 		Filter:     aws.String(filter),
// 	}
// 	return p.client.ListUsers(ctx, input)
// }
//
// // AdminUpdateUserAttributes updates user attributes as an admin.
// // Docs: https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_AdminUpdateUserAttributes.html
// func (p *Provider) AdminUpdateUserAttributes(ctx context.Context, username string, attrs []types.AttributeType) error {
// 	input := &cognitoidentityprovider.AdminUpdateUserAttributesInput{
// 		UserPoolId:     aws.String(p.config.UserPoolID),
// 		Username:       aws.String(username),
// 		UserAttributes: attrs,
// 	}
// 	_, err := p.client.AdminUpdateUserAttributes(ctx, input)
// 	return err
// }

// Ensure Provider implements the AuthProvider interface.
var _ authprovider.AuthProvider = (*Provider)(nil)
