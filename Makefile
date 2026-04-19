.PHONY: test bench lint fmt vet coverage docker-up docker-down build clean all ci
# Variables
BINARY_NAME=qrcode
GO=go
GOLANGCI_LINT=golangci-lint
DOCKER_COMPOSE=docker-compose
build:
        $(GO) build ./...
test:
        $(GO) test -race -count=1 ./...
bench:
        $(GO) test -run=^$$ -bench=. -benchmem ./...
lint:
        $(GOLANGCI_LINT) run ./...
fmt:
        $(GO) fmt ./...
vet:
        $(GO) vet ./...
coverage:
        $(GO) test -race -coverprofile=coverage.out ./...
        $(GO) tool cover -html=coverage.out -o coverage.html
clean:
        rm -f coverage.out coverage.html
all: fmt vet lint test
ci: lint vet test coverage
