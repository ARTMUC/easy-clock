package config

import (
	"os"
)

type Config struct {
	Port             string
	SessionSecret    string
	BrevoAPIKey      string
	BrevoSenderEmail string
	BrevoSenderName  string
	AppBaseURL       string
}

func New() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		SessionSecret:    getEnv("SESSION_SECRET", "change-me-in-production-32chars!!"),
		BrevoAPIKey:      os.Getenv("BREVO_API_KEY"),
		BrevoSenderEmail: getEnv("BREVO_SENDER_EMAIL", "noreply@example.com"),
		BrevoSenderName:  getEnv("BREVO_SENDER_NAME", "Starter App"),
		AppBaseURL:       getEnv("APP_BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
