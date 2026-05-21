package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	MigrationPath string
	DBUrl         string
	GHToken       string
	SMTPHost      string
	SMTPPort      int
	SMTPUser      string
	SMTPPass      string
	SMTPFrom      string
}

func NewConfig() *Config {
	portStr := os.Getenv("SMTP_PORT")
	port := 587
	if p, err := strconv.Atoi(portStr); err == nil {
		port = p
	}

	return &Config{
		MigrationPath: "migrations",
		DBUrl: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
		),
		GHToken:  os.Getenv("GH_TOKEN"),
		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: port,
		SMTPUser: os.Getenv("SMTP_USER"),
		SMTPPass: os.Getenv("SMTP_PASS"),
		SMTPFrom: os.Getenv("SMTP_FROM"),
	}
}


