/*
Package sql contains a Backend implementation based on a SQL database.
*/
package sql

import (
	"database/sql"

	inventory "github.com/benrm/mtg-inventory/golang/mtg-inventory"
)

// Backend contains everything needed to run a SQL backend
type Backend struct {
	DB *sql.DB
}

var _ inventory.Backend = &Backend{}
