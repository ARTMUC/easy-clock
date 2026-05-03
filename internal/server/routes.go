package router

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	userapplication "easy-clock/internal/application/user"
	"easy-clock/internal/eventbus"
	"easy-clock/internal/handler"
	"easy-clock/internal/middleware"
	emailinfra "easy-clock/internal/infrastructure/email"
	userpersistence "easy-clock/internal/infrastructure/persistence/user"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.GET("/health", s.healthHandler)

	gormDB := s.db.GetDB()
	cfg := s.cfg

	r.Static("/static", "./static")
	r.StaticFile("/manifest.json", "./static/manifest.json")
	r.StaticFile("/sw.js", "./static/sw.js")

	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("app_session", store))

	bus := eventbus.New()
	bus.RegisterDefaultHandlers()

	userRepo := userpersistence.NewUserRepository(gormDB)
	emailClient := emailinfra.NewBrevoClient(cfg.BrevoAPIKey, cfg.BrevoSenderEmail, cfg.BrevoSenderName)
	userSvc := userapplication.NewService(userRepo, bus, emailClient, cfg.AppBaseURL)

	authH := handler.NewAuthHandler(userSvc)
	r.GET("/login", authH.ShowLogin)
	r.POST("/login", authH.HandleLogin)
	r.GET("/register", authH.ShowRegister)
	r.POST("/register", authH.HandleRegister)
	r.GET("/verify", authH.HandleVerify)
	r.POST("/logout", authH.Logout)

	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())

	protected.GET("/", func(c *gin.Context) { c.Redirect(http.StatusSeeOther, "/dashboard") })

	dashH := handler.NewDashboardHandler()
	protected.GET("/dashboard", dashH.ShowDashboard)

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
