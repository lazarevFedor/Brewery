package handlers

// import (
// 	appErrors "Brewery/internal/errors"
// 	"errors"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// func handleError(c *gin.Context, err error) {
// 	var appErr *appErrors.AppError

// 	if errors.As(err, &appErr) {
// 		c.JSON(appErr.HTTPStatus, gin.H{
// 			"error":   appErr.Code,
// 			"message": appErr.Message,
// 		})
// 		return
// 	}

// 	// fallback
// 	c.JSON(http.StatusInternalServerError, gin.H{
// 		"error":   appErrors.CodeInternalError,
// 		"message": "Unexpected error occurred",
// 	})
// }
