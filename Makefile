BIN=raevtar
DB_PATH=$(HOME)/.raevtar/data.db
TEMPL_VERSION=v0.3.906
TEMPL=go run github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION)
TAILWIND_VERSION=3.4.19
TAILWIND=npx --yes tailwindcss@$(TAILWIND_VERSION)

all: build

css:
	$(TAILWIND) -i static/css/tailwind.src.css -o static/css/style.css --minify

build: templ-gen css
	go build -o $(BIN) ./cmd/server/

test: templ-gen
	go test ./...

run: build
	./$(BIN)

dev:
	@echo "Watching for changes... (requires entr or similar)"
	find . -name '*.go' -o -name '*.templ' | entr -r sh -c '$(TEMPL) generate && go run ./cmd/server/'

templ-gen:
	$(TEMPL) generate

migrate:
	@echo "Migrations run automatically on startup"

seed:
	go run ./cmd/server/ -seed

clean:
	rm -f $(BIN)

db-reset:
	rm -f $(DB_PATH)

.PHONY: all css build test run dev templ-gen migrate seed clean db-reset
