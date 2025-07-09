package api

import (
	"net/http"

	"shield/modules/authn/internal/api/dto"
	"shield/modules/authn/internal/auth"

	"github.com/gin-gonic/gin"
)

// OrgSignup handles organization signup with admin user creation.
// @Summary Register a new organization
// @Description Creates a new organization with an admin user account.
// @Tags Organization
// @Accept json
// @Produce json
// @Param orgSignupRequest body dto.OrgSignupRequest true "Organization Signup Request"
// @Success 201 {object} dto.OrgSignupResponse "Organization registered successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request payload"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/org/signup [post]
func (h *AuthHandler) OrgSignup(c *gin.Context) {
	var req dto.OrgSignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	serviceReq := auth.OrgSignupRequest{
		OrgName:       req.OrgName,
		AdminEmail:    req.AdminEmail,
		AdminPassword: req.AdminPassword,
	}

	resp, err := h.authService.OrgSignup(c.Request.Context(), serviceReq)
	if err != nil {
		// TODO: Map service layer errors to HTTP errors
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to create organization"})
		return
	}

	c.JSON(http.StatusCreated, dto.OrgSignupResponse{
		OrgID:       resp.OrgID,
		AdminUserID: resp.AdminUserID,
		Message:     "Organization created successfully. Admin user verification required.",
	})
}

// GetOrgDetails handles getting organization details.
// @Summary Get organization details
// @Description Retrieves details of an organization.
// @Tags Organization
// @Produce json
// @Param orgId path string true "Organization ID"
// @Success 200 {object} dto.OrgDetails "Organization details"
// @Failure 404 {object} dto.ErrorResponse "Organization not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/org/{orgId} [get]
func (h *AuthHandler) GetOrgDetails(c *gin.Context) {
	orgID := c.Param("orgId")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Organization ID is required"})
		return
	}

	// TODO: Implement GetOrgDetails in auth service
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{Error: "Not implemented yet"})
}

// UpdateOrg handles updating organization settings.
// @Summary Update organization
// @Description Updates organization settings such as SSO configuration.
// @Tags Organization
// @Accept json
// @Produce json
// @Param orgId path string true "Organization ID"
// @Param updateOrgRequest body dto.UpdateOrgRequest true "Update Organization Request"
// @Success 200 {object} dto.SuccessResponse "Organization updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request payload"
// @Failure 404 {object} dto.ErrorResponse "Organization not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /auth/org/{orgId} [put]
func (h *AuthHandler) UpdateOrg(c *gin.Context) {
	orgID := c.Param("orgId")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Organization ID is required"})
		return
	}

	var req dto.UpdateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// TODO: Implement UpdateOrg in auth service
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{Error: "Not implemented yet"})
}
