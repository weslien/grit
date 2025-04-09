.PHONY: install uninstall version build clean test
test:
	go test -v -cover ./...
clean:
	rm -rf bin

build:
	go build -o bin/grit

install:
	@echo "Installing grit..."
	@go install
	@echo "✅"

uninstall:
	rm -f $$(shell go env GOPATH)/bin/grit

schema:
	@echo "Generating grit schema..."
	go run cmd/schema/main.go

version:
	@go version