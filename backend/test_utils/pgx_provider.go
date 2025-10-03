package test_utils

import (
	"context"
	"embed"
	"testing"
	"time"

	"vdm/core/dependencies/database"
	"vdm/core/dependencies/env"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func CleanUpPgxProvider(c context.Context, t *testing.T, container testcontainers.Container, pgxProvider database.PgxProvider) {
	pgxProvider.Close()
	if err := container.Terminate(c); err != nil {
		t.Logf("failed to terminate container: %v", err)
	}
}

func NewTestContainerPgxProvider(c context.Context, t *testing.T) (testcontainers.Container, database.PgxProvider) {
	container, ip, port := startPostgresContainer(c, t)

	cfg := env.DatabaseConfig{
		Host:            ip,
		User:            "postgres",
		Password:        "postgres",
		Name:            "test_db",
		Port:            port.Port(),
		SSLMode:         "disable",
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: 30 * time.Second,
		MaxOpenConns:    4,
		MaxIdleConns:    2,
	}

	provider, err := database.NewPgxProvider(c, cfg)
	if err != nil {
		_ = container.Terminate(c)
		t.Fatal(err)
	}

	// Ensure DB is ready
	if err := pingWithRetry(c, provider, 8, 750*time.Millisecond); err != nil {
		CleanUpPgxProvider(c, t, container, provider)
		t.Fatal(err)
	}

	sqlDB := stdlib.OpenDBFromPool(provider.Pool())
	goose.SetBaseFS(migrationsFS)
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		CleanUpPgxProvider(c, t, container, provider)
		t.Fatal(err)
	}

	return container, provider
}

func pingWithRetry(ctx context.Context, p database.PgxProvider, retries int, delay time.Duration) error {
	pool := p.Pool()
	var err error
	for i := 0; i < retries; i++ {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = pool.Ping(pingCtx)
		cancel()
		if err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return err
}
