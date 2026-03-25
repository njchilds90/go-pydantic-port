.PHONY: build test lint fmt vet bench coverage

build:
	go build ./...

test:
	go test -race ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .

vet:
	go vet ./...

bench:
	go test -bench=. -run=^$ ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
