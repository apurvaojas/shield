package api

import (
	"net/http"

	"shield/modules/authn/internal/api/dto"
	"shield/modules/authn/internal/auth"
	"shield/modules/authn/internal/models"

	"github.com/gin-gonic/gin"
)

// SetupMFA handles MFA setup initiation.
// @Summary Setup MFA for a user
// @Description Initiates the MFA setup process (e.g., TOTP QR code, SMS setup).
// @Tags Authentication
// @Accept json
// @Produce json
// @Param mfaSetupRequest body dto.MFASetupRequest true "MFA Setup Request"
// @Success 200 {object} dto.MFASetupResponse "MFA setup initiated"
// @Failure 400 {object} dto.ErrorResponse "Invalid request payload"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/mfa/setup [post]
func (h *AuthHandler) SetupMFA(c *gin.Context) {
	var req dto.MFASetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	var mfaMethod models.MFAMethod
	switch req.Method {
	case string(models.MFAMethodTOTP):
		mfaMethod = models.MFAMethodTOTP
	case string(models.MFAMethodSMS):
		mfaMethod = models.MFAMethodSMS
	default:
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid MFA method"})
		return
	}

	serviceReq := auth.SetupMFARequest{
		UserID: req.UserID,
		Method: mfaMethod,
	}

	resp, err := h.authService.SetupMFA(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to setup MFA"})
		return
	}

	c.JSON(http.StatusOK, dto.MFASetupResponse{
		Secret:    resp.Secret,
		QRCodeURI: resp.QRCodeURI,
	})
}

// VerifyMFA handles MFA code verification.
// @Summary Verify MFA code
// @Description Verifies an MFA code (e.g., TOTP, SMS code) and completes login.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param mfaVerifyRequest body dto.MFAVerifyRequest true "MFA Verify Request"
// @Success 200 {object} dto.SuccessResponse "MFA verified, login complete"
// @Failure 400 {object} dto.ErrorResponse "Invalid request payload or MFA code"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/mfa/verify [post]
func (h *AuthHandler) VerifyMFA(c *gin.Context) {
	var req dto.MFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	serviceReq := auth.VerifyMFARequest{
		UserID:  req.UserID,
		MFACode: req.Code,
	}

	resp, err := h.authService.VerifyMFA(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid MFA code"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: resp.Status})
}
