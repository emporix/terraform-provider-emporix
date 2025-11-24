.PHONY: build install test clean fmt lint setup

# Default target
default: build

# Setup - regenerate go.sum with correct checksums
setup:
	@echo "Regenerating go.sum with correct checksums..."
	rm -f go.sum
	go mod download
	go mod tidy
	@echo "Setup complete!"

# Build the provider
build:
	go build -o terraform-provider-emporix

# Install the provider locally for development
install: build
	@echo "================================================"
	@echo "Installing Emporix Terraform Provider"
	@echo "================================================"
	@echo ""
	@echo "Detected platform: $$(go env GOOS)_$$(go env GOARCH)"
	@echo "Installation path: ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/"
	@echo ""
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	@cp terraform-provider-emporix ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/
	@chmod +x ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/terraform-provider-emporix
	@echo "✓ Provider binary copied"
	@echo "✓ Permissions set to executable"
	@echo ""
	@echo "Verifying installation..."
	@ls -lh ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/terraform-provider-emporix
	@echo ""
	@echo "✅ Installation complete!"
	@echo ""
	@echo "⚠️  IMPORTANT: If you're running Terraform under Rosetta on Apple Silicon,"
	@echo "    or experiencing platform detection issues, use 'make install-universal'"
	@echo ""
	@echo "Next steps:"
	@echo "  1. cd /your/terraform/project"
	@echo "  2. rm -rf .terraform .terraform.lock.hcl"
	@echo "  3. terraform init"
	@echo ""

# Install to both darwin_amd64 and darwin_arm64 (useful for Apple Silicon with Rosetta)
install-universal: build
	@echo "================================================"
	@echo "Universal Installation (macOS)"
	@echo "================================================"
	@echo ""
	@echo "Installing to both darwin_amd64 and darwin_arm64..."
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/darwin_amd64
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/darwin_arm64
	@cp terraform-provider-emporix ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/darwin_amd64/
	@cp terraform-provider-emporix ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/darwin_arm64/
	@chmod +x ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/darwin_amd64/terraform-provider-emporix
	@chmod +x ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/darwin_arm64/terraform-provider-emporix
	@echo ""
	@echo "✅ Installed to both architectures:"
	@ls -lh ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/darwin_*/terraform-provider-emporix
	@echo ""
	@echo "This covers both native ARM64 and Rosetta (x86_64) execution."
	@echo ""

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -cover ./...

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f terraform-provider-emporix
	rm -rf dist/

# Download dependencies
deps:
	go mod download
	go mod tidy

# Generate documentation (if using terraform-plugin-docs)
docs:
	@which tfplugindocs > /dev/null 2>&1 || echo "tfplugindocs not installed. Install with: go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest"
	@which tfplugindocs > /dev/null 2>&1 && tfplugindocs generate || true

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o dist/terraform-provider-emporix_linux_amd64
	GOOS=darwin GOARCH=amd64 go build -o dist/terraform-provider-emporix_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build -o dist/terraform-provider-emporix_darwin_arm64
	GOOS=windows GOARCH=amd64 go build -o dist/terraform-provider-emporix_windows_amd64.exe

# Help target
help:
	@echo "Available targets:"
	@echo "  setup               - Regenerate go.sum with correct checksums (run this first!)"
	@echo "  build               - Build the provider"
	@echo "  install             - Build and install the provider locally"
	@echo "  install-universal   - Install to both darwin_amd64 and darwin_arm64 (macOS)"
	@echo "  test                - Run tests"
	@echo "  test-coverage       - Run tests with coverage"
	@echo "  fmt                 - Format code"
	@echo "  lint                - Run linter"
	@echo "  clean               - Clean build artifacts"
	@echo "  deps                - Download and tidy dependencies"
	@echo "  docs                - Generate documentation"
	@echo "  build-all           - Build for all platforms"
	@echo "  help                - Show this help message"
