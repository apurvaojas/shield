package services

import (
	"org-forms-config-management/services/identityprovider"
)

type IdentityProviderService struct {
    provider identityprovider.IdentityProvider
}

func NewIdentityProviderService(provider identityprovider.IdentityProvider) *IdentityProviderService {
    return &IdentityProviderService{provider: provider}
}

func (service *IdentityProviderService) RegisterUser(userName string, password string, name string) (string, error) {
    return service.provider.RegisterUser(userName, password, name)
}

func (service *IdentityProviderService) VerifyEmail(username string, confirmationCode string) error {
    return service.provider.VerifyEmail(username, confirmationCode)
}

func (service *IdentityProviderService) ResendVerificationCode(username string) error {
    return service.provider.ResendVerificationCode(username)
}