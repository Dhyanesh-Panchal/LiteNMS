package db

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type ConfigDB struct {
	db *sql.DB
}

func NewConfigDB(connectionString string) (*ConfigDB, error) {

	db, err := sql.Open("postgres", connectionString)

	if err != nil {

		return nil, err

	}

	return &ConfigDB{db: db}, nil

}

func (configDB *ConfigDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return configDB.db.Query(query, args...)
}

func (configDB *ConfigDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return configDB.db.QueryRow(query, args...)
}

func (configDB *ConfigDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return configDB.db.Exec(query, args...)
}

// Close closes the database connection
func (configDB *ConfigDB) Close() error {
	return configDB.db.Close()
}
