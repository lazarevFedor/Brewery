package handlers

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getIdParam(c *gin.Context) (int, error) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})

		return 0, errors.New("invalid id")
	}

	return id, nil
}

func readRequestBody(c *gin.Context) ([]byte, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})

		return nil, errors.New("invalid request body")
	}

	return body, nil
}
