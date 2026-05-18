package handlers

import (
	"Brewery/internal/apperrors"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleError(c *gin.Context, err error) {
	var appErr *apperrors.AppError

	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, gin.H{
			"error":   appErr.APIErr.ErrorCode,
			"message": appErr.APIErr.Message,
		})
		return
	}

	// fallback
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   apperrors.CodeInternalError,
		"message": "Unexpected error occurred",
	})
}
