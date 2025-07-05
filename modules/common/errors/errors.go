package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrUnauthorized     = &AppError{"UNAUTHORIZED", "Unauthorized access", http.StatusUnauthorized}
	ErrForbidden        = &AppError{"FORBIDDEN", "Access forbidden", http.StatusForbidden}
	ErrInvalidToken     = &AppError{"INVALID_TOKEN", "Invalid or expired token", http.StatusUnauthorized}
	ErrInvalidNonce     = &AppError{"INVALID_NONCE", "Invalid nonce", http.StatusBadRequest}
	ErrUserNotFound     = &AppError{"USER_NOT_FOUND", "User not found", http.StatusNotFound}
	ErrInternalServer   = &AppError{"INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError}
	ErrRateLimitExceeded = &AppError{"RATE_LIMIT_EXCEEDED", "Rate limit exceeded", http.StatusTooManyRequests}
)

func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

type ErrorHandler struct{}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

func (h *ErrorHandler) HandleError(c *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		c.JSON(appErr.Status, gin.H{
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			},
		})
		return
	}

	// Log unexpected errors
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "An unexpected error occurred",
		},
	})
}

func (h *ErrorHandler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			h.HandleError(c, err)
		}
	}
}