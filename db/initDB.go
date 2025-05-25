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

type config struct {
	Driver   string
	Name     string
	Host     string
	User     string
	Password string
}

var dbInstance *Queries

// Get returns the instantiated DB instance.
func Get() *Queries {
	return dbInstance
}

func init() {
	config := config{
		Driver: DriverSqlite3,
		Name:   "app.db",
		// // Password: os.Getenv("DB_PASSWORD"),
		// User:     os.Getenv("DB_USER"),
		// Host:     os.Getenv("DB_HOST"),
	}

	connection, err := sql.Open(config.Driver, config.Name)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = New(connection)
}
