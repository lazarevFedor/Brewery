// Package handlers содержит реализацию HTTP-обработчиков для управления пивоварней.
package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// defaultOffset это дефолтное значение для смещения при пагинации.
	defaultOffset = 0

	// defaultLimit это дефолтное значения для лимита при пагинации.
	defaultLimit = 20

	// maxLimit это максимальное значение для лимита при пагинации.
	maxLimit = 100

	// filterParamArgsNum - кол-во аргументов параметра фильтрации.
	// Т. е. field:operation:number - 3 части.
	// Необходим для валидации параметра фильтрации
	filterParamArgsNum = 3
)

func writeError(c *gin.Context, code int, errType, message string) {
	c.JSON(code, gin.H{
		"error":   errType,
		"message": message,
	})
}

var numFields = []string{"rating", "abv", "ibu", "amount", "country_id", "city_id"}

// validateFilterParam проверяет правильность параметра фильтрации и возвращает его в виде структуры FilterParameter.
func validateFilterParam(filter string) (*entities.FilterParameter, error) {
	filterParams := strings.Split(filter, ":")
	if len(filterParams) != filterParamArgsNum {
		return nil, errors.New("неправильный параметр")
	}

	if !slices.Contains(numFields, filterParams[0]) {
		return nil, errors.New("неверное поле")
	}
	filterEntity := &entities.FilterParameter{
		FieldName: filterParams[0],
	}

	rawVal := filterParams[2]
	val, err := strconv.ParseFloat(rawVal, 32)
	if err != nil {
		return nil, errors.New(apperrors.CodeInvalidParameters)
	}
	filterEntity.Value = float32(val)
	var oper entities.Operation
	switch filterParams[1] {
	case "eq":
		oper = entities.OpEqual
	case "gt":
		oper = entities.OpGreater
	case "ge":
		oper = entities.OpGreaterEqual
	case "le":
		oper = entities.OpLess
	case "lt":
		oper = entities.OpLessEqual
	case "ne":
		oper = entities.OpNotEqual
	default:
		return nil, errors.New("неверная операция")
	}
	filterEntity.Operation = oper

	return filterEntity, nil
}

// getUintParam извлекает и валидирует целочисленный ненулевой параметр из URL, например id.
func getUintParam(c *gin.Context, name string) (uint, error) {
	param := c.Param(name)
	if param == "" {
		return 0, nil
	}

	uintParam, err := strconv.Atoi(param)
	if err != nil {
		return 0, err
	}
	if uintParam <= 0 {
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
