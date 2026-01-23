.PHONY: all build build-api build-scraper run run-api run-scraper test clean deps lint

# Go variables
GOCMD=go
GOBIN=$(PWD)/bin
API_BINARY=api
SCRAPER_BINARY=scraper

# Build flags
LDFLAGS=-s -w

all: deps build

deps:
	$(GOCMD) mod download
	$(GOCMD) mod tidy

build: build-api build-scraper

build-api:
	$(GOCMD) build -ldflags "$(LDFLAGS)" -o $(GOBIN)/$(API_BINARY) ./cmd/api

build-scraper:
	$(GOCMD) build -ldflags "$(LDFLAGS)" -o $(GOBIN)/$(SCRAPER_BINARY) ./cmd/scraper

run: build
	$(GOBIN)/$(API_BINARY)

run-api: build-api
	$(GOBIN)/$(API_BINARY)

run-scraper: build-scraper
	$(GOBIN)/$(SCRAPER_BINARY)

test:
	$(GOCMD) test -v ./...

test-coverage:
	$(GOCMD) test -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf $(GOBIN) coverage.out coverage.html

install-air:
	$(GOCMD) install github.com/air-verse/air@latest

air:
	air

dev: install-air air

lint:
	$(GOCMD) vet ./...

fmt:
	$(GOCMD) fmt ./...

.PHONY: migrate-up migrate-down

DATABASE_URL?=postgres://localhost/opengov?sslmode=disable

migrate-up:
	cat migrations/*.sql | psql "$(DATABASE_URL)"

migrate-down:
	@echo "Manual rollback required - drop and recreate database"
