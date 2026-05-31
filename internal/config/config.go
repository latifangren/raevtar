package config

import (
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Addr         string       // listen address, default ":8080"
	DatabasePath string       // SQLite file path, default "~/.raevtar/data.db"
	Domain       string       // raevtar.tech
	LogLevel     slog.Level   // debug | info | warn | error
	AdminKey     string       // API key for write operations (cron auto-post)
	AdminUser    string       // Admin panel username
	AdminPass    string       // Admin panel password
	IsProduction bool         // production mode — stricter checks
}

func Load() *Config {
	cfg := &Config{
		Addr:         getEnv("RAEVTAR_ADDR", ":8080"),
		DatabasePath: getEnv("RAEVTAR_DB", expandHome("~/.raevtar/data.db")),
		Domain:       getEnv("RAEVTAR_DOMAIN", "raevtar.tech"),
		LogLevel:     parseLogLevel(getEnv("RAEVTAR_LOG_LEVEL", "info")),
		AdminKey:     getEnv("RAEVTAR_ADMIN_KEY", ""),
		AdminUser:    getEnv("RAEVTAR_ADMIN_USER", "admin"),
		AdminPass:    getEnv("RAEVTAR_ADMIN_PASS", ""),
		IsProduction: os.Getenv("RAEVTAR_ENV") == "production",
	}

	if cfg.AdminKey == "" {
		slog.Warn("RAEVTAR_ADMIN_KEY kosong — API write endpoint gak akan bisa dipake")
	}

	if cfg.AdminPass == "" {
		slog.Warn("RAEVTAR_ADMIN_PASS kosong — admin panel login gak akan berfungsi")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func expandHome(path string) string {
	if len(path) > 1 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		return home + path[1:]
	}
	return path
}

func parseLogLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
