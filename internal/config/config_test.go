package config

import (
	"net"
	"log/slog"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Validate — existing tests kept and extended
// ---------------------------------------------------------------------------

func TestValidateAllowsMissingSecretsOutsideProduction(t *testing.T) {
	cfg := &Config{}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestValidateRequiresProductionSecrets(t *testing.T) {
	cfg := &Config{IsProduction: true}
	err := cfg.Validate()
	if err == nil {
		t.Fatalf("Validate() error = nil, want missing secrets")
	}
	if !strings.Contains(err.Error(), "RAEVTAR_ADMIN_KEY") || !strings.Contains(err.Error(), "RAEVTAR_ADMIN_PASS") {
		t.Fatalf("Validate() error = %q, want both missing secret names", err.Error())
	}
}

func TestValidateAcceptsProductionSecrets(t *testing.T) {
	cfg := &Config{IsProduction: true, AdminKey: "key", AdminPass: "password"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

// ---------------------------------------------------------------------------
// Load — defaults
// ---------------------------------------------------------------------------

func TestLoadDefaults(t *testing.T) {
	cfg := Load()
	if cfg.Addr != ":8080" {
		t.Fatalf("Addr = %q, want %q", cfg.Addr, ":8080")
	}
	if cfg.Domain != "raevtar.tech" {
		t.Fatalf("Domain = %q, want %q", cfg.Domain, "raevtar.tech")
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Fatalf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelInfo)
	}
	if cfg.DatabasePath == "" {
		t.Fatalf("DatabasePath = empty, expected default path")
	}
	if cfg.MediaDir == "" {
		t.Fatalf("MediaDir = empty, expected default path")
	}
	if cfg.AdminUser != "admin" {
		t.Fatalf("AdminUser = %q, want %q", cfg.AdminUser, "admin")
	}
	if cfg.IsProduction {
		t.Fatalf("IsProduction = true, want false")
	}
}

// ---------------------------------------------------------------------------
// Load — custom env vars
// ---------------------------------------------------------------------------

func TestLoadCustom(t *testing.T) {
	t.Setenv("RAEVTAR_ADDR", ":9090")
	t.Setenv("RAEVTAR_DOMAIN", "example.com")
	t.Setenv("RAEVTAR_LOG_LEVEL", "debug")
	t.Setenv("RAEVTAR_ADMIN_USER", "myadmin")
	t.Setenv("RAEVTAR_ADMIN_KEY", "secret-key")
	t.Setenv("RAEVTAR_ADMIN_PASS", "secret-pass")

	cfg := Load()
	if cfg.Addr != ":9090" {
		t.Fatalf("Addr = %q, want %q", cfg.Addr, ":9090")
	}
	if cfg.Domain != "example.com" {
		t.Fatalf("Domain = %q, want %q", cfg.Domain, "example.com")
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Fatalf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelDebug)
	}
	if cfg.AdminUser != "myadmin" {
		t.Fatalf("AdminUser = %q, want %q", cfg.AdminUser, "myadmin")
	}
	if cfg.AdminKey != "secret-key" {
		t.Fatalf("AdminKey = %q, want %q", cfg.AdminKey, "secret-key")
	}
	if cfg.AdminPass != "secret-pass" {
		t.Fatalf("AdminPass = %q, want %q", cfg.AdminPass, "secret-pass")
	}
}

// ---------------------------------------------------------------------------
// Load — production secret checks
// ---------------------------------------------------------------------------

func TestLoadAdminKeyRequiredInProduction(t *testing.T) {
	t.Setenv("RAEVTAR_ENV", "production")
	t.Setenv("RAEVTAR_ADMIN_KEY", "")
	t.Setenv("RAEVTAR_ADMIN_PASS", "some-pass")

	cfg := Load()
	if !cfg.IsProduction {
		t.Fatalf("IsProduction = false, want true")
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate() = nil, want error for missing AdminKey")
	}
}

func TestLoadAdminPassRequiredInProduction(t *testing.T) {
	t.Setenv("RAEVTAR_ENV", "production")
	t.Setenv("RAEVTAR_ADMIN_KEY", "some-key")
	t.Setenv("RAEVTAR_ADMIN_PASS", "")

	cfg := Load()
	if !cfg.IsProduction {
		t.Fatalf("IsProduction = false, want true")
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate() = nil, want error for missing AdminPass")
	}
}

func TestLoadAdminKeyNotRequiredInDev(t *testing.T) {
	t.Setenv("RAEVTAR_ENV", "")
	t.Setenv("RAEVTAR_ADMIN_KEY", "")
	t.Setenv("RAEVTAR_ADMIN_PASS", "")

	cfg := Load()
	if cfg.IsProduction {
		t.Fatalf("IsProduction = true, want false")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil in dev mode", err)
	}
}

// ---------------------------------------------------------------------------
// Load — TrustedProxyCIDRs
// ---------------------------------------------------------------------------

func TestLoadTrustedProxyCIDRs(t *testing.T) {
	t.Setenv("RAEVTAR_TRUSTED_PROXY_CIDRS", "10.0.0.0/8,192.168.0.0/16")

	cfg := Load()
	if len(cfg.TrustedProxyCIDRs) != 2 {
		t.Fatalf("len(TrustedProxyCIDRs) = %d, want 2", len(cfg.TrustedProxyCIDRs))
	}
	if cfg.TrustedProxyCIDRs[0] != "10.0.0.0/8" {
		t.Fatalf("TrustedProxyCIDRs[0] = %q, want %q", cfg.TrustedProxyCIDRs[0], "10.0.0.0/8")
	}
	if cfg.TrustedProxyCIDRs[1] != "192.168.0.0/16" {
		t.Fatalf("TrustedProxyCIDRs[1] = %q, want %q", cfg.TrustedProxyCIDRs[1], "192.168.0.0/16")
	}
}

func TestLoadTrustedProxyCIDRsEmpty(t *testing.T) {
	t.Setenv("RAEVTAR_TRUSTED_PROXY_CIDRS", "")

	cfg := Load()
	if len(cfg.TrustedProxyCIDRs) != 0 {
		t.Fatalf("len(TrustedProxyCIDRs) = %d, want 0", len(cfg.TrustedProxyCIDRs))
	}
}

func TestLoadTrustedProxyCIDRsWhitespace(t *testing.T) {
	t.Setenv("RAEVTAR_TRUSTED_PROXY_CIDRS", "  10.0.0.0/8 ,  , 192.168.0.0/16  ")

	cfg := Load()
	if len(cfg.TrustedProxyCIDRs) != 2 {
		t.Fatalf("len(TrustedProxyCIDRs) = %d, want 2", len(cfg.TrustedProxyCIDRs))
	}
	if cfg.TrustedProxyCIDRs[0] != "10.0.0.0/8" {
		t.Fatalf("TrustedProxyCIDRs[0] = %q, want %q", cfg.TrustedProxyCIDRs[0], "10.0.0.0/8")
	}
}

// ---------------------------------------------------------------------------
// ParseCIDRs
// ---------------------------------------------------------------------------

func TestParseCIDRsMultiple(t *testing.T) {
	nets, err := ParseCIDRs("10.0.0.0/8,192.168.0.0/16")
	if err != nil {
		t.Fatalf("ParseCIDRs() error = %v, want nil", err)
	}
	if len(nets) != 2 {
		t.Fatalf("len = %d, want 2", len(nets))
	}
	if !nets[0].Contains(net.ParseIP("10.1.2.3")) {
		t.Fatalf("10.0.0.0/8 should contain 10.1.2.3")
	}
	if !nets[1].Contains(net.ParseIP("192.168.1.1")) {
		t.Fatalf("192.168.0.0/16 should contain 192.168.1.1")
	}
}

func TestParseCIDRsEmpty(t *testing.T) {
	nets, err := ParseCIDRs("")
	if err != nil {
		t.Fatalf("ParseCIDRs(\"\") error = %v, want nil", err)
	}
	if nets != nil {
		t.Fatalf("ParseCIDRs(\"\") = %v, want nil", nets)
	}
}

func TestParseCIDRsWhitespaceOnly(t *testing.T) {
	nets, err := ParseCIDRs("  ")
	if err != nil {
		t.Fatalf("ParseCIDRs(\"  \") error = %v, want nil", err)
	}
	if nets != nil {
		t.Fatalf("ParseCIDRs(\"  \") = %v, want nil", nets)
	}
}

func TestParseCIDRsInvalid(t *testing.T) {
	_, err := ParseCIDRs("not-a-cidr")
	if err == nil {
		t.Fatalf("ParseCIDRs(\"not-a-cidr\") = nil, want error")
	}
}

func TestParseCIDRsMixed(t *testing.T) {
	_, err := ParseCIDRs("10.0.0.0/8,invalid,192.168.0.0/16")
	if err == nil {
		t.Fatalf("ParseCIDRs() with invalid entry = nil, want error")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("ParseCIDRs() error = %q, want mention of invalid value", err.Error())
	}
}

// ---------------------------------------------------------------------------
// IsTrustedIP
// ---------------------------------------------------------------------------

func TestIsTrustedIPMatching(t *testing.T) {
	_, n1, _ := net.ParseCIDR("10.0.0.0/8")
	_, n2, _ := net.ParseCIDR("192.168.0.0/16")
	trusted := []*net.IPNet{n1, n2}

	if !IsTrustedIP(net.ParseIP("10.1.2.3"), trusted) {
		t.Fatalf("IsTrustedIP(10.1.2.3 in 10.0.0.0/8) = false, want true")
	}
	if !IsTrustedIP(net.ParseIP("192.168.1.1"), trusted) {
		t.Fatalf("IsTrustedIP(192.168.1.1 in 192.168.0.0/16) = false, want true")
	}
}

func TestIsTrustedIPNotMatching(t *testing.T) {
	_, n, _ := net.ParseCIDR("10.0.0.0/8")
	trusted := []*net.IPNet{n}

	if IsTrustedIP(net.ParseIP("8.8.8.8"), trusted) {
		t.Fatalf("IsTrustedIP(8.8.8.8) = true, want false")
	}
}

func TestIsTrustedIPEmptyList(t *testing.T) {
	if IsTrustedIP(net.ParseIP("10.1.2.3"), nil) {
		t.Fatalf("IsTrustedIP with nil list = true, want false")
	}
	if IsTrustedIP(net.ParseIP("10.1.2.3"), []*net.IPNet{}) {
		t.Fatalf("IsTrustedIP with empty list = true, want false")
	}
}

func TestIsTrustedIPNilIP(t *testing.T) {
	_, n, _ := net.ParseCIDR("10.0.0.0/8")
	trusted := []*net.IPNet{n}

	if IsTrustedIP(nil, trusted) {
		t.Fatalf("IsTrustedIP(nil) = true, want false")
	}
}

// ---------------------------------------------------------------------------
// LogLevel
// ---------------------------------------------------------------------------

func TestLogLevelDefault(t *testing.T) {
	t.Setenv("RAEVTAR_LOG_LEVEL", "")

	cfg := Load()
	if cfg.LogLevel != slog.LevelInfo {
		t.Fatalf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelInfo)
	}
}

func TestLogLevelCustom(t *testing.T) {
	t.Setenv("RAEVTAR_LOG_LEVEL", "warn")

	cfg := Load()
	if cfg.LogLevel != slog.LevelWarn {
		t.Fatalf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelWarn)
	}
}

func TestLogLevelUnknown(t *testing.T) {
	t.Setenv("RAEVTAR_LOG_LEVEL", "bogus")

	cfg := Load()
	if cfg.LogLevel != slog.LevelInfo {
		t.Fatalf("LogLevel = %v, want %v (fallback)", cfg.LogLevel, slog.LevelInfo)
	}
}
