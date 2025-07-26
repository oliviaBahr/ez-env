.PHONY: build install clean test

# Build the binary
build:
	go build -o git-ez-env

# Install to /usr/local/bin (for Homebrew)
install: build
	@echo "Installing to /usr/local/bin..."
	@if [ ! -w /usr/local/bin ]; then \
		echo "Error: Cannot write to /usr/local/bin"; \
		echo "Please run with sudo or install to a different directory"; \
		exit 1; \
	fi
	cp git-ez-env /usr/local/bin/
	chmod +x /usr/local/bin/git-ez-env
	@echo "✓ ez-env installed successfully!"
	@echo ""
	@echo "You can now use:"
	@echo "  git ez-env init"
	@echo "  git ez-env add <file>"

# Clean build artifacts
clean:
	rm -f git-ez-env

# Run tests
test:
	go test ./...

# Build for release (stripped binary)
release: build
	strip git-ez-env

# Build for different platforms
build-linux:
	GOOS=linux GOARCH=amd64 go build -o git-ez-env-linux-amd64

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o git-ez-env-darwin-amd64

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o git-ez-env-darwin-arm64 