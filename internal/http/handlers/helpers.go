package handlers

import (
	"errors"
	"io"
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

func writeError(c *gin.Context, code int, errType, message string) {
	c.JSON(code, gin.H{
		"error":   errType,
		"message": message,
	})
}

// getUintParam извлекает и валидирует целочисленный ненулевой параметр из URL, например id.
func getUintParam(c *gin.Context, name string) (uint, error) {
	param := c.Param(name)

	uintParam, err := strconv.Atoi(param)
	if err != nil || uintParam <= 0 {
		return 0, errors.New("invalid id")
	}

	return uint(uintParam), nil
}

// readRequestBody читает тело HTTP-запроса и возвращает его в виде байтового среза.
func readRequestBody(c *gin.Context) ([]byte, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
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
			return 0, 0, errors.New("invalid offset")
		}

		offset = parsedOffset
	}

	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit <= 0 {
			return 0, 0, errors.New("invalid limit")
		}

		if parsedLimit > maxLimit {
			parsedLimit = maxLimit
		}

		limit = parsedLimit
	}

	return uint64(offset), uint64(limit), nil
}
