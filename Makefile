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
	docker run -i \
	--rm --privileged=true \
	-p 8888:8888 -p 8080:8080 -p 9090:9090 \
	-v $(shell pwd)/docker-assets/prometheus:/etc/prometheus/ \
	-v $(shell pwd)/docker-assets/metrics:/www/ \
	--name panopticon-test \
	panopticon

gotest: fmt
	$(GO) test -v $(pkgs)

fmt:
	$(GO) fmt ./...
