package rest

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackietana/ticket-platform/auth-service/internal/domain"
)

type AuthService interface {
	SignUp(ctx context.Context, usr domain.User) (string, error)
	SignIn(ctx context.Context, usr domain.User, clientIp, userAgent string) (string, error)
	ValidateSession(ctx context.Context, token, clientIP, userAgent string) (string, error)
}

type Handler struct {
	auth AuthService
}

func NewHandler(auth AuthService) *Handler {
	return &Handler{auth}
}

func (h *Handler) Init() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/sign-up", h.SignUp)
			auth.POST("/sign-in", h.SignIn)
		}
	}

	return router
}
