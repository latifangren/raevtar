BIN=raevtar
DB_PATH=$(HOME)/.raevtar/data.db
TEMPL_VERSION=v0.3.906
TEMPL=go run github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION)
TAILWIND_VERSION=3.4.19
TAILWIND=npx --yes tailwindcss@$(TAILWIND_VERSION)
RAEVTAR_USER?=latif
RAEVTAR_HOME?=$(HOME)/raevtar

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

generate-service:
	sed 's|{{RAEVTAR_USER}}|$(RAEVTAR_USER)|g; s|{{RAEVTAR_HOME}}|$(RAEVTAR_HOME)|g' raevtar.service.tmpl > raevtar.service

install-service: generate-service
	@echo "==> Install raevtar.service to /etc/systemd/system/ (requires sudo) =="
	sudo cp raevtar.service /etc/systemd/system/raevtar.service
	sudo systemctl daemon-reload

.PHONY: all css build test run dev templ-gen migrate seed clean db-reset generate-service install-service
