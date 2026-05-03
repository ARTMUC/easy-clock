package router

import (
	"net/http"

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

	// --- persistence ---
	userRepo := userpersistence.NewUserRepository(gormDB)
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
	userSvc := userapplication.NewService(userRepo, bus, emailClient, cfg.AppBaseURL)

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

	// --- public routes ---
	r.GET("/login", authH.ShowLogin)
	r.POST("/login", authH.HandleLogin)
	r.GET("/register", authH.ShowRegister)
	r.POST("/register", authH.HandleRegister)
	r.GET("/verify", authH.HandleVerify)
	r.POST("/logout", authH.Logout)

	r.GET("/clock/:token", clockH.Show)
	r.GET("/api/clock/:token", clockH.State)

	// --- protected routes ---
	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())

	protected.GET("/", func(c *gin.Context) { c.Redirect(http.StatusSeeOther, "/dashboard") })

	dashH := handler.NewDashboardHandler()
	protected.GET("/dashboard", dashH.ShowDashboard)

	// children
	api := protected.Group("/api")
	api.GET("/children", childH.List)
	api.POST("/children", childH.Create)
	api.GET("/children/:id", childH.Get)
	api.PUT("/children/:id", childH.Update)
	api.DELETE("/children/:id", childH.Delete)
	api.PUT("/children/:id/default-profile", childH.SetDefaultProfile)

	// profiles
	api.GET("/children/:childID/profiles", profileH.List)
	api.POST("/children/:childID/profiles", profileH.Create)
	api.GET("/profiles/:id", profileH.Get)
	api.PUT("/profiles/:id", profileH.Update)
	api.DELETE("/profiles/:id", profileH.Delete)

	// activities
	api.POST("/profiles/:id/activities", profileH.AddActivity)
	api.PUT("/activities/:id", profileH.UpdateActivity)
	api.DELETE("/activities/:id", profileH.DeleteActivity)

	// schedule
	api.GET("/children/:childID/schedule", scheduleH.Get)
	api.POST("/children/:childID/schedule", scheduleH.Assign)
	api.DELETE("/children/:childID/schedule/:day", scheduleH.ClearDay)

	// events
	api.GET("/children/:childID/events", eventH.List)
	api.POST("/children/:childID/events", eventH.Create)
	api.PUT("/events/:id", eventH.Update)
	api.DELETE("/events/:id", eventH.Delete)

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
