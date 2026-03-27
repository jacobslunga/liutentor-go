package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SupabaseURL        string
	SupabaseServiceKey string
	AppEnv             string
	Port               string
}

func Load() *Config {
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using system env")
		}
	}

	cfg := &Config{
		SupabaseURL:        os.Getenv("SUPABASE_URL"),
		SupabaseServiceKey: os.Getenv("SUPABASE_SERVICE_KEY"),
		AppEnv:             os.Getenv("APP_ENV"),
		Port:               os.Getenv("PORT"),
	}

	if cfg.SupabaseURL == "" || cfg.SupabaseServiceKey == "" {
		log.Fatal("Missing required environment variables")
	}

	return cfg
}
