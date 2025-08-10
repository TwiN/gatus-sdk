# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Test parameters
COVERAGE_OUT=coverage.out
COVERAGE_HTML=coverage.html

.PHONY: all build clean test coverage coverage-html race fmt vet tidy help

# Default target
all: fmt vet test

# Build the project
build:
	$(GOBUILD) -v ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
coverage:
	$(GOTEST) -coverprofile=$(COVERAGE_OUT) -covermode=count ./...
	$(GOCMD) tool cover -func=$(COVERAGE_OUT)

# Generate HTML coverage report
coverage-html: coverage
	$(GOCMD) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Run tests with race detector
race:
	$(GOTEST) -race ./...

# Format Go code
fmt:
	$(GOFMT) ./...

# Run go vet
vet:
	$(GOVET) ./...

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Run golint if available
lint:
	@command -v golint >/dev/null 2>&1 && golint ./... || echo "golint not installed, skipping"

# Run staticcheck if available
staticcheck:
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

# Run all quality checks
quality: fmt vet lint staticcheck

# Continuous integration target
ci: quality test coverage race

# Development target - run before committing
dev: quality test coverage

# Show help
help:
	@echo "Available targets:"
	@echo "  all         - Format, vet, and test (default)"
	@echo "  build       - Build the project"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  coverage    - Run tests with coverage report"
	@echo "  coverage-html - Generate HTML coverage report"
	@echo "  race        - Run tests with race detector"
	@echo "  fmt         - Format Go code"
	@echo "  fmt         - Format Go code"
	@echo "  vet         - Run go vet"
	@echo "  tidy        - Tidy dependencies"
	@echo "  lint        - Run golint (if installed)"
	@echo "  staticcheck - Run staticcheck (if installed)"
	@echo "  quality     - Run all quality checks"
	@echo "  ci          - Run CI pipeline (quality + tests + coverage + race)"
	@echo "  dev         - Run development checks (quality + tests + coverage)"
	@echo "  help        - Show this help message"