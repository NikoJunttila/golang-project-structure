package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

const (
	DriverSqlite3 = "sqlite"
	DriverMysql   = "mysql"
)

type Config struct {
	Driver   string
	Name     string
	Host     string
	User     string
	Password string
}

var dbInstance *Queries

// InitDefault initializes the DB with hardcoded defaults (used in main for now)
func InitDefault() {
	Init(Config{
		Driver: DriverSqlite3,
		Name:   "app.db",
	})
}

// Init sets up the database connection using a custom config
func Init(cfg Config) {
	connection, err := sql.Open(cfg.Driver, cfg.Name)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = New(connection)
}

// Get returns the instantiated DB instance.
func Get() *Queries {
	return dbInstance
}
