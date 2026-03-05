# Makefile for go-pydantic-port library

.PHONY: build test lint fmt vet bench coverage

build:
	go build -o go-pydantic-port ./...

test:
	go test -race ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...

vet:
	go vet ./...

bench:
	go test -bench=. ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
