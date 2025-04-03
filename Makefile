.PHONY: install uninstall version build

build:
	go build -o bin/grit

install:
	go install

uninstall:
	rm -f $$(shell go env GOPATH)/bin/grit

version:
	@go version