# Real Test Examples

This file contains examples of how to write real tests for your modules.

## Unit Test Example

```go
// tests/unit/academy_service_test.go
package unit

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockAcademyRepo struct {
	courses []Course
}

func (m *MockAcademyRepo) GetCourses() ([]Course, error) {
	return m.courses, nil
}

// Test using table-driven approach
func TestAcademyServiceGetCourses(t *testing.T) {
	tests := []struct {
		name      string
		courses   []Course
		wantError bool
	}{
		{
			name:      "should return courses",
			courses:   []Course{{ID: "1", Title: "Go Basics"}},
			wantError: false,
		},
		{
			name:      "should handle empty courses",
			courses:   []Course{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockAcademyRepo{courses: tt.courses}
			
			// Act
			result := repo.courses
			
			// Assert
			if tt.wantError {
				assert.Empty(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}
```

## Integration Test Example

```go
// tests/integration/academy_repo_test.go
package integration

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

func TestAcademyRepoWithDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "test_db",
			"POSTGRES_PASSWORD": "test_pass",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	// Get connection string from container
	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432/tcp")
	connStr := "postgres://postgres:test_pass@" + host + ":" + port.Port() + "/test_db"

	// Connect to database
	db, err := gorm.Open(postgres.Open(connStr))
	require.NoError(t, err)

	// Run migrations (example)
	// db.AutoMigrate(&Course{})

	// Test repository operations
	// repo := academy.NewRepository(db)
	// courses, err := repo.GetAll(ctx)
	// assert.NoError(t, err)
	// assert.NotNil(t, courses)
}
```

## API Test Example

```go
// tests/api/academy_api_test.go
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAcademyGetCourses(t *testing.T) {
	// Setup test server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"1","title":"Go Basics"}]`))
	})

	// Create test request
	req, err := http.NewRequest("GET", "/api/academy/courses", nil)
	assert.NoError(t, err)

	// Record response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "Go Basics")
}

func TestAcademyUpdateCourse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"1","title":"Updated Course"}`))
	})

	req, _ := http.NewRequest("PUT", "/api/academy/courses/1", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Updated Course")
}
```

## Best Practices

### ✅ DO

- Use table-driven tests
- Name tests clearly: `TestFunctionNameScenario`
- Use subtests with `t.Run()`
- Mock external dependencies
- Test error cases
- Use `require` for fatal checks
- Use `assert` for non-fatal checks

### ❌ DON'T

- Don't write one giant test function
- Don't ignore error cases
- Don't test implementation details
- Don't create test files without the `_test.go` suffix
- Don't use `t.Fatal()` when `require` is available
- Don't create external dependencies in unit tests

## Running Specific Tests

```bash
# Run single test file
go test ./tests/unit/academy_service_test.go -v

# Run single test function
go test ./tests/unit/academy_service_test.go -v -run TestAcademyServiceGetCourses

# Run with verbose output
go test ./tests/unit/... -v

# Run with coverage
go test ./tests/unit/... -cover

# Run with timeout
go test ./tests/integration/... -timeout 5m
```

## Coverage Goals

| Layer | Target | How to Improve |
|-------|--------|----------------|
| Unit | 80%+ | Add tests for all service methods |
| Integration | 60%+ | Test database queries and transactions |
| API | 75%+ | Test all HTTP methods and status codes |

## Next Steps

1. Replace placeholder tests with real test cases
2. Add fixtures and factories for test data
3. Setup CI/CD to run tests on every commit
4. Monitor coverage with tools like codecov
5. Add performance benchmarks

---

**Questions?** Check:
- `TESTING_STRUCTURE.md` - Complete guide
- `TESTING_QUICK_START.md` - Quick reference
- Testify docs: https://github.com/stretchr/testify
- Testcontainers: https://golang.testcontainers.org/
