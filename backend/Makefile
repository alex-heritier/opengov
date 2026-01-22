.PHONY: all build build-api build-scraping run run-api run-scraping test clean deps lint

# Go variables
GOCMD=go
GOBIN=$(PWD)/bin
API_BINARY=api
SCRAPING_BINARY=scraping

# Build flags
LDFLAGS=-s -w

all: deps build

deps:
	$(GOCMD) mod download
	$(GOCMD) mod tidy

build: build-api build-scraping

build-api:
	$(GOCMD) build -ldflags "$(LDFLAGS)" -o $(GOBIN)/$(API_BINARY) ./cmd/api

build-scraping:
	$(GOCMD) build -ldflags "$(LDFLAGS)" -o $(GOBIN)/$(SCRAPING_BINARY) ./cmd/scraping

run: build
	$(GOBIN)/$(API_BINARY)

run-api: build-api
	$(GOBIN)/$(API_BINARY)

run-scraping: build-scraping
	$(GOBIN)/$(SCRAPING_BINARY)

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
