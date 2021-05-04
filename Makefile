VERSION := $(shell git describe --tags $(shell git rev-list --tags --max-count=1))-next
COMMIT := $(shell git rev-parse --short HEAD)
BUILDFLAGS = -ldflags "-X github.com/linuxsuren/cobra-extension/version.version=$(VERSION) \
	-X github.com/linuxsuren/cobra-extension/version.commit=$(COMMIT) \
	-X github.com/linuxsuren/cobra-extension/version.date=$(shell date +'%Y-%m-%d') -w -s"

simple-build:
	CGO_ENABLE=0 go build $(BUILDFLAGS) -o bin/ks

build: pre-build simple-build
	upx bin/ks

build-linux: pre-build
	CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build $(BUILDFLAGS) -o bin/linux/ks
	upx bin/linux/ks

build-plugin: pre-build
	CGO_ENABLE=0 go build ${BUILDFLAGS} -o bin/kubectl-ks kubectl-plugin/*.go
	upx bin/kubectl-ks

build-plugin-linux: pre-build
	CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build ${BUILDFLAGS} -o bin/linux/kubectl-ks kubectl-plugin/*.go
	upx bin/linux/kubectl-ks

pre-build: export GOPROXY=https://goproxy.io
pre-build: fmt lint mod-tidy

tools:  export GOPROXY=https://goproxy.io
tools:
	go get -u golang.org/x/lint/golint

mod-tidy:
	go mod tidy

fmt:
	go fmt ./...

lint:
	golint -set_exit_status ./...

copy: build
	sudo cp bin/ks /usr/local/bin/ks
simple-copy: simple-build
	sudo cp bin/ks /usr/local/bin/ks

copy-plugin: build-plugin
	sudo cp bin/kubectl-ks /usr/local/bin/kubectl-ks

goreleaser-test:
	goreleaser release --rm-dist --skip-publish --snapshot

image: build-linux
	cp bin/linux/ks build/ks
	docker build ./build -t surenpi/ks

image-push:
	docker push surenpi/ks

update:
	git fetch
	git reset --hard origin/$(shell git branch --show-current)
