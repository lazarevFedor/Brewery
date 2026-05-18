package handlers

import (
	"Brewery/internal/apperrors"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleError(c *gin.Context, err error) {
	if appErr, ok := errors.AsType[*apperrors.AppError](err); ok {
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
