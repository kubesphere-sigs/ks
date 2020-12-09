VERSION := dev-$(shell git describe --tags $(shell git rev-list --tags --max-count=1))
COMMIT := $(shell git rev-parse --short HEAD)
BUILDFLAGS = -ldflags "-X github.com/linuxsuren/cobra-extension/version.version=$(VERSION) \
	-X github.com/linuxsuren/cobra-extension/version.commit=$(COMMIT) \
	-X github.com/linuxsuren/cobra-extension/version.date=$(shell date +'%Y-%m-%d') -w -s"

build: pre-build
	CGO_ENABLE=0 go build $(BUILDFLAGS) -o bin/ks

build-plugin: pre-build
	CGO_ENABLE=0 go build ${BUILDFLAGS} -o bin/kubectl-ks kubectl-plugin/*

pre-build: fmt lint

fmt:
	go fmt ./...

lint:
	golint ./...

copy: build
	sudo cp bin/ks /usr/local/bin/ks

copy-plugin: build-plugin
	sudo cp bin/kubectl-ks /usr/local/bin/kubectl-ks

goreleaser-test:
	goreleaser release --rm-dist --skip-publish --snapshot
