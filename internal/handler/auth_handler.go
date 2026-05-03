package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	userapplication "easy-clock/internal/application/user"
	domainuser "easy-clock/internal/domain/user"
	"easy-clock/internal/i18n"
	"easy-clock/internal/middleware"
	"easy-clock/internal/views/pages"
)

type AuthHandler struct {
	svc *userapplication.Service
}

func NewAuthHandler(svc *userapplication.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) ShowLogin(c *gin.Context) {
	renderTempl(c, pages.LoginPage("", lang(c)))
}

func (h *AuthHandler) HandleLogin(c *gin.Context) {
	l := lang(c)
	var req userapplication.LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		renderTempl(c, pages.LoginPage(i18n.Msg(i18n.MsgFillAllFields, l), l))
		return
	}
	dto, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		var msg string
		switch {
		case errors.Is(err, domainuser.ErrNotActive):
			msg = i18n.Msg(i18n.MsgNotActive, l)
		case errors.Is(err, domainuser.ErrInvalidCredentials):
			msg = i18n.Msg(i18n.MsgInvalidCredentials, l)
		default:
			slog.Error("login failed", "error", err)
			msg = i18n.Msg(i18n.MsgServerError, l)
		}
		renderTempl(c, pages.LoginPage(msg, l))
		return
	}
	session := sessions.Default(c)
	session.Set(middleware.SessionUserKey, dto.ID)
	_ = session.Save()
	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func (h *AuthHandler) ShowRegister(c *gin.Context) {
	renderTempl(c, pages.RegisterPage("", lang(c)))
}

func (h *AuthHandler) HandleRegister(c *gin.Context) {
	l := lang(c)
	var req userapplication.RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		renderTempl(c, pages.RegisterPage(i18n.Msg(i18n.MsgFillAllFields, l), l))
		return
	}
	dto, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		var msg string
		switch {
		case errors.Is(err, domainuser.ErrEmailTaken):
			msg = i18n.Msg(i18n.MsgEmailTaken, l)
		default:
			slog.Error("register failed", "error", err)
			msg = i18n.Msg(i18n.MsgVerificationEmailFailed, l)
		}
		renderTempl(c, pages.RegisterPage(msg, l))
		return
	}
	renderTempl(c, pages.CheckEmailPage(dto.Email, l))
}

func (h *AuthHandler) HandleVerify(c *gin.Context) {
	l := lang(c)
	token := c.Query("token")
	if err := h.svc.VerifyEmail(c.Request.Context(), token); err != nil {
		slog.Error("verify email failed", "error", err)
		renderTempl(c, pages.VerifyPage(false, i18n.Msg(i18n.MsgInvalidToken, l), l))
		return
	}
	renderTempl(c, pages.VerifyPage(true, i18n.Msg(i18n.MsgAccountActivated, l), l))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	_ = session.Save()
	c.Redirect(http.StatusSeeOther, "/login")
}

// -----------------------------------------------------------------
// JSON API auth endpoints (JWT flow)
// -----------------------------------------------------------------

func (h *AuthHandler) APIRegister(c *gin.Context) {
	var req userapplication.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindErr(c, err)
		return
	}
	dto, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		slog.Error("api register", "error", err)
		if errors.Is(err, domainuser.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		return
	}
	c.JSON(http.StatusCreated, dto)
}

func (h *AuthHandler) APILogin(c *gin.Context) {
	var req userapplication.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindErr(c, err)
		return
	}
	pair, err := h.svc.LoginWithTokens(c.Request.Context(), req)
	if err != nil {
		slog.Error("api login", "error", err)
		if errors.Is(err, domainuser.ErrInvalidCredentials) || errors.Is(err, domainuser.ErrNotActive) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}
	c.JSON(http.StatusOK, pair)
}

func (h *AuthHandler) APIRefresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		bindErr(c, err)
		return
	}
	pair, err := h.svc.Refresh(c.Request.Context(), body.RefreshToken)
	if err != nil {
		slog.Error("api refresh", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}
	c.JSON(http.StatusOK, pair)
}

func (h *AuthHandler) APILogout(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		bindErr(c, err)
		return
	}
	if err := h.svc.RevokeToken(c.Request.Context(), body.RefreshToken); err != nil {
		slog.Error("api logout", "error", err)
	}
	c.Status(http.StatusNoContent)
}
