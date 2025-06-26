package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SupabaseURL      string
	SupabaseSvcKey   string
	GeminiAPIKey     string
	UnidocLicenseKey string
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	AppConfig = &Config{
		SupabaseURL:      getEnv("SUPABASE_URL", ""),
		SupabaseSvcKey:   getEnv("SUPABASE_SERVICE_KEY", ""),
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),
		UnidocLicenseKey: getEnv("UNIDOC_LICENSE_KEY", ""),
	}

	if AppConfig.SupabaseURL == "" || AppConfig.SupabaseSvcKey == "" || AppConfig.GeminiAPIKey == "" || AppConfig.UnidocLicenseKey == "" {
		log.Fatal("SUPABASE_URL and SUPABASE_SERVICE_KEY and GEMINI_API_KEY and UNIDOC_LICENSE_KEY must be set")
	}

}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
