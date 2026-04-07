package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	// defaultOffset это дефолтное значение для смещения при пагинации.
	defaultOffset = 0
	// defaultLimit это дефолтное значения для лимита при пагинации.
	defaultLimit = 20
	// maxLimit это максимальное значение для лимита при пагинации.
	maxLimit = 100
)


// getIdParam извлекает и валидирует параметр id из URL.
func getIdParam(c *gin.Context) (uint, error) {
	idStr := c.Param("id")

	fmt.Println("params:", c.Params)

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})

		return 0, errors.New("invalid id")
	}

	return uint(id), nil
}

func getBeerIDParam(c *gin.Context) (uint, error) {
	beerIDStr := c.Param("beer_id")

	beerID, err := strconv.Atoi(beerIDStr)
	if err != nil || beerID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid beer id"})

		return 0, fmt.Errorf("invalid beer id: %w", err)
	}

	return uint(beerID), nil
}

func getCategoryIDParam(c *gin.Context) (uint, error) {
	ctgIDStr := c.Param("category_id")

	beerID, err := strconv.Atoi(ctgIDStr)
	if err != nil || beerID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})

		return 0, fmt.Errorf("invalid category id: %w", err)
	}

	return uint(beerID), nil
}

// readRequestBody читает тело HTTP-запроса и возвращает его в виде байтового среза.
func readRequestBody(c *gin.Context) ([]byte, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})

		return nil, errors.New("invalid request body")
	}

	return body, nil
}

// getPaginationParams извлекает параметры пагинации из HTTP-запроса
func getPaginationParams(c *gin.Context) (uint64, uint64, error) {
	offset := defaultOffset
	limit := defaultLimit

	if rawOffset := c.Query("offset"); rawOffset != "" {
		parsedOffset, err := strconv.Atoi(rawOffset)
		if err != nil || parsedOffset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})

			return 0, 0, errors.New("invalid offset")
		}

		offset = parsedOffset
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

	return uint64(offset), uint64(limit), nil
}
