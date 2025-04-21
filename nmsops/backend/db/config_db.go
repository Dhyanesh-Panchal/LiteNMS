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

func (mdb *ConfigDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return mdb.db.Query(query, args...)
}

func (mdb *ConfigDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return mdb.db.QueryRow(query, args...)
}

func (mdb *ConfigDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return mdb.db.Exec(query, args...)
}

// Close closes the database connection
func (mdb *ConfigDB) Close() error {
	return mdb.db.Close()
}
