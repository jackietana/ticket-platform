package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackietana/ticket-platform/auth-service/internal/dto"
)

const CTX_TIMEOUT = time.Second * 5

// @Summary Sign Up
// @Description Method for registration
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.UserRequest true "User credentials"
// @Success 201 {object} dto.SignUpResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/sign-up [post]
func (h *Handler) signUp(c *gin.Context) {
	var userReq dto.UserRequest
	if err := c.ShouldBindJSON(&userReq); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid input body"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), CTX_TIMEOUT)
	defer cancel()

	id, err := h.auth.SignUp(ctx, userReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.SignUpResponse{ID: id, Message: "successfully signed up"})
}

// @Summary Sign In
// @Description Method for authentication
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.UserRequest true "User credentials"
// @Success 200 {object} dto.SignInResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/auth/sign-in [post]
func (h *Handler) signIn(c *gin.Context) {
	var userReq dto.UserRequest
	if err := c.ShouldBindJSON(&userReq); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid input body"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), CTX_TIMEOUT)
	defer cancel()

	token, err := h.auth.SignIn(ctx, userReq, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SignInResponse{Token: token})
}
