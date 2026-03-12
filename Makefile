VERSION ?= dev
LDFLAGS := -X main.version=$(VERSION)
BINARY := cc-filter
INSTALL_PATH := /usr/local/bin

.PHONY: build test clean install uninstall release

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY)

test:
	go test ./...

clean:
	rm -f $(BINARY)

install: build
	install -m 755 $(BINARY) $(INSTALL_PATH)/$(BINARY)

uninstall:
	rm -f $(INSTALL_PATH)/$(BINARY)

release: test
	@if [ "$(VERSION)" = "dev" ]; then echo "Usage: make release VERSION=v0.0.5-rc1"; exit 1; fi
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-amd64
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64
	@echo "Binaries in dist/"
