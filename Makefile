.DEFAULT_GOAL := build

# Format Code
format:
	@echo "Formatting code:"
	go fmt ./...
.PHONY:format

# Check Code Style
lint: format
	@echo "Linting code:"
	staticcheck ./...
	shadow ./...
	go vet ./...
.PHONY:lint

# Test Code
test: lint
	@echo "Testing code:"
	go test ./...
.PHONY:test

# Update Dependencies
dependencies:
	@echo "Updating dependencies:"
	go get -u ./...
	go mod tidy
.PHONY:dependencies

# Install/Update Tools
tools:
	@echo "Installing/updating tools:"
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
	go install github.com/mitchellh/gox@latest
.PHONY:tools

# Build the command-line application
build:
	@echo "Building gpt command for local use:"
	go build -o gpt
.PHONY:build

# Install the command-line application
install:
	@echo "Installing from source:"
	go install

# Generate Documentation
docs: build
	@echo "Generating documentation:"
	rm ./docs/*.md
	./gpt docs
.PHONY:docs

# Create release binaries
# Valid OS/Arch values: https://go.dev/src/go/build/syslist.go
release:
	@echo "Creating release binaries:"
	mkdir -p dist
	gox -output="dist/{{.OS}}_{{.Arch}}/{{.Dir}}" -os="darwin linux windows" -arch="amd64 arm64"
.PHONY:release
