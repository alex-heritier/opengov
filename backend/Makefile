.PHONY: all build run test clean deps lint

# Go variables
GOCMD=go
GOBIN=$(PWD)/bin
BINARY_NAME=server

# Build flags
LDFLAGS=-s -w

all: deps build

deps:
	$(GOCMD) mod download
	$(GOCMD) mod tidy

build:
	$(GOCMD) build -ldflags "$(LDFLAGS)" -o $(GOBIN)/$(BINARY_NAME) ./cmd/server

run: build
	$(GOBIN)/$(BINARY_NAME)

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
