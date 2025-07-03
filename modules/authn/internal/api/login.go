package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	 "github.com/tentackles/shield/modules/authn/internal/api/dto"
	"github.com/tentackles/shield/modules/authn/internal/auth"
	"github.com/tentackles/shield/modules/authn/internal/auth/session"
)

// Login handles user authentication.
// @Summary Authenticate user
// @Description Authenticates a user with email and password.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param loginRequest body dto.LoginRequest true "Login Request"
// @Success 200 {object} dto.LoginResponse "User authenticated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request payload"
// @Failure 401 {object} dto.ErrorResponse "Invalid credentials"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	serviceReq := auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	// Extract client info from request
	clientInfo := session.ClientInfo{
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		DeviceID:  c.GetHeader("X-Device-ID"), // Optional device identifier
	}

	resp, err := h.authService.Login(c.Request.Context(), serviceReq, clientInfo)
	if err != nil {
		// TODO: Map service layer errors to HTTP errors
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    int64(resp.ExpiresIn),
		TokenType:    "Bearer",
	})
}

// RefreshToken handles token refresh.
// @Summary Refresh access token
// @Description Refreshes an access token using a refresh token.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param refreshTokenRequest body dto.RefreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} dto.RefreshTokenResponse "Token refreshed successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request payload"
// @Failure 401 {object} dto.ErrorResponse "Invalid refresh token"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	serviceReq := auth.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), serviceReq)
	if err != nil {
		// TODO: Map service layer errors to HTTP errors
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, dto.RefreshTokenResponse{
		AccessToken: resp.AccessToken,
		ExpiresIn:   int64(resp.ExpiresIn),
		TokenType:   "Bearer", // Standard token type
	})
}

// Logout handles user logout.
// @Summary Logout user
// @Description Logs out a user and invalidates their session.
// @Tags Authentication
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse "User logged out successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO: Extract user/session info from context (set by middleware)
	// TODO: Implement logout in auth service

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "User logged out successfully"})
}
