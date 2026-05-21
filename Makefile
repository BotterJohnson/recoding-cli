BINARY=recoding-cli
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build run test lint clean

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) .

run:
	go run . $(ARGS)

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -f $(BINARY)
