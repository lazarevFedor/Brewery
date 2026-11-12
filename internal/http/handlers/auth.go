package handlers

import (
	"Brewery/internal/apperrors"
	"Brewery/pkg/logger"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

const (
	day = 24 * time.Hour

	// CodeUnauthorized - код ошибки для неверных учетных данных и отсутствующего/невалидного токена.
	CodeUnauthorized = "UNAUTHORIZED"
)

// AuthHandlers определяет интерфейс для обработки HTTP-запросов авторизации администратора.
type AuthHandlers interface {
	Login(c *gin.Context)
}

// authHandlers реализует интерфейс AuthHandlers.
type authHandlers struct {
}

// NewAuthHandlers создает новый экземпляр AuthHandlers.
func NewAuthHandlers() AuthHandlers {
	return &authHandlers{}
}

// AdminClaims - данные JWT-токена администратора.
type AdminClaims struct {
	jwt.RegisteredClaims

	Username string `json:"username"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login обрабатывает POST /login и выдает JWT-токен администратору.
func (h *authHandlers) Login(c *gin.Context) {
	log, ok := logger.GetLoggerFromCtx(c.Request.Context())
	if !ok {
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		return
	}

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, apperrors.CodeBadRequest, "username and password required")
		log.Warn(c.Request.Context(), "invalid JSON", zap.Error(err))
		return
	}

	adminUser, adminPass := adminLogin()
	if req.Username != adminUser || req.Password != adminPass {
		writeError(c, http.StatusForbidden, CodeUnauthorized, "Unauthorized")
		log.Warn(c.Request.Context(), "invalid credentials")
		return
	}

	token, err := GenerateToken(req.Username)
	if err != nil {
		writeError(c, http.StatusInternalServerError, apperrors.CodeInternalError, "Unexpected error occurred")
		log.Error(c.Request.Context(), "could not generate token", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GenerateToken выпускает JWT-токен администратора сроком на сутки.
func GenerateToken(username string) (string, error) {
	claims := AdminClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(day)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret())
}

// ParseToken проверяет JWT-токен и возвращает его данные.
func ParseToken(tokenStr string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AdminClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret(), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(*AdminClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
