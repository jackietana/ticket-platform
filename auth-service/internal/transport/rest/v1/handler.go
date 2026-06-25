package rest

import (
	"context"

	"github.com/gin-gonic/gin"
	_ "github.com/jackietana/ticket-platform/auth-service/docs"
	"github.com/jackietana/ticket-platform/auth-service/internal/dto"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//go:generate mockgen -source=handler.go -destination=mocks/mock.go

type AuthService interface {
	SignUp(ctx context.Context, usr dto.UserRequest) (string, error)
	SignIn(ctx context.Context, usr dto.UserRequest, clientIp, userAgent string) (string, error)
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
			auth.POST("/sign-up", h.signUp)
			auth.POST("/sign-in", h.signIn)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
