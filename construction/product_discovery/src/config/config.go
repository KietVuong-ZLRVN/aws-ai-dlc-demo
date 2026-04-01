package config

import "os"

// Config holds application-level configuration.
type Config struct {
	Port string // env PORT, default "8080"
}

// Load reads configuration from the environment.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{
		Port: port,
	}
}
