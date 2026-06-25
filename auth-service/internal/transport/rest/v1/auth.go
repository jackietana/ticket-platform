package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackietana/ticket-platform/auth-service/internal/domain"
)

const CTX_TIMEOUT = time.Second * 5

func (h *Handler) SignUp(c *gin.Context) {
	var usrReq domain.UserRequest
	if err := c.ShouldBindJSON(&usrReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input body"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), CTX_TIMEOUT)
	defer cancel()

	id, err := h.auth.SignUp(ctx, domain.User{
		Email:    usrReq.Email,
		Password: usrReq.Password,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "successfully signed up"})
}

func (h *Handler) SignIn(c *gin.Context) {
	var usrReq domain.UserRequest
	if err := c.ShouldBindJSON(&usrReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input body"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), CTX_TIMEOUT)
	defer cancel()

	token, err := h.auth.SignIn(ctx, domain.User{
		Email:    usrReq.Email,
		Password: usrReq.Password,
	}, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
