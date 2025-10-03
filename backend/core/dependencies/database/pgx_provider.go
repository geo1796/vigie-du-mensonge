package database

import (
	"context"
	"vdm/core/dependencies/database/queries"
	"vdm/core/dependencies/env"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxProvider interface {
	Pool() *pgxpool.Pool
	Queries() *queries.Queries
	Close()
}

type pgxProvider struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

func (p *pgxProvider) Pool() *pgxpool.Pool {
	return p.pool
}

func (p *pgxProvider) Queries() *queries.Queries {
	return p.queries
}

func (p *pgxProvider) Close() {
	p.pool.Close()
}

func NewPgxProvider(c context.Context, cfg env.DatabaseConfig) (PgxProvider, error) {
	pool, err := pgxpool.New(c, cfg.DSN())
	if err != nil {
		return nil, err
	}

	return &pgxProvider{
		pool:    pool,
		queries: queries.New(pool),
	}, nil
}
