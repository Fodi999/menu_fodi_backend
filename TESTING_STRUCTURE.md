# Testing Structure - Uber Engineering Style

This project implements a professional testing structure inspired by Uber Engineering best practices.

## 📁 Test Structure

```
tests/
├── unit/              # Unit tests for service logic
│   ├── *_service_test.go     # 21 modules
├── integration/       # Integration tests with PostgreSQL (testcontainers-go)
│   ├── *_repo_test.go        # 21 modules
└── api/              # API endpoint tests with httptest
    └── *_api_test.go         # 21 modules
```

## 🧪 Test Types

### Unit Tests (`tests/unit/`)
- Test individual service logic in isolation
- Fast execution (milliseconds)
- No external dependencies
- Use mocking for external services
- Files: `*_service_test.go`

**Example:** Testing business logic of Academy service
```bash
make test
```

### Integration Tests (`tests/integration/`)
- Test repository layer with real PostgreSQL database
- Use **testcontainers-go** for containerized databases
- Verify database operations and queries
- Files: `*_repo_test.go`

**Requirements:**
- Docker must be running
- PostgreSQL image will be pulled automatically

```bash
make integration
```

### API Tests (`tests/api/`)
- Test HTTP endpoints and handlers
- Use **httptest** for request/response testing
- Verify routes, status codes, and responses
- Files: `*_api_test.go`

```bash
make api
```

## 🚀 Running Tests

### Run All Tests
```bash
make all
```

### Run Specific Test Suite
```bash
# Unit tests only
make test

# Integration tests only (requires Docker)
make integration

# API tests only
make api

# Short run (unit + api, skip integration)
make short
```

### Coverage Report
```bash
# Generate HTML coverage report
make coverage

# This creates coverage.html with visual coverage representation
```

### Run Specific Test
```bash
make test-one
# Then enter the test name when prompted
```

### Watch Mode
```bash
# Requires 'entr' package: brew install entr
make watch
```

## 📋 Modules Tested (21 Total)

Each module has three test files:

1. **Academy** - Course management
   - `tests/unit/academy_service_test.go`
   - `tests/integration/academy_repo_test.go`
   - `tests/api/academy_api_test.go`

2. **Admin** - Administration
3. **AI** - AI features
4. **AI Core** - AI engine core
5. **Auth** - Authentication & JWT
6. **Business** - Business logic
7. **Contact** - Contact management
8. **Fridge** - Fridge/inventory
9. **Health** - Health checks
10. **Hint** - Hints system
11. **Ingredients** - Ingredient management
12. **Leaderboard** - Ranking system
13. **Marketplace** - Marketplace
14. **Meal Plan** - Meal planning
15. **Metrics** - Metrics collection
16. **Nutrition** - Nutrition analysis
17. **Recipes** - Recipe management
18. **Semi Finished** - Semi-finished products
19. **Stats** - Statistics
20. **User** - User management
21. **Wallet** - Wallet/payments

## 🔧 Testing Dependencies

Required Go packages (add to `go.mod`):

```bash
go get github.com/stretchr/testify
go get github.com/testcontainers/testcontainers-go
```

## 📝 Test Configuration

### Environment File: `.env.test`
Contains test-specific configuration:
- Database connection (PostgreSQL on port 5433)
- JWT secret for testing
- API port (8888)
- Groq AI test key
- Logging level

Load before running tests:
```bash
source .env.test
go test ./tests/...
```

Or automatically via docker-compose (optional):
```bash
docker-compose -f docker-compose.test.yml up
make all
```

## 💡 Best Practices

### Unit Tests
✅ Use `testify/assert` for assertions
✅ Mock external dependencies
✅ Test error cases
✅ Keep tests isolated

### Integration Tests
✅ Use testcontainers for reproducible environments
✅ Clean up resources with `defer container.Terminate()`
✅ Use same schema as production
✅ Skip with `-short` flag if needed

### API Tests
✅ Use `httptest.NewRecorder()`
✅ Test all HTTP methods (GET, POST, PUT, DELETE)
✅ Verify status codes
✅ Check response bodies

## 📊 Coverage Goals

- **Unit Tests:** 80%+ coverage
- **Integration Tests:** 60%+ coverage  
- **API Tests:** 75%+ coverage

View coverage:
```bash
make coverage
# Opens coverage.html in browser
```

## 🐛 Debugging Tests

### Verbose Output
```bash
go test ./tests/unit/... -v
```

### Specific Test with Debugging
```bash
go test ./tests/unit/academy_service_test.go -v -run TestAcademyServiceExample
```

### With Race Detector
```bash
go test ./tests/unit/... -race
```

### With Timeout
```bash
go test ./tests/integration/... -timeout 5m
```

## 🔍 CI/CD Integration

For GitHub Actions, add to `.github/workflows/test.yml`:

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_PASSWORD: postgres
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: make all
```

## 📚 Additional Resources

- [Testify Documentation](https://github.com/stretchr/testify)
- [Testcontainers Go](https://golang.testcontainers.org/)
- [Go httptest Package](https://golang.org/pkg/net/http/httptest/)
- [Uber Go Code Review Comments](https://github.com/uber-go/guide)

## 🎯 Next Steps

1. **Implement Mock Data**: Create fixtures and factories for consistent test data
2. **Add Benchmarks**: `tests/benchmarks/` for performance testing
3. **Setup CI/CD**: Integrate tests into GitHub Actions or other CI systems
4. **Increase Coverage**: Aim for 80%+ coverage on critical paths
5. **Integration Tests**: Expand with real database operations
