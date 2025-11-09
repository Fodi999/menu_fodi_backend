package integration
package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ============================================
// Test Helpers for Integration Tests
// ============================================

// StartPostgresContainer starts a PostgreSQL container for testing
func StartPostgresContainer(t *testing.T) (testcontainers.Container, string) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "test_db",
			"POSTGRES_PASSWORD": "test_password",
			"POSTGRES_USER":     "test_user",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	return container, getConnectionString(t, container)
}

// getConnectionString constructs database connection string
func getConnectionString(t *testing.T, container testcontainers.Container) string {
	t.Helper()
	ctx := context.Background()
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	return "postgres://test_user:test_password@" + host + ":" + port.Port() + "/test_db?sslmode=disable"
}

// SkipIfNoDocker skips test if Docker is not available
func SkipIfNoDocker(t *testing.T) {
	t.Helper()
	if _, err := testcontainers.DockerClient(); err != nil {
		t.Skip("Docker not available, skipping integration test")
	}
}

// SkipIntegration skips test if -short flag is used (integration tests are slow)
func SkipIntegration(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}
