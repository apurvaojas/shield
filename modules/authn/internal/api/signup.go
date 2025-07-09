package api

import (
	"net/http"

	"shield/modules/authn/internal/api/dto"
	"shield/modules/authn/internal/auth"

	"github.com/gin-gonic/gin"
)

// Signup handles the user registration process.
// @Summary Register a new individual user
// @Description Creates a new user account with email and password.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param signupRequest body dto.SignupRequest true "Signup Request"
// @Success 201 {object} dto.SignupResponse "User registered successfully, verification pending"
// @Failure 400 {object} dto.ErrorResponse "Invalid request payload"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/signup [post]
func (h *AuthHandler) Signup(c *gin.Context) {
	var req dto.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	serviceReq := auth.SignupUserRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.authService.SignupUser(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to signup user"})
		return
	}

	response := dto.SignupResponse{
		UserID:               resp.UserID,
		RequiresConfirmation: resp.RequiresConfirmation,
	}

	if resp.CodeDeliveryDetails != nil {
		response.CodeDeliveryDetails = &dto.CodeDeliveryDetails{
			AttributeName:  resp.CodeDeliveryDetails.AttributeName,
			DeliveryMedium: resp.CodeDeliveryDetails.DeliveryMedium,
			Destination:    resp.CodeDeliveryDetails.Destination,
		}
	}

	if resp.RequiresConfirmation {
		response.Message = "Verification code sent to your email"
	} else {
		response.Message = "Account created successfully"
	}

	c.JSON(http.StatusCreated, response)
}

// ConfirmSignup handles the confirmation of user signup via verification code.
// @Summary Confirm user signup
// @Description Confirms user signup with verification code sent to email.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param confirmSignupRequest body dto.ConfirmSignupRequest true "Confirm Signup Request"
// @Success 200 {object} dto.SuccessResponse "User confirmed successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request payload"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/confirm [post]
func (h *AuthHandler) ConfirmSignup(c *gin.Context) {
	var req dto.ConfirmSignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	serviceReq := auth.ConfirmSignupRequest{
		Email:            req.Email,
		VerificationCode: req.VerificationCode,
	}

	_, err := h.authService.ConfirmUserSignup(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to confirm signup"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "Account confirmed successfully"})
}
