/*
Package sql contains a Backend implementation based on a SQL database.
*/
package sql

import (
	"context"
	"database/sql"

	inventory "github.com/benrm/mtg-inventory/golang/mtg-inventory"
)

type fullUser struct {
	ID   int64
	User *inventory.User
}

type stmtPreparer interface {
	PrepareContext(context.Context, string) (*sql.Stmt, error)
}

// Backend contains everything needed to run a SQL backend
type Backend struct {
	DB *sql.DB
}

// NewBackend returns an instantiated Backend
func NewBackend(db *sql.DB) inventory.Backend {
	return &Backend{
		DB: db,
	}
}
