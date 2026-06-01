package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Addr              string     // listen address, default ":8080"
	DatabasePath      string     // SQLite file path, default "~/.raevtar/data.db"
	Domain            string     // raevtar.tech
	LogLevel          slog.Level // debug | info | warn | error
	AdminKey          string     // API key for write operations (cron auto-post)
	AdminUser         string     // Admin panel username
	AdminPass         string     // Admin panel password
	MediaDir          string
	IsProduction      bool // production mode — stricter checks
	TrustedProxyCIDRs []string
}

func Load() *Config {
	cfg := &Config{
		Addr:              getEnv("RAEVTAR_ADDR", ":8080"),
		DatabasePath:      getEnv("RAEVTAR_DB", expandHome("~/.raevtar/data.db")),
		Domain:            getEnv("RAEVTAR_DOMAIN", "raevtar.tech"),
		LogLevel:          parseLogLevel(getEnv("RAEVTAR_LOG_LEVEL", "info")),
		AdminKey:          getEnv("RAEVTAR_ADMIN_KEY", ""),
		AdminUser:         getEnv("RAEVTAR_ADMIN_USER", "admin"),
		AdminPass:         getEnv("RAEVTAR_ADMIN_PASS", ""),
		MediaDir:          getEnv("RAEVTAR_MEDIA_DIR", expandHome("~/.raevtar/uploads")),
		IsProduction:      os.Getenv("RAEVTAR_ENV") == "production",
		TrustedProxyCIDRs: parseCSV(os.Getenv("RAEVTAR_TRUSTED_PROXY_CIDRS")),
	}

	if cfg.AdminKey == "" {
		slog.Warn("RAEVTAR_ADMIN_KEY kosong — API write endpoint gak akan bisa dipake")
	}

	if cfg.AdminPass == "" {
		slog.Warn("RAEVTAR_ADMIN_PASS kosong — admin panel login gak akan berfungsi")
	}

	return cfg
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config required")
	}
	if !c.IsProduction {
		return nil
	}
	missing := make([]string, 0, 2)
	if c.AdminKey == "" {
		missing = append(missing, "RAEVTAR_ADMIN_KEY")
	}
	if c.AdminPass == "" {
		missing = append(missing, "RAEVTAR_ADMIN_PASS")
	}
	if len(missing) > 0 {
		return fmt.Errorf("production config requires %s", strings.Join(missing, ", "))
	}
	return nil
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

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
