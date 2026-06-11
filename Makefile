.PHONY: build install test clean

BINARY=pslens
INSTALL_DIR=$(HOME)/.local/bin

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/$(BINARY)

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY)
	rm -rf dist/

# Release: requires goreleaser and a git tag
release:
	goreleaser release --clean

.PHONY: help
help:
	@echo "Targets: build, install, test, vet, clean, release"