.PHONY: test integration api e2e all clean

# Load environment from .env.test if needed
ENV_TEST := source .env.test &&

# Run unit tests with coverage
test:
	@echo "Running unit tests..."
	source .env.test && go test ./tests/unit/... -v -cover

# Run integration tests with testcontainers
integration:
	@echo "Running integration tests..."
	source .env.test && go test ./tests/integration/... -v

# Run API tests
api:
	@echo "Running API tests..."
	source .env.test && go test ./tests/api/... -v

# Run E2E tests (end-to-end with real router)
e2e:
	@echo "Running E2E tests with real router..."
	source .env.test && go test ./tests/e2e/... -v -timeout 30s

# Run all tests (unit, integration, api, e2e)
all: test api integration e2e
	@echo "✅ All tests completed!"

# Clean test cache
clean:
	@echo "Cleaning test cache..."
	go clean -testcache
	rm -f coverage.out coverage.html

# Run tests with short flag (skip integration)
short:
	@echo "Running short tests (unit + api)..."
	source .env.test && go test ./tests/unit/... ./tests/api/... -v -short

# Run tests with coverage report and open in browser
coverage:
	@echo "Running tests with coverage report..."
	source .env.test && go test ./tests/unit/... ./tests/api/... ./tests/e2e/... -v -coverprofile=coverage.out
	@echo "Coverage report generated!"

# Generate and open HTML coverage report
coverage-html: coverage
	@echo "Opening coverage report in browser..."
	go tool cover -html=coverage.out -o coverage.html
	@open coverage.html || xdg-open coverage.html || echo "Please open coverage.html manually"

# Run benchmarks
bench:
	@echo "Running performance benchmarks..."
	source .env.test && go test ./tests/benchmarks/... -bench=. -benchmem -v

# Run benchmarks with memory statistics
bench-mem:
	@echo "Running benchmarks with detailed memory stats..."
	source .env.test && go test ./tests/benchmarks/... -bench=. -benchmem -memprofile=mem.prof

# Run specific test
test-one:
	@read -p "Enter test name: " testname; \
	source .env.test && go test ./tests/unit/... -v -run "$$testname"

# Run integration tests only (requires Docker)
integration-only:
	@echo "Running integration tests only..."
	source .env.test && go test ./tests/integration/... -v -count=1

# Watch mode for tests (requires entr or similar)
watch:
	@echo "Watching test files for changes..."
	ls tests/**/*_test.go | entr -r make short

# Run only AI module tests (unit, api, e2e)
test-ai:
	@echo "Running AI module tests..."
	source .env.test && go test ./tests/unit/ai_service_test.go ./tests/api/ai_api_test.go ./tests/e2e/ai_e2e_test.go -v

# Run AI tests with coverage
test-ai-coverage:
	@echo "Running AI module tests with coverage..."
	source .env.test && go test ./tests/unit/ai_service_test.go ./tests/api/ai_api_test.go ./tests/e2e/ai_e2e_test.go -v -coverprofile=coverage-ai.out
	go tool cover -html=coverage-ai.out -o coverage-ai.html
	@open coverage-ai.html || xdg-open coverage-ai.html || echo "Please open coverage-ai.html manually"

.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Core Tests:"
	@echo "  make test              - Run unit tests"
	@echo "  make api               - Run API tests"
	@echo "  make e2e               - Run E2E tests (end-to-end with real router)"
	@echo "  make integration       - Run integration tests (requires Docker)"
	@echo ""
	@echo "Test Combinations:"
	@echo "  make all               - Run all tests (unit + api + integration + e2e)"
	@echo "  make short             - Run unit + api tests (skip integration)"
	@echo "  make test-ai           - Run AI module tests (unit + api + e2e)"
	@echo ""
	@echo "Coverage & Performance:"
	@echo "  make coverage          - Generate coverage report"
	@echo "  make coverage-html     - Generate and open HTML coverage report"
	@echo "  make test-ai-coverage  - AI module coverage report"
	@echo "  make bench             - Run performance benchmarks"
	@echo "  make bench-mem         - Run benchmarks with memory profile"
	@echo ""
	@echo "Utilities:"
	@echo "  make test-one          - Run specific test by name"
	@echo "  make watch             - Watch and run tests on file changes"
	@echo "  make integration-only  - Run integration tests only"
	@echo "  make clean             - Clean test cache"
	@echo "  make help              - Show this help message"
	@echo ""
	@echo "Configuration:"
	@echo "  - Tests use .env.test for environment variables"
	@echo "  - Coverage reports saved to coverage.html and coverage-ai.html"
	@echo "  - E2E tests require GROQ_API_KEY in .env.test"
	@echo "  - Integration tests require Docker"
