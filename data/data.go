package data

import (
	"code.google.com/p/go-uuid/uuid"
	"database/sql"
	_ "github.com/denisenkom/go-mssqldb"
	"log"
	"time"
)

type Repository interface {
	GetDb() *sql.DB
}

type Database struct {
	db               *sql.DB
	driverName       string
	connectionString string
}

func NewDatabase(driverName string, connectionString string) *Database {
	if driverName == "" || connectionString == "" {
		log.Panicln("Both driver name and connection string are required")
	}
	db, err := sql.Open(driverName, connectionString)
	defer db.Close()
	if err != nil {
		log.Panicf("%s\n", err.Error())
	}
	return &Database{
		db: db,
	}
}

func (d *Database) Open() (*sql.DB, error) {
	return sql.Open(d.driverName, d.connectionString)
}

func (d *Database) GetDriverName() string {
	return d.driverName
}

func (d *Database) GetConnectionString() string {
	return d.connectionString
}

func (d *Database) Close() {
	d.db.Close()
}
