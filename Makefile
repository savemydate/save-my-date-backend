.PHONY: all build run test swag clean

all: test build

build:
	go build -v -o bin/server.exe cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test -v -race ./...

swag:
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

clean:
	if exist bin rmdir /s /q bin
