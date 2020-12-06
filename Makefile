build: fmt
	CGO_ENABLE=0 go build -ldflags "-w -s" -o bin/ks

fmt:
	go fmt ./...

copy: build
	sudo cp bin/ks /usr/local/bin/ks
