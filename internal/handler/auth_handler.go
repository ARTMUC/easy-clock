package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	userapplication "starter/internal/application/user"
	domainuser "starter/internal/domain/user"
	"starter/internal/middleware"
	"starter/internal/views/pages"
)

type AuthHandler struct {
	svc *userapplication.Service
}

func NewAuthHandler(svc *userapplication.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) ShowLogin(c *gin.Context) {
	renderTempl(c, pages.LoginPage(""))
}

func (h *AuthHandler) HandleLogin(c *gin.Context) {
	var req userapplication.LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		renderTempl(c, pages.LoginPage("Please fill in all fields."))
		return
	}

	dto, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotActive):
			renderTempl(c, pages.LoginPage("Please confirm your email address before logging in."))
		case errors.Is(err, domainuser.ErrInvalidCredentials):
			renderTempl(c, pages.LoginPage("Invalid email or password."))
		default:
			slog.Error("login failed", "error", err)
			renderTempl(c, pages.LoginPage("Server error. Please try again."))
		}
		return
	}

	session := sessions.Default(c)
	session.Set(middleware.SessionUserKey, dto.ID)
	_ = session.Save()

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func (h *AuthHandler) ShowRegister(c *gin.Context) {
	renderTempl(c, pages.RegisterPage(""))
}

func (h *AuthHandler) HandleRegister(c *gin.Context) {
	var req userapplication.RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		renderTempl(c, pages.RegisterPage("Please fill in all fields (password min. 8 characters)."))
		return
	}

	dto, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrEmailTaken):
			renderTempl(c, pages.RegisterPage("This email is already taken."))
		default:
			slog.Error("register failed", "error", err)
			renderTempl(c, pages.RegisterPage("Failed to send verification email. Check your address or try again."))
		}
		return
	}

	renderTempl(c, pages.CheckEmailPage(dto.Email))
}

func (h *AuthHandler) HandleVerify(c *gin.Context) {
	token := c.Query("token")
	err := h.svc.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		slog.Error("verify email failed", "error", err)
		renderTempl(c, pages.VerifyPage(false, "The verification link is invalid or has expired."))
		return
	}
	renderTempl(c, pages.VerifyPage(true, "Your account has been activated. You can now log in."))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	_ = session.Save()
	c.Redirect(http.StatusSeeOther, "/login")
}
