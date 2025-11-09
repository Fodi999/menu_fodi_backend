# Testing Quick Start Guide

## 🚀 Quick Commands

```bash
# Run all tests
make all

# Run only unit tests (fastest)
make test

# Run only API tests
make api

# Run integration tests (requires Docker)
make integration

# Run tests without integration (skip Docker)
make short

# Generate coverage report
make coverage

# Clean test cache
make clean

# Show all available commands
make help
```

## 📦 Prerequisites

### Required
- Go 1.20+ (you have 1.24.3)
- `testify` package (assertion library)

### For Integration Tests Only
- Docker (running)
- PostgreSQL image (auto-downloaded)

## ⚡ First Time Setup

### 1. Install Test Dependencies
```bash
cd /Users/dmitrijfomin/Desktop/backend

# Dependencies are already in go.mod:
# - github.com/stretchr/testify v1.8.4
# - github.com/testcontainers/testcontainers-go v0.27.0

go mod download
```

### 2. Run Quick Test
```bash
# Start with just unit tests (no Docker required)
make test

# Output:
# Running unit tests...
# ok  	./tests/unit	0.234s
```

### 3. Run Integration Tests (Optional)
```bash
# Requires Docker to be running
docker ps  # Check Docker is running

make integration

# First run will:
# 1. Download postgres:15-alpine image
# 2. Start container
# 3. Run tests
# 4. Clean up container
```

## 📊 Test Structure at a Glance

```
tests/
├── unit/              (21 files) ← Fast, no Docker needed
│   ├── academy_service_test.go
│   ├── admin_service_test.go
│   ├── ...
│   └── wallet_service_test.go
│
├── integration/       (21 files) ← Requires Docker
│   ├── academy_repo_test.go
│   ├── admin_repo_test.go
│   ├── ...
│   └── wallet_repo_test.go
│
└── api/              (21 files) ← HTTP testing
    ├── academy_api_test.go
    ├── admin_api_test.go
    ├── ...
    └── wallet_api_test.go
```

## 🎯 Test Execution Times (Approximate)

| Test Suite | Time | Requirements |
|-----------|------|--------------|
| Unit Tests | ~1-2s | Go installed |
| API Tests | ~1-2s | Go installed |
| Integration Tests | ~30-60s | Docker running |
| All Tests | ~35-65s | Docker running |

## 💻 Common Scenarios

### Scenario 1: Quick Feedback During Development
```bash
# Test your service without Docker
make test

# OR with coverage
make coverage
```

### Scenario 2: Full Test Suite Before Commit
```bash
# Everything including integration tests
make all

# OR short version (skip Docker)
make short
```

### Scenario 3: Debug Specific Test
```bash
# Run one test with verbose output
go test ./tests/unit/academy_service_test.go -v

# Run with debugging
go test ./tests/unit/academy_service_test.go -v -run TestAcademyServiceExample
```

### Scenario 4: CI/CD Pipeline
```bash
# In GitHub Actions or similar
make all

# With coverage
make coverage
```

## 🔍 Troubleshooting

### Issue: "Docker daemon not running"
**Solution:**
```bash
# Start Docker Desktop or Docker daemon
docker ps

# Then run integration tests
make integration
```

### Issue: "Package testify not found"
**Solution:**
```bash
go get github.com/stretchr/testify
go mod download
```

### Issue: "Address already in use" (port 5433)
**Solution:**
```bash
# Check what's using the port
lsof -i :5433

# Kill process or use different port in .env.test
# DB_PORT=5434
```

### Issue: Tests timeout
**Solution:**
```bash
# Increase timeout for slow tests
go test ./tests/integration/... -timeout 5m
```

## 📈 Coverage Analysis

### Generate HTML Report
```bash
make coverage

# Creates: coverage.html
# Open in browser to see which lines are tested
```

### View Coverage in Terminal
```bash
go test ./tests/unit/... -cover

# Output:
# ok  	github.com/dmitrijfomin/menu-fodifood/backend/tests/unit	1.234s	coverage: 65.3% of statements
```

### Improve Coverage
```bash
# Find uncovered packages
go test ./tests/unit/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Add tests for uncovered functions
# Edit: tests/unit/<module>_service_test.go
```

## 🎓 Learning Path

1. **Start here:** `make test` (unit tests only)
2. **Then:** `make api` (HTTP testing)
3. **Finally:** `make integration` (database testing)
4. **Advanced:** `make coverage` (coverage analysis)

## 📝 Adding Your First Test

### Example: Add a test for Academy service

Edit: `tests/unit/academy_service_test.go`

```go
package unit

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

// Add this test
func TestAcademyServiceGetCourses(t *testing.T) {
	// Arrange
	expectedCourses := 5

	// Act
	// courses := service.GetCourses()

	// Assert
	assert.Equal(t, expectedCourses, 5)
}
```

### Run your test
```bash
go test ./tests/unit/academy_service_test.go -v
```

## 🚀 Next Steps

1. Replace example tests with real test cases
2. Add fixtures/factories for test data
3. Setup CI/CD pipeline (.github/workflows/)
4. Increase coverage to 80%+
5. Add performance benchmarks

## 📚 Resources

- Test configuration: `.env.test`
- Makefile commands: `Makefile`
- Full documentation: `TESTING_STRUCTURE.md`
- Test files location: `tests/`

---

**Ready to test?** Start with:
```bash
make test
```
