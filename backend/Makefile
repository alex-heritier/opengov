.PHONY: all build build-api build-scraper run run-api run-scraper test clean deps lint

# Go variables
GOCMD=go
GOBIN=$(PWD)/dist
API_BINARY=api
SCRAPER_BINARY=scraper

all: deps build

deps:
	$(GOCMD) mod download
	$(GOCMD) mod tidy

build: build-api build-scraper

build-api:
	./bin/build-api.sh

build-scraper:
	./bin/build-scraper.sh

run: build
	./bin/run-api.sh

run-api: build-api
	./bin/run-api.sh

run-scraper: build-scraper
	./bin/run-scraper.sh

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
