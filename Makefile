GO = go
pkgs = $(shell $(GO) list ./... | grep -v /vendor/)
BIN_DIR ?= $(shell pwd)

info:
	@echo "build:  Go build"
	@echo "docker: build and run in docker container"
	@echo "gotest: run go tests and reformats"
	@echo "fmt:    reformats code"

build: gotest
	$(GO) build -o panopticon

docker: gotest build
	docker build -t panopticon:latest .
	docker run --rm --privileged=true -p 8888:8888 -i --name panopticon-test panopticon 

gotest: fmt
	$(GO) test -v $(pkgs)

fmt:
	$(GO) fmt
