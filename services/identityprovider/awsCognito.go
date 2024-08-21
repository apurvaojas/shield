package identityprovider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"org-forms-config-management/models/requestModels"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	// "github.com/spf13/viper"
)

type AWSCognito struct {
	cognitoClient *cognitoidentityprovider.Client
}

// NewAWSCognito creates a new instance of AWSCognito.
func (awsCognito *AWSCognito) init() (string, error) {
	// viper.AutomaticEnv()
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return "", err
	}

	awsCognito.cognitoClient = cognitoidentityprovider.NewFromConfig(sdkConfig)
	return "", nil
}

func (awsCognito *AWSCognito) ResendVerificationCode(username string) error {
	err := error(nil)
	if awsCognito.cognitoClient == nil {
		_, err = awsCognito.init()
		if err != nil {
			return err
		}
	}

	secretHash, err := generateSecretHash(username, "5vf304hht0uhhf1jo7ql0asb5p", "kfoafsftrtpbjbig4o8kg6pp04s6uam6lmeupv54s1f58o3serb")

	if err != nil {
		fmt.Println("Error generating secret hash:", err)
		return err
	}
	_, err = awsCognito.cognitoClient.ResendConfirmationCode(context.TODO(), &cognitoidentityprovider.ResendConfirmationCodeInput{
		ClientId:   aws.String("5vf304hht0uhhf1jo7ql0asb5p"),
		Username:   aws.String(username),
		SecretHash: &secretHash,
	})
	if err != nil {
		fmt.Println("Couldn't resend confirmation code for user", username)
		fmt.Println(err)
	}
	return err
}

// VerifyEmail implements IdentityProvider.
func (awsCognito *AWSCognito) VerifyEmail(userEmail string, confirmationCode string) error {

	err := error(nil)
	if awsCognito.cognitoClient == nil {
		_, err = awsCognito.init()
		if err != nil {
			return err
		}
	}
	secretHash, err := generateSecretHash(userEmail, "5vf304hht0uhhf1jo7ql0asb5p", "kfoafsftrtpbjbig4o8kg6pp04s6uam6lmeupv54s1f58o3serb")

	if err != nil {
		fmt.Println("Error generating secret hash:", err)
		return err
	}

	_, err = awsCognito.cognitoClient.ConfirmSignUp(context.TODO(), &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String("5vf304hht0uhhf1jo7ql0asb5p"),
		ConfirmationCode: aws.String(confirmationCode),
		Username:         aws.String(userEmail),
		SecretHash:       &secretHash,
	})
	if err != nil {
		fmt.Println("Couldn't confirm sign up for user", userEmail)
		fmt.Println(err)
	}
	return err
}

func (awsCognito *AWSCognito) RegisterUser(userEmail string, password string, name string) (string, error) {

	userId, err := "", error(nil)
	if awsCognito.cognitoClient == nil {
		userId, err = awsCognito.init()
		if err != nil {
			return userId, err
		}
	}
	secretHash, err := generateSecretHash(userEmail, "5vf304hht0uhhf1jo7ql0asb5p", "kfoafsftrtpbjbig4o8kg6pp04s6uam6lmeupv54s1f58o3serb")
	if err != nil {
		fmt.Println("Error generating secret hash:", err)
		return userId, err
	}
	output, err := awsCognito.cognitoClient.SignUp(context.TODO(), &cognitoidentityprovider.SignUpInput{
		ClientId:   aws.String("5vf304hht0uhhf1jo7ql0asb5p"),
		Password:   aws.String(password),
		Username:   aws.String(userEmail),
		SecretHash: aws.String(secretHash),

		UserAttributes: []types.AttributeType{
			{Name: aws.String("email"), Value: aws.String(userEmail)},
			{
				Name:  aws.String("picture"),
				Value: aws.String("NA"),
			},
			{
				Name:  aws.String("name"),
				Value: aws.String(name),
			},
		},
	})
	if err != nil {
		var invalidPassword *types.InvalidPasswordException
		if errors.As(err, &invalidPassword) {
			log.Println(*invalidPassword.Message)
		} else {
			log.Printf("Couldn't sign up user %v. Here's why: %v\n", userEmail, err)
		}
	} else {
		userId = *output.UserSub
	}
	return userId, err
}

func generateSecretHash(username, clientId, clientSecret string) (string, error) {
	// Create a new HMAC by defining the hash type and the key (clientSecret)
	h := hmac.New(sha256.New, []byte(clientSecret))

	// Write the data to be hashed (username + clientId)
	_, err := h.Write([]byte(username + clientId))
	if err != nil {
		return "", err
	}

	// Get the final HMAC result
	secretHash := h.Sum(nil)

	// Encode the result to base64
	secretHashBase64 := base64.StdEncoding.EncodeToString(secretHash)

	return secretHashBase64, nil
}


// cognito federeated identity pool, sign in with google, facebook, etc
func (awsCognito *AWSCognito) RegisterFederatedIdentityProvider(signUpData *requestModels.SignUp) error {
	// https://docs.aws.amazon.com/cognito/latest/developerguide/google.html

	err := error(nil)
	if awsCognito.cognitoClient == nil {
		_, err = awsCognito.init()
		if err != nil {
			return err
		}
	}

	_, err = awsCognito.cognitoClient.CreateIdentityProvider(context.TODO(), &cognitoidentityprovider.CreateIdentityProviderInput{
		ProviderName: aws.String("Google"),
		ProviderType: "Google",
		ProviderDetails: map[string]string{
			"client_id":     "GOOGLE_CLIENT_ID",
			"client_secret": "GOOGLE_CLIENT_SECRET",
			"authorize_scopes": "email openid profile",
		},
		UserPoolId: aws.String("USER_POOL_ID"),
	});
	if err != nil {
		fmt.Println("Couldn't create identity provider")
		fmt.Println(err)
	}
	return err

}

func (awsCognito *AWSCognito) getProvoderDetails(providerName string) (map[string]string, error) {
	err := error(nil)
	var details map[string]string

	details = make(map[string]string)
	//Github
	details["client_id"] = "Ov23lik7QPHeol5djOPu"
	details["client_secret"] = "fafcdc20a6a4ddad2bf28e7d2a0634fd9d83b081"
	details["authorize_scopes"] = "read:user,user:email"

	return output.ProviderDescription.ProviderDetails, nil
}


//https://dev.organic-forms.com/oauth2/authorize?identity_provider=github&redirect_uri=https://dev.organic-forms.com/api/auth/callback&response_type=CODE&client_id=5vf304hht0uhhf1jo7ql0asb5p&scope=email%20openid
//https://dev.organic-forms.com/oauth2/authorize?identity_provider=github&redirect_uri=https://dev.organic-forms.com/api/auth/callback&response_type=CODE&client_id=5vf304hht0uhhf1jo7ql0asb5p&scope=email%20openid
//https://dev.organic-forms.com/oauth2/authorize?identity_provider=linkedin&redirect_uri=https://dev.organic-forms.com/api/auth/callback&response_type=CODE&client_id=5vf304hht0uhhf1jo7ql0asb5p&scope=email%20openid

//https://dev.organic-forms.com/oauth2/authorize?identity_provider=Google&redirect_uri=https://dev.organic-forms.com/api/auth/callback&response_type=CODE&client_id=5vf304hht0uhhf1jo7ql0asb5p&scope=email%20openid