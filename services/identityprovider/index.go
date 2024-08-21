package identityprovider

type IdentityProvider interface {
    RegisterUser(userName string, password string, name string) (string, error)
    VerifyEmail(username string, confirmationCode string) error
    ResendVerificationCode(username string) error
}