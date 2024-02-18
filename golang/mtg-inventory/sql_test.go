package inventory

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestSQL(t *testing.T) {
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not set, skipping SQL tests")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("Failed to open db connection: %s", err.Error())
	}

	var user *User
	ret := t.Run("TestSQLAddUser", func(t *testing.T) {
		user, err = AddUser(context.Background(), db, &User{
			Username: "benrm",
			Email:    "benrm@github.com",
		})
		if err != nil {
			t.Fatalf("Failed to add user: %s", err.Error())
		}
	})
	if !ret {
		t.FailNow()
	}

	ret = t.Run("TestSQLAddCards", func(t *testing.T) {
		err := AddCards(context.Background(), db, []*CardRow{
			{
				Quantity: 1,
				Card: &Card{
					OracleID:   "fake-oracle-ID",
					ScryfallID: "fake-scryfall-ID",
					Foil:       false,
				},
				Owner:  user,
				Keeper: user,
			},
		})
		if err != nil {
			t.Fatalf("Failed to insert cards: %s", err.Error())
		}

		err = AddCards(context.Background(), db, []*CardRow{
			{
				Quantity: 1,
				Card: &Card{
					OracleID:   "fake-oracle-ID",
					ScryfallID: "fake-scryfall-ID",
					Foil:       false,
				},
				Owner:  user,
				Keeper: user,
			},
		})
		if err != nil {
			t.Fatalf("Failed to update cards: %s", err.Error())
		}
	})
	if !ret {
		t.FailNow()
	}
}
