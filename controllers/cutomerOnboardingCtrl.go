package controllers

import (
	"net/http"
	"org-forms-config-management/infra/logger"
	"org-forms-config-management/models/requestModels"
	services "org-forms-config-management/services"

	"github.com/gin-gonic/gin"
)

type CustomerOnboardingCtrl struct{}

// SignUp godoc
//
//	@Summary		Sign up for a new account
//	@Description	Supports Both Individual and Organization Account
//	@Tags			Accounts
//	@Accept			json
//	@Produce		json
//	@Param			account	body		requestModels.SignUp	true	"Add account"
//	@Success		200		{string}	string					"success"
//	@Router			/api/v1/onboarding/signup [post]
func (ctrl *CustomerOnboardingCtrl) SignUp(ctx *gin.Context) {

	var signUp requestModels.SignUp
	if err := ctx.ShouldBindJSON(&signUp); err != nil {
		logger.Errorf("Error while binding request body: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body\n" + err.Error()})
		return
	}

	signUpService := &services.SignUpService{}
	userId, err := signUpService.SignUp(&signUp)
	if err != nil {
		logger.Errorf("Error while signing up: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error while signing up\n" + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success", "userId": userId})
}

// VerifyEmail godoc
//
//	@Summary		Verify email
//	@Description	Verify email
//	@Tags			Accounts
//	@Accept			json
//	@Produce		json
//	@Param			account	body		requestModels.VerifyEmail	true	"verify email"
//	@Success		200		{string}	string						"success"
//	@Router			/api/v1/onboarding/verifyEmail [post]
func (ctrl *CustomerOnboardingCtrl) VerifyEmail(ctx *gin.Context) {
	var verifyEmail requestModels.VerifyEmail
	if err := ctx.ShouldBindJSON(&verifyEmail); err != nil {
		logger.Errorf("Error while binding request body: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body\n" + err.Error()})
		return
	}

	signUpService := &services.SignUpService{}
	err := signUpService.VerifyEmail(verifyEmail.UserEmail, verifyEmail.ConfirmationCode)
	if err != nil {
		logger.Errorf("Error while verifying email: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error while verifying email\n" + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

// ResendConfirmationCode godoc
//
//	@Summary		Resend confirmation code
//	@Description	Resend confirmation code
//	@Tags			Accounts
//	@Accept			json
//	@Produce		json
//	@Param			account	body		requestModels.ResendVerificationCode	true	"resend confirmation code"
//	@Success		200		{string}	string										"success"
//	@Router			/api/v1/onboarding/resendConfirmationCode [post]
func (ctrl *CustomerOnboardingCtrl) ResendConfirmationCode(ctx *gin.Context) {
	var resendConfirmationCode requestModels.ResendVerificationCode
	if err := ctx.ShouldBindJSON(&resendConfirmationCode); err != nil {
		logger.Errorf("Error while binding request body: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body\n" + err.Error()})
		return
	}

	signUpService := &services.SignUpService{}
	err := signUpService.ResendVerificationCode(resendConfirmationCode.UserEmail)
	if err != nil {
		logger.Errorf("Error while resending confirmation code: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error while resending confirmation code\n" + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}
