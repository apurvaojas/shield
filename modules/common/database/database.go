package database

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewConnectionFromViper creates a DB connection using viper config
func NewConnection() (*gorm.DB, error) {
	// Validate required configuration
	host := viper.GetString("database.host")
	if host == "" {
		return nil, fmt.Errorf("database.host is required")
	}

	port := viper.GetInt("database.port")
	if port == 0 {
		port = 5432 // Default PostgreSQL port
	}

	user := viper.GetString("database.user")
	if user == "" {
		return nil, fmt.Errorf("database.user is required")
	}

	password := viper.GetString("database.password")
	// if password == "" {
	// 	return nil, fmt.Errorf("database.password is required")
	// }

	dbname := viper.GetString("database.name")
	if dbname == "" {
		return nil, fmt.Errorf("database.name is required")
	}

	sslmode := viper.GetString("database.sslMode")
	if sslmode == "" {
		sslmode = "require" // Default to require SSL
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		host, user, password, dbname, port, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool from YAML config
	maxIdleConns := viper.GetInt("database.maxIdleConns")
	if maxIdleConns == 0 {
		maxIdleConns = 10 // Default
	}

	maxOpenConns := viper.GetInt("database.maxOpenConns")
	if maxOpenConns == 0 {
		maxOpenConns = 100 // Default
	}

	connMaxLifetime := viper.GetDuration("database.connMaxLifetime")
	if connMaxLifetime == 0 {
		connMaxLifetime = time.Hour // Default
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	return db, nil
}
