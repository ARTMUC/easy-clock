package router

import (
	"fmt"
	"net/http"
	"time"

	"easy-clock/internal/config"
	database "easy-clock/internal/db"

	"gorm.io/gorm"
)

type Server struct {
	port   int
	db     database.Service
	gormDB *gorm.DB
	cfg    *config.Config
}

func NewServer() *http.Server {
	cfg := config.New()

	dbService := database.New()
	if dbService == nil {
		panic("database.New() returned nil — db initialization failed")
	}

	gormDB := dbService.GetDB()
	if gormDB == nil {
		panic("dbService.GetDB() returned nil — Gorm db is not initialized")
	}

	port := 8080
	if p := cfg.Port; p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	srv := &Server{
		port:   port,
		db:     dbService,
		gormDB: gormDB,
		cfg:    cfg,
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", srv.port),
		Handler:      srv.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return httpServer
}
