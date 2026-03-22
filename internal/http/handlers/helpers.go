package handlers

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
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

func getPaginationParams(c *gin.Context) (int, int, error) {
	page := defaultPage
	limit := defaultLimit

	if rawPage := c.Query("page"); rawPage != "" {
		parsedPage, err := strconv.Atoi(rawPage)
		if err != nil || parsedPage <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})

			return 0, 0, errors.New("invalid page")
		}

		page = parsedPage
	}

	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})

			return 0, 0, errors.New("invalid limit")
		}

		if parsedLimit > maxLimit {
			parsedLimit = maxLimit
		}

		limit = parsedLimit
	}

	return page, limit, nil
}
