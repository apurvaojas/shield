package database

import (
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"log"
)

var (
	DB    *gorm.DB
	err   error
	DBErr error
)

// DBConnection create database connection
func DBConnection(masterDSN, replicaDSN string) error {
	var db = DB

	logMode := viper.GetBool("DB_LOG_MODE")
	debug := viper.GetBool("DEBUG")

	loglevel := logger.Silent
	if logMode {
		loglevel = logger.Info
	}

	db, err = gorm.Open(postgres.Open(masterDSN), &gorm.Config{
		Logger: logger.Default.LogMode(loglevel),
	})
	if !debug {
		db.Use(dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{
				postgres.Open(replicaDSN),
			},
			Policy: dbresolver.RandomPolicy{},
		}))
	}

	if err != nil {
		DBErr = err
		log.Println("Db connection error")
		return err
	}

	err = db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error

	if err != nil {
		return err
	}

	err = db.AutoMigrate(migrationModels...)

	if err != nil {
		return err
	}
	DB = db
	log.Println("############### DB CONNECTION SUCCESSFUL ###################")
	return nil
}

// GetDB connection
func GetDB() *gorm.DB {
	return DB
}

// GetDBError connection error
func GetDBError() error {
	return DBErr
}
