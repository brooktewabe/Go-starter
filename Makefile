# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=main
BINARY_UNIX=$(BINARY_NAME)_unix

# Docker parameters
DOCKER_IMAGE=user-management-api
DOCKER_TAG=latest

.PHONY: all build clean test coverage deps lint docker-build docker-run docker-stop help

all: test build

## Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/server

## Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/server

## Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

## Run tests
test:
	$(GOTEST) -v ./...

## Run tests with coverage
coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## Run linter
lint:
	golangci-lint run

## Fix linting issues
lint-fix:
	golangci-lint run --fix

## Install linter
install-linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2

## Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/server
	./$(BINARY_NAME)

## Run with hot reload (requires air)
dev:
	air

## Install air for hot reload
install-air:
	go install github.com/cosmtrek/air@latest

## Generate Swagger docs
swagger:
	swag init -g cmd/server/main.go -o docs

## Install swagger
install-swagger:
	go install github.com/swaggo/swag/cmd/swag@latest

## Docker build
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

## Docker run with compose
docker-up:
	docker-compose up -d

## Docker stop
docker-down:
	docker-compose down

## Docker logs
docker-logs:
	docker-compose logs -f api

## Create uploads directory
create-dirs:
	mkdir -p uploads/images uploads/documents

## Setup development environment
setup: deps install-linter install-swagger install-air create-dirs
	@echo "Development environment setup complete!"

## Format code
fmt:
	$(GOCMD) fmt ./...

## Vet code
vet:
	$(GOCMD) vet ./...

## Security check
security:
	gosec ./...

## Install security checker
install-security:
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

## Full check (format, vet, lint, test)
check: fmt vet lint test

## Help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  build-linux   - Build for Linux"
	@echo "  clean         - Clean build files"
	@echo "  test          - Run tests"
	@echo "  coverage      - Run tests with coverage"
	@echo "  deps          - Download dependencies"
	@echo "  lint          - Run linter"
	@echo "  lint-fix      - Fix linting issues"
	@echo "  run           - Build and run the application"
	@echo "  dev           - Run with hot reload"
	@echo "  swagger       - Generate Swagger docs"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-up     - Start with Docker Compose"
	@echo "  docker-down   - Stop Docker Compose"
	@echo "  docker-logs   - View Docker logs"
	@echo "  setup         - Setup development environment"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  security      - Run security check"
	@echo "  check         - Run all checks"
	@echo "  help          - Show this help"