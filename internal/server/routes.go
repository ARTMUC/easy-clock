package router

import (
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"easy-clock/internal/app"
	userapplication "easy-clock/internal/application/user"
	"easy-clock/internal/eventbus"
	"easy-clock/internal/handler"
	emailinfra "easy-clock/internal/infrastructure/email"
	kidclock "easy-clock/internal/infrastructure/persistence/kidclock"
	userpersistence "easy-clock/internal/infrastructure/persistence/user"
	"easy-clock/internal/middleware"
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
	r.Use(middleware.DetectLanguage())

	// --- persistence ---
	userRepo := userpersistence.NewUserRepository(gormDB)
	refreshTokenRepo := userpersistence.NewRefreshTokenRepository(gormDB)
	childRepo := kidclock.NewChildRepository(gormDB)
	profileRepo := kidclock.NewProfileRepository(gormDB)
	activityRepo := kidclock.NewActivityRepository(gormDB)
	presetRepo := kidclock.NewPresetActivityRepository(gormDB)
	scheduleRepo := kidclock.NewScheduleRepository(gormDB)
	eventRepo := kidclock.NewEventRepository(gormDB)

	// --- services ---
	bus := eventbus.New()
	bus.RegisterDefaultHandlers()
	emailClient := emailinfra.NewBrevoClient(cfg.BrevoAPIKey, cfg.BrevoSenderEmail, cfg.BrevoSenderName)
	userSvc := userapplication.NewService(userRepo, refreshTokenRepo, bus, emailClient, cfg.AppBaseURL, []byte(cfg.JWTSecret))

	childSvc := app.NewChildService(childRepo, profileRepo)
	profileSvc := app.NewProfileService(profileRepo, activityRepo, presetRepo, childRepo)
	scheduleSvc := app.NewScheduleService(scheduleRepo, profileRepo, childRepo)
	eventSvc := app.NewEventService(eventRepo, profileRepo, childRepo)
	clockSvc := app.NewClockService(childRepo, eventRepo, scheduleRepo, profileRepo)

	// --- handlers ---
	authH := handler.NewAuthHandler(userSvc)
	clockH := handler.NewClockHandler(clockSvc)
	childH := handler.NewChildHandler(childSvc)
	profileH := handler.NewProfileHandler(profileSvc)
	scheduleH := handler.NewScheduleHandler(scheduleSvc)
	eventH := handler.NewEventHandler(eventSvc)
	dashH := handler.NewDashboardHandler(childSvc)
	childCfgH := handler.NewChildConfigHandler(childSvc, profileSvc, scheduleSvc, eventSvc)
	profileCfgH := handler.NewProfileConfigHandler(profileSvc)
	presetH := handler.NewPresetHandler(profileSvc)

	// --- public routes ---
	r.GET("/login", authH.ShowLogin)
	r.POST("/login", authH.HandleLogin)
	r.GET("/register", authH.ShowRegister)
	r.POST("/register", authH.HandleRegister)
	r.GET("/verify", authH.HandleVerify)
	r.POST("/logout", authH.Logout)

	r.GET("/clock/:token", clockH.Show)

	// public API
	r.GET("/api/clock/:token", clockH.State)
	r.GET("/api/time", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"millis": time.Now().UnixMilli()})
	})
	r.GET("/api/preset-activities", presetH.List)
	apiAuth := r.Group("/api/auth")
	apiAuth.POST("/register", authH.APIRegister)
	apiAuth.POST("/login", authH.APILogin)
	apiAuth.POST("/refresh", authH.APIRefresh)
	apiAuth.POST("/logout", authH.APILogout)

	// --- protected page routes (session) ---
	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())

	protected.GET("/", func(c *gin.Context) { c.Redirect(http.StatusSeeOther, "/dashboard") })

	protected.GET("/dashboard", dashH.ShowDashboard)
	protected.GET("/children/:id", childCfgH.Show)
	protected.GET("/profiles/:id", profileCfgH.Show)

	form := protected.Group("/config")
	form.POST("/children", dashH.CreateChild)
	form.POST("/children/:id/delete", childCfgH.DeleteChild)
	form.POST("/children/:id/profiles", childCfgH.CreateProfile)
	form.POST("/children/:id/default-profile", childCfgH.SetDefaultProfile)
	form.POST("/children/:id/schedule/:day", childCfgH.AssignScheduleDay)
	form.POST("/profiles/:id/delete", profileCfgH.Delete)
	form.POST("/profiles/:id/activities", profileCfgH.AddActivity)
	form.POST("/activities/:id/delete", profileCfgH.DeleteActivity)
	form.POST("/children/:id/events", childCfgH.CreateEvent)
	form.POST("/events/:id/delete", childCfgH.DeleteEvent)
	form.POST("/children/:id/avatar", childCfgH.UploadAvatar)
	form.POST("/upload", handler.UploadHandler)

	// --- protected API routes (JWT Bearer) ---
	jwtSecret := []byte(cfg.JWTSecret)
	api := r.Group("/api")
	api.Use(middleware.RequireJWT(jwtSecret))

	// children
	api.GET("/children", childH.List)
	api.POST("/children", childH.Create)
	api.GET("/children/:id", childH.Get)
	api.PUT("/children/:id", childH.Update)
	api.DELETE("/children/:id", childH.Delete)
	api.PUT("/children/:id/default-profile", childH.SetDefaultProfile)

	// profiles
	api.GET("/children/:id/profiles", profileH.List)
	api.POST("/children/:id/profiles", profileH.Create)
	api.GET("/profiles/:id", profileH.Get)
	api.PUT("/profiles/:id", profileH.Update)
	api.DELETE("/profiles/:id", profileH.Delete)

	// activities
	api.POST("/profiles/:id/activities", profileH.AddActivity)
	api.PUT("/activities/:id", profileH.UpdateActivity)
	api.DELETE("/activities/:id", profileH.DeleteActivity)

	// schedule
	api.GET("/children/:id/schedule", scheduleH.Get)
	api.POST("/children/:id/schedule", scheduleH.Assign)
	api.DELETE("/children/:id/schedule/:day", scheduleH.ClearDay)

	// events
	api.GET("/children/:id/events", eventH.List)
	api.POST("/children/:id/events", eventH.Create)
	api.PUT("/events/:id", eventH.Update)
	api.DELETE("/events/:id", eventH.Delete)

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
