# Makefile
BINARY_NAME=heimdall

build:
	go build -o bin/$(BINARY_NAME) cmd/heimdall/main.go

run:
	go run cmd/heimdall/main.go

clean:
	go clean
	rm -f bin/$(BINARY_NAME)