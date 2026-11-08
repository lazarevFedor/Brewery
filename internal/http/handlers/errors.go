package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/pkg/logger"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// codeNotImplemented - код ошибки для запросов, которые описаны в swagger, но еще не реализованы.
const codeNotImplemented = "NOT_IMPLEMENTED"

// handleError логирует ошибку сервиса и отправляет ответ. Ошибки apperrors.AppError
// отдаются со своим статусом и кодом, остальные - как 500.
func handleError(c *gin.Context, log *logger.Logger, msg string, err error) {
	if appErr, ok := errors.AsType[*apperrors.AppError](err); ok {
		if appErr.HTTPStatus >= http.StatusInternalServerError {
			log.Error(c.Request.Context(), msg, zap.Error(err))
		} else {
			log.Warn(c.Request.Context(), msg, zap.Error(err))
		}

		writeError(c, appErr.HTTPStatus, appErr.APIErr.ErrorCode, appErr.APIErr.Message)
		return
	}

	log.Error(c.Request.Context(), msg, zap.Error(err))
	writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
}

// notImplemented отвечает 501 для запросов, под которые еще нет реализации сервиса.
func notImplemented(c *gin.Context) {
	writeError(c, http.StatusNotImplemented, codeNotImplemented, "Not implemented yet")
}

// badRequest логирует некорректный запрос и отвечает 400.
func badRequest(c *gin.Context, log *logger.Logger, code string, err error) {
	log.Warn(c.Request.Context(), "bad request", zap.Error(err))
	writeError(c, http.StatusBadRequest, code, err.Error())
}
