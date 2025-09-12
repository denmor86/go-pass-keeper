package config

import (
	"fmt"

	"github.com/caarlos0/env"
	"github.com/spf13/pflag"
)

// Config модель настроек сервера
type Config struct {
	ListenAddr  string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	DatabaseDSN string `env:"DATABASE_URI" envDefault:""`
	JWTSecret   string `env:"JWT_SECRET" envDefault:"secret"`
}

// NewConfig - создание новой конфигурации
func NewConfig() *Config {

	var args Config
	if err := env.Parse(&args); err != nil {
		panic(fmt.Sprintf("Failed to parse enviroment var: %s", err.Error()))
	}

	var (
		server   = pflag.StringP("server", "a", args.ListenAddr, "Server listen address in a form host:port.")
		logLevel = pflag.StringP("log_level", "l", args.LogLevel, "Log level.")
		DSN      = pflag.StringP("dsn", "d", args.DatabaseDSN, "Database DSN")
		secret   = pflag.StringP("secret", "s", args.JWTSecret, "Secret to JWT")
	)
	pflag.Parse()

	return &Config{
		ListenAddr:  *server,
		LogLevel:    *logLevel,
		DatabaseDSN: *DSN,
		JWTSecret:   *secret,
	}
}
