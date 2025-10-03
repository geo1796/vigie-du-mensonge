package dependencies

import (
	"vdm/core/dependencies/database"
	"vdm/core/dependencies/env"
	"vdm/core/dependencies/mailer"

	"gorm.io/gorm"
)

type Dependencies struct {
	Config        env.Config
	gormConnector database.GormConnector
	PgxProvider   database.PgxProvider
	Mailer        mailer.Mailer
}

func (d *Dependencies) GormDB() *gorm.DB {
	return d.gormConnector.GormDB()
}

func New(cfg env.Config, dbConnector database.GormConnector, pgxProvider database.PgxProvider, mailer mailer.Mailer) *Dependencies {
	return &Dependencies{
		Config:        cfg,
		gormConnector: dbConnector,
		PgxProvider:   pgxProvider,
		Mailer:        mailer,
	}
}
