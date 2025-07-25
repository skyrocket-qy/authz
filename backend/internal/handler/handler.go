package handler

import (
	"srv/internal/logic"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	logic logic.Logic
}

func NewHandler(logic logic.Logic) *Handler {
	return &Handler{
		logic: logic,
	}
}

// @Tags		auth
// @Param		user	body		logic.LoginIn	true	"request body"
// @Success	200		{object}	logic.LoginOut
// @Router		/v1/login [post].
func (h *Handler) Login(c *gin.Context, req *logic.LoginIn) (*logic.LoginOut, error) {
	return h.logic.Login(c.Request.Context(), req)
}

// @Tags		auth
// @Param		user	body	authcontrollerx.ChangePassword.Req	true	"req"
// @Success	200
// @Router		/v1/change-password [post].
func (h *Handler) ChangePassword(c *gin.Context, req *logic.ChangePasswordIn) error {
	return h.logic.ChangePassword(c.Request.Context(), req)
}

// @Param		user	body	logic.ConfirmForgotPasswordIn	true	"Request body"
// @Success	200
// @Router		/v1/confirm-forgot-password [post]
// @Tags		auth.
func (h *Handler) ConfirmForgotPassword(c *gin.Context, req *logic.ConfirmForgotPasswordIn) error {
	return h.logic.ConfirmForgotPassword(c.Request.Context(), req)
}

// @Tags		auth
// @Param		user	body	logic.ConfirmSignUpIn	true	"Request body"
// @Success	200
// @Router		/v1/confirm-sign-up [post].
func (h *Handler) ConfirmSignUp(c *gin.Context, req *logic.ConfirmSignUpIn) error {
	return h.logic.ConfirmSignUp(c.Request.Context(), req)
}

// @Tags		auth
// @Param		user	body	logic.ForgotPasswordIn	true	"Request body"
// @Success	200
// @Router		/v1/forgot-password [post].
func (h *Handler) ForgotPassword(c *gin.Context, req *logic.ForgotPasswordIn) error {
	return h.logic.ForgotPassword(c.Request.Context(), req)
}

// @Tags		auth
// @Param		user	body		logic.LoginIn	true	"request body"
// @Success	200		{object}	logic.LoginOut
// @Router		/v1/invite-user [post].
func (h *Handler) InviteUser(c *gin.Context, req *logic.InviteUserIn) error {
	return h.logic.InviteUser(c.Request.Context(), req)
}

// @Tags		auth
// @Param		user	body		logic.RefreshTokenIn	true	"req"
// @Success	200		{object}	logic.RefreshTokenOut
// @Router		/v1/refresh-token [post].
func (h *Handler) RefreshToken(c *gin.Context, req *logic.RefreshTokenIn) (*logic.RefreshTokenOut, error) {
	return h.logic.RefreshToken(c.Request.Context(), req)
}

// @Tags		auth
// @Param		user	body	logic.ResendConfirmationCodeIn	true	"request body"
// @Success	200
// @Router		/v1/resend-confirmation-code [post].
func (h *Handler) ResendConfirmationCode(c *gin.Context, req *logic.ResendConfirmationCodeIn) error {
	return h.logic.ResendConfirmationCode(c.Request.Context(), req)
}

// @Tags		auth
// @Param		user	body	logic.SetNewPasswordIn	true	"request body"
// @Success	200
// @Router		/v1/set-new-password [post].
func (h *Handler) SetNewPassword(c *gin.Context, req *logic.SetNewPasswordIn) error {
	return h.logic.SetNewPassword(c.Request.Context(), req)
}

// @Tags		auth
// @Param		user	body	logic.SignUpIn	true	"req"
// @Success	200
// @Router		/v1/sign-up [post].
func (h *Handler) SignUp(c *gin.Context, req *logic.SignUpIn) error {
	return h.logic.SignUp(c.Request.Context(), req)
}
