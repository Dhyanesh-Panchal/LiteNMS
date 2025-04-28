package db

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type ConfigDBClient struct {
	db *sql.DB
}

func InitConfigDBClient(connectionString string) (*ConfigDBClient, error) {

	db, err := sql.Open("postgres", connectionString)

	if err != nil {

		return nil, err

	}

	return &ConfigDBClient{db: db}, nil

}

func (configDB *ConfigDBClient) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return configDB.db.Query(query, args...)
}

func (configDB *ConfigDBClient) QueryRow(query string, args ...interface{}) *sql.Row {
	return configDB.db.QueryRow(query, args...)
}

func (configDB *ConfigDBClient) Exec(query string, args ...interface{}) (sql.Result, error) {
	return configDB.db.Exec(query, args...)
}

// Close closes the database connection
func (configDB *ConfigDBClient) Close() error {
	return configDB.db.Close()
}
