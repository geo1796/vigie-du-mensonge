package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vdm/api"
	"vdm/core/dependencies"
	"vdm/core/dependencies/database"
	"vdm/core/dependencies/env"
	"vdm/core/dependencies/mailer"
	"vdm/core/fiberx"
	"vdm/core/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	cfg, err := env.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", logger.Err(err))
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid config", logger.Err(err))
		os.Exit(1)
	}

	dbConn, err := database.NewGormConnector(cfg.Database)
	if err != nil {
		logger.Error("failed to init database", logger.Err(err))
		os.Exit(1)
	}

	defer func(dbConn database.GormConnector) {
		if err := dbConn.Close(); err != nil {
			logger.Error("failed to close database connection", logger.Err(err))
		}
	}(dbConn)

	c := context.Background()
	pgxProvider, err := database.NewPgxProvider(c, cfg.Database)
	if err != nil {
		logger.Error("failed to init pgx provider", logger.Err(err))
		os.Exit(1)
	}

	defer func(pgxProvider database.PgxProvider) {
		pgxProvider.Close()
	}(pgxProvider)

	deps := dependencies.New(cfg, dbConn, pgxProvider, mailer.New(cfg.Mailer))

	app := fiberx.NewApp()
	app.Use(recover.New())
	app.Use(requestid.New())

	app.Get("/livez", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	app.Get("/healthz", func(c *fiber.Ctx) error {
		sqlDB, err := deps.GormDB().DB()
		if err != nil {
			return c.SendStatus(fiber.StatusServiceUnavailable)
		}
		ctx, cancel := context.WithTimeout(c.Context(), 300*time.Millisecond)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			return c.SendStatus(fiber.StatusServiceUnavailable)
		}
		return c.SendStatus(fiber.StatusOK)
	})

	if cfg.ActiveProfile == "docker" {
		app.Get("/docs", func(c *fiber.Ctx) error {
			return c.SendFile("/app/docs/docs.html")
		})
	}

	api.Register(app, deps)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		if err := app.Shutdown(); err != nil {
			logger.Error("error shutting down server", logger.Err(err))
		}
	}()

	if err := app.Listen("0.0.0.0:8080"); err != nil {
		logger.Error("unexpected error, server about to shut down", logger.Err(err))
	}
}
