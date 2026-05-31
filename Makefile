BIN=raevtar
DB_PATH=$(HOME)/.raevtar/data.db

all: build

css:
	npx tailwindcss -i static/css/tailwind.src.css -o static/css/style.css --minify

build: css
	go build -o $(BIN) ./cmd/server/

run: build
	./$(BIN)

dev:
	@echo "Watching for changes... (requires entr or similar)"
	find . -name '*.go' -o -name '*.templ' | entr -r go run ./cmd/server/

templ-gen:
	templ generate

migrate:
	@echo "Migrations run automatically on startup"

seed:
	go run ./cmd/server/ -seed

clean:
	rm -f $(BIN)

db-reset:
	rm -f $(DB_PATH)

.PHONY: all build run dev templ-gen migrate seed clean db-reset
