BINARY_NAME=goclaw
MODULE=github.com/StellariumFoundation/goclaw

.PHONY: build run clean fmt vet

build:
	go build -o $(BINARY_NAME) .

run:
	go run .

clean:
	rm -f $(BINARY_NAME)
	go clean

fmt:
	go fmt ./...

vet:
	go vet ./...
