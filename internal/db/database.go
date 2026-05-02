package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Service represents a service that interacts with a database.
type Service interface {
	Health() map[string]string
	Close() error
	GetDB() *gorm.DB
}

type service struct {
	db *gorm.DB
}

// New connects to the database and returns a Service.
// Migrations are NOT run automatically – use `cmd/migrate` to apply them.
func New() Service {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		os.Getenv("GONE_DB_USERNAME"),
		os.Getenv("GONE_DB_PASSWORD"),
		os.Getenv("GONE_DB_HOST"),
		os.Getenv("GONE_DB_PORT"),
		os.Getenv("GONE_DB_DATABASE"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	return &service{db: db}
}

func (s *service) GetDB() *gorm.DB {
	return s.db
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlDB, _ := s.db.DB()
	if err := sqlDB.PingContext(ctx); err != nil {
		return map[string]string{"status": "down", "error": err.Error()}
	}
	return map[string]string{"status": "up"}
}

func (s *service) Close() error {
	sqlDB, _ := s.db.DB()
	return sqlDB.Close()
}
