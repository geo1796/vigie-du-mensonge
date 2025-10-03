package test_utils

import (
	"context"
	"fmt"
	"testing"
	"time"
	"vdm/core/dependencies/database"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func CleanUpTestData(c context.Context, t *testing.T, container testcontainers.Container, connector database.GormConnector) {
	if err := connector.Close(); err != nil {
		t.Logf("failed to close connector: %v", err)
	}
	if err := container.Terminate(c); err != nil {
		t.Logf("failed to terminate container: %v", err)
	}
}

func NewTestContainerConnector(c context.Context, t *testing.T) (testcontainers.Container, database.GormConnector) {
	container, ip, port := startPostgresContainer(c, t)

	connector, err := newPostgresTestContainerConnector(ip, port.Port(), "disable")
	if err != nil {
		if termErr := container.Terminate(c); termErr != nil {
			t.Logf("failed to terminate postgres container after connector init error: %v", termErr)
		}
		t.Fatal(err)
	}

	if err := connector.Migrate(); err != nil {
		// Close connector before terminating container
		if closeErr := connector.Close(); closeErr != nil {
			t.Logf("failed to close connector after migrate error: %v", closeErr)
		}
		if termErr := container.Terminate(c); termErr != nil {
			t.Logf("failed to terminate postgres container after migrate error: %v", termErr)
		}
		t.Fatal(err)
	}

	return container, connector
}

func startPostgresContainer(c context.Context, t *testing.T) (testcontainers.Container, string, nat.Port) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17-alpine",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "test_db",
		},
	}

	postgresC, err := testcontainers.GenericContainer(c, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %s", err)
	}

	ip, err := postgresC.Host(c)
	if err != nil {
		if termErr := postgresC.Terminate(c); termErr != nil {
			t.Logf("failed to terminate postgres container after host error: %v", termErr)
		}
		t.Fatalf("failed to get container host: %s", err)
	}

	port, err := postgresC.MappedPort(c, "5432")
	if err != nil {
		if termErr := postgresC.Terminate(c); termErr != nil {
			t.Logf("failed to terminate postgres container after mapped port error: %v", termErr)
		}
		t.Fatalf("failed to get container port: %s", err)
	}

	return postgresC, ip, port
}

func newPostgresTestContainerConnector(host, port, sslMode string) (database.GormConnector, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, "postgres", "postgres", "test_db", port, sslMode)

	var db *gorm.DB
	var err error

	// Retry logic
	maxRetries := 5
	for range maxRetries {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}

		// Wait before retrying
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return &database.PgGormConnector{DB: db}, nil
}
