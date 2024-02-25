/*
Package sql contains a Backend implementation based on a SQL database.
*/
package sql

import (
	"database/sql"
)

// Backend contains everything needed to run a SQL backend
type Backend struct {
	DB *sql.DB
}

// NewBackend returns an instantiated Backend
func NewBackend(db *sql.DB) *Backend {
	return &Backend{
		DB: db,
	}
}
