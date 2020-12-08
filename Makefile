VERSION := dev-$(shell git describe --tags $(shell git rev-list --tags --max-count=1))
COMMIT := $(shell git rev-parse --short HEAD)
BUILDFLAGS = -ldflags "-X github.com/linuxsuren/cobra-extension/version.version=$(VERSION) \
	-X github.com/linuxsuren/cobra-extension/version.commit=$(COMMIT) \
	-X github.com/linuxsuren/cobra-extension/version.date=$(shell date +'%Y-%m-%d') -w -s"

build: fmt
	CGO_ENABLE=0 go build $(BUILDFLAGS) -o bin/ks

build-plugin: fmt
	CGO_ENABLE=0 go build ${BUILDFLAGS} -o bin/kubectl-ks kubectl-plugin/*

fmt:
	go fmt ./...

copy: build
	sudo cp bin/ks /usr/local/bin/ks

copy-plugin: build-plugin
	sudo cp bin/kubectl-ks /usr/local/bin/kubectl-ks

goreleaser-test:
	goreleaser release --rm-dist --skip-publish --snapshot
