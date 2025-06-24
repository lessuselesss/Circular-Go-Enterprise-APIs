.PHONY: test test-unit test-integration test-all clean

# Run all tests (unit + integration)
test-all:
	go test -tags=integration ./...

# Run only unit tests (excludes integration tests)
test-unit:
	go test ./...

# Run only integration tests
test-integration:
	go test -tags=integration ./test/integration/... ./api/...

# Run tests with verbose output
test-verbose:
	go test -v -tags=integration ./...

# Run unit tests with verbose output
test-unit-verbose:
	go test -v ./...

# Run integration tests with verbose output
test-integration-verbose:
	go test -v -tags=integration ./test/integration/... ./api/...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out -tags=integration ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run unit tests with coverage
test-unit-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Unit test coverage report generated: coverage.html"

# Clean up generated files
clean:
	rm -f coverage.out coverage.html

# Help target
help:
	@echo "Available targets:"
	@echo "  test-all              - Run all tests (unit + integration)"
	@echo "  test-unit             - Run only unit tests"
	@echo "  test-integration      - Run only integration tests"
	@echo "  test-verbose          - Run all tests with verbose output"
	@echo "  test-unit-verbose     - Run unit tests with verbose output"
	@echo "  test-integration-verbose - Run integration tests with verbose output"
	@echo "  test-coverage         - Run all tests with coverage report"
	@echo "  test-unit-coverage    - Run unit tests with coverage report"
	@echo "  clean                 - Clean up generated files"
	@echo "  help                  - Show this help message" 