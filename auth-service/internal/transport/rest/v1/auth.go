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
	var usr domain.User
	if err := c.BindJSON(&usr); err != nil {
		c.String(http.StatusBadRequest, "invalid input body")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), CTX_TIMEOUT)
	defer cancel()

	if err := h.auth.SignUp(ctx, usr); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusCreated, "seccussfully signed up")
}

func (h *Handler) SignIn(c *gin.Context) {
	var inp domain.UserInput
	if err := c.BindJSON(&inp); err != nil {
		c.String(http.StatusBadRequest, "invalid input body")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), CTX_TIMEOUT)
	defer cancel()

	token, err := h.auth.SignIn(ctx, inp)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
