.PHONY: build install clean test

build:
	go build -o mirusync .

install: build
	sudo mv mirusync /usr/local/bin/

clean:
	rm -f mirusync
	go clean

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

deps:
	go mod download
	go mod tidy

