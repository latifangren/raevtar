package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	StaticDir         string // absolute path to static/ directory
	AgentDir          string // agent install dir, default "/usr/local/bin"

	// Hardening / rate limits
	RateLimitRequests   int           // max requests per window (default 60)
	RateLimitWindow     time.Duration // window duration (default 1m)
	ReadTimeout         time.Duration // HTTP read timeout (default 10s)
	WriteTimeout        time.Duration // HTTP write timeout (default 30s)
	IdleTimeout         time.Duration // HTTP idle timeout (default 60s)
	ShutdownTimeout     time.Duration // graceful shutdown timeout (default 15s)
	MaxUploadMB         int           // max upload size in MB (default 6)
	LoginFailureLimit   int           // max login failures per user/IP combo (default 5)
	LoginIPFailureLimit int           // max login failures per IP (default 20)
}

func Load() *Config {
	// Compute StaticDir from binary location so it's not CWD-dependent
	execPath, _ := os.Executable()
	staticDir := filepath.Join(filepath.Dir(execPath), "static")

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
		StaticDir:         getEnv("RAEVTAR_STATIC_DIR", staticDir),
		AgentDir:          getEnv("RAEVTAR_AGENT_DIR", "/usr/local/bin"),

		// Hardening / rate limits
		RateLimitRequests:   getEnvInt("RAEVTAR_RATE_LIMIT_REQUESTS", 60),
		RateLimitWindow:     getEnvDuration("RAEVTAR_RATE_LIMIT_WINDOW", 60*time.Second),
		ReadTimeout:         getEnvDuration("RAEVTAR_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:        getEnvDuration("RAEVTAR_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:         getEnvDuration("RAEVTAR_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout:     getEnvDuration("RAEVTAR_SHUTDOWN_TIMEOUT", 15*time.Second),
		MaxUploadMB:         getEnvInt("RAEVTAR_MAX_UPLOAD_MB", 6),
		LoginFailureLimit:   getEnvInt("RAEVTAR_LOGIN_FAILURE_LIMIT", 5),
		LoginIPFailureLimit: getEnvInt("RAEVTAR_LOGIN_IP_FAILURE_LIMIT", 20),
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

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
