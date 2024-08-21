package services

import (
	"org-forms-config-management/infra/database"
	"org-forms-config-management/models"
	"org-forms-config-management/models/requestModels"
	identityprovider "org-forms-config-management/services/identityprovider"
)

type SignUpService struct {
	identityService *IdentityProviderService
}

func (service *SignUpService) initialize() {
	// Initialize the service
	// Create a new transaction
	awsCognito := &identityprovider.AWSCognito{}
	service.identityService = NewIdentityProviderService(awsCognito)
}

func (service *SignUpService) SignUp(signUpData *requestModels.SignUp) (string, error) {
	// Implement the logic to save the sign up data to User, Organization, and SSOConfig tables
	// Return an error if the operation fails

	if service.identityService == nil {
		service.initialize()
	}

	if signUpData.UserType == "INDIVIDUAL" {
		// Create a new SSOConfig
		signUpData.Organization = requestModels.OrganizationInfo{
			Name:        signUpData.Name,
			EmailDomain: signUpData.Email,
			HasSSO:      false,
		}
		return registerOrganization(signUpData, service.identityService)

	} else {
		// Create a new SSOConfig
		return registerOrganization(signUpData, service.identityService)

	}

}

func registerOrganization(signUpData *requestModels.SignUp, identityService *IdentityProviderService) (string, error) {
	organisation := signUpData.Organization
	organisationDB := &models.Organization{
		Name:        organisation.Name,
		SSOType:     organisation.SSOOptions.SSOType,
		EmailDomain: organisation.EmailDomain,
	}

	// Add the organisation to the database
	err := database.DB.Create(&organisationDB).Error

	if err != nil {
		return "", err
	}

	if organisation.HasSSO {

		// Create a new SSOConfig
		ssoConfig := &models.SSOConfig{
			OrganizationID: organisationDB.ID,
			ClientID:       organisation.SSOOptions.ClientID,
			ClientSecret:   organisation.SSOOptions.ClientSecret,
			RedirectURL:    organisation.SSOOptions.RedirectURL,
		}

		err = database.DB.Create(&ssoConfig).Error

		if err != nil {
			return "", err
		}
	}

	return registerOrgAdminUser(signUpData, organisationDB.ID, identityService)
}

func registerOrgAdminUser(signUpData *requestModels.SignUp, orgId string, identityService *IdentityProviderService) (string, error) {

	userId, err := identityService.RegisterUser(signUpData.Email, signUpData.Password, signUpData.Name)

	if err != nil {
		return "", err
	}

	// Create a new User
	user := &models.User{
		Email:          signUpData.Email,
		ID:             userId,
		Name:           signUpData.Name,
		UserType:       signUpData.UserType,
		EmailVerified:  false,
		Role:           "ORG_ADMIN",
		OrganizationID: orgId,
	}

	// Add the user to the database
	err = database.DB.Create(&user).Error

	if err != nil {
		return "", err
	}

	return userId, nil
}

func (service *SignUpService) VerifyEmail(userEmail string, confirmationCode string) error {
	// Implement the logic to verify the email address
	// Return an error if the operation fails
	if service.identityService == nil {
		service.initialize()
	}

	// Call the VerifyEmail method of the IdentityProviderService

	err := service.identityService.VerifyEmail(userEmail, confirmationCode)

	if err != nil {
		return err
	}

	// Update the user's emailVerified field in the database
	err = database.DB.Model(&models.User{}).Where("email = ?", userEmail).Update("email_verified", true).Error

	if err != nil {
		return err
	}

	return nil
}

func (service *SignUpService) ResendVerificationCode(userEmail string) error {
	// Implement the logic to resend the confirmation code
	// Return an error if the operation fails
	if service.identityService == nil {
		service.initialize()
	}

	// Call the ResendConfirmationCode method of the IdentityProviderService
	err := service.identityService.ResendVerificationCode(userEmail)

	if err != nil {
		return err
	}

	return nil
}
