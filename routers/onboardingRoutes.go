package routers

import (
	"org-forms-config-management/controllers"

	"github.com/gin-gonic/gin"
)

func CustomerOnboardingRoutes(router *gin.Engine) {
	ctrl := controllers.CustomerOnboardingCtrl{}
	onboarding := router.Group("/api/v1/onboarding")
	{
		onboarding.POST("/signup", ctrl.SignUp)
		onboarding.POST("/verifyEmail", ctrl.VerifyEmail)
		onboarding.POST("/resendConfirmationCode", ctrl.ResendConfirmationCode)
		// onboarding.POST("/enableMFA", ctrl.EnableMFA)
	}

	login := router.Group("/api/v1/login")
	{
		login.POST("/", ctrl.Login)
		login.POST("/federatedIDP", ctrl.FederatedLogin)
		login.POST("/forgotPassword", ctrl.forgotPassword)
		login.GET("/federatedIDP/callback", ctrl.FederatedLoginCallback)
		login.POST("/federatedIDP/confirmLinkage", ctrl.FederatedIDPConfirmLinkage)
	}
}
