package config

import (
	"os"
	"strings"
)

type Config struct {
	Port string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8081"
	} else if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	return Config{Port: port}
}
