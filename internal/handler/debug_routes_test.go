package handler

import (
	"fmt"
	"net/http"
	"testing"
	"raevtar/internal/config"
	"raevtar/internal/repo"
	"raevtar/internal/service"

	"github.com/go-chi/chi/v5"
)

func TestDebugRoutes(t *testing.T) {
	cfg := config.Load()
	cfg.DatabasePath = "/tmp/test_debug_routes.db"
	
	db := repo.InitSQLite(cfg.DatabasePath)
	repos := repo.New(db)
	svc := service.New(repos, cfg)
	
	h := New(svc, cfg)
	
	// chi.Walk walks all routes
	mux, ok := h.(chi.Routes)
	if !ok {
		t.Fatalf("handler.New did not return chi.Routes (got %T)", h)
	}
	chi.Walk(mux, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		fmt.Printf("%-6s %s\n", method, route)
		return nil
	})
}
