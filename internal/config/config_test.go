package config

import (
	"strings"
	"testing"
)

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
