package env

import (
	"fmt"
	"strconv"
	"time"
)

type DatabaseConfig struct {
	Host            string
	User            string
	Password        string
	Name            string
	Port            string
	SSLMode         string
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
}

func (cfg DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode)
}

func loadDatabaseConfig() (DatabaseConfig, error) {
	dbConnMaxLifetime, err := time.ParseDuration(getEnv("DB_CONN_MAX_LIFETIME", "10m"))
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to parse DB_CONN_MAX_LIFETIME: %v", err)
	}

	dbConnMaxIdleTime, err := time.ParseDuration(getEnv("DB_CONN_MAX_IDLE_TIME", "5m"))
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to parse DB_CONN_MAX_IDLE_TIME: %v", err)
	}

	dbMaxOpenConns, err := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "2"))
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to parse DB_MAX_OPEN_CONNS: %v", err)
	}

	dbMaxIdleConns, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "1"))
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to parse DB_MAX_IDLE_CONNS: %v", err)
	}

	return DatabaseConfig{
		Host:            getEnv("DB_HOST", ""),
		User:            getEnv("DB_USER", ""),
		Password:        getEnv("DB_PASSWORD", ""),
		Name:            getEnv("DB_NAME", ""),
		Port:            getEnv("DB_PORT", ""),
		SSLMode:         getEnv("DB_SSL_MODE", ""),
		ConnMaxLifetime: dbConnMaxLifetime,
		ConnMaxIdleTime: dbConnMaxIdleTime,
		MaxOpenConns:    dbMaxOpenConns,
		MaxIdleConns:    dbMaxIdleConns,
	}, nil
}
