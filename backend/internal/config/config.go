package config

import (
	"os"
)

// Config holds every external dependency's connection info, loaded from
// environment variables. See .env.example at the repo root for the full
// list with descriptions.
type Config struct {
	Port string

	DatabaseURL string
	RedisURL    string

	Auth0Domain   string
	Auth0Audience string

	AzureStorageAccount   string
	AzureStorageKey       string
	AzureStorageContainer string

	MeilisearchHost string
	MeilisearchKey  string

	ResendAPIKey string
	TwilioSID    string
	TwilioToken  string
	TwilioFrom   string

	PostHogAPIKey string
	PostHogHost   string
}

func Load() Config {
	return Config{
		Port: getEnv("PORT", "8080"),

		DatabaseURL: getEnv("DATABASE_URL", "postgres://yven:yven@localhost:5432/yven?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),

		Auth0Domain:   getEnv("AUTH0_DOMAIN", ""),
		Auth0Audience: getEnv("AUTH0_AUDIENCE", ""),

		AzureStorageAccount:   getEnv("AZURE_STORAGE_ACCOUNT", ""),
		AzureStorageKey:       getEnv("AZURE_STORAGE_KEY", ""),
		AzureStorageContainer: getEnv("AZURE_STORAGE_CONTAINER", "yven-uploads"),

		MeilisearchHost: getEnv("MEILISEARCH_HOST", "http://localhost:7700"),
		MeilisearchKey:  getEnv("MEILISEARCH_KEY", ""),

		ResendAPIKey: getEnv("RESEND_API_KEY", ""),
		TwilioSID:    getEnv("TWILIO_SID", ""),
		TwilioToken:  getEnv("TWILIO_TOKEN", ""),
		TwilioFrom:   getEnv("TWILIO_FROM", ""),

		PostHogAPIKey: getEnv("POSTHOG_API_KEY", ""),
		PostHogHost:   getEnv("POSTHOG_HOST", "https://app.posthog.com"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
