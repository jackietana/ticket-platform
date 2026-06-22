package rest

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackietana/ticket-platform/auth-service/internal/domain"
)

type AuthService interface {
	SignUp(ctx context.Context, usr domain.User) error
	SignIn(ctx context.Context, inp domain.UserInput) (string, error)
	GetUserIdByToken(ctx context.Context, inp string) (string, error)
}

type Handler struct {
	auth AuthService
}

func NewHandler(auth AuthService) *Handler {
	return &Handler{auth}
}

func (h *Handler) Init() *gin.Engine {
	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/sign-up", h.SignUp)
			auth.GET("/sign-in", h.SignIn)
		}
	}

	return router
}
