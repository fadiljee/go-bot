.PHONY: run test lint build tidy deadcode docker-up docker-down

GOPATH_BIN := $(shell go env GOPATH)/bin

run:
	go run ./cmd/server

test:
	go test -race -cover ./...

lint:
	PATH="$(GOPATH_BIN):$$PATH" golangci-lint run
	PATH="$(GOPATH_BIN):$$PATH" staticcheck ./...
	go vet ./...

deadcode:
	PATH="$(GOPATH_BIN):$$PATH" deadcode -test ./...

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bot-istimewa ./cmd/server
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o scraper-cli ./cmd/scraper-cli

tidy:
	go mod tidy

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

