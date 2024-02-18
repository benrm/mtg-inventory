package sql

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestSQL(t *testing.T) {
	var err error
	var deleteAll bool
	deleteAllStr := os.Getenv("TEST_DELETE_ALL_ROWS")
	if deleteAllStr != "" {
		deleteAll, err = strconv.ParseBool(deleteAllStr)
		if err != nil {
			t.Fatalf("Failed to parse TEST_DELETE_ALL_ROWS: %s", err.Error())
		}
	}

	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not set, skipping SQL tests")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("Failed to open db connection: %s", err.Error())
	}

	if deleteAll {
		_, err = db.Exec("DELETE FROM transferred_cards")
		if err != nil {
			t.Fatalf("Failed to delete from transferred_cards")
		}
		_, err = db.Exec("DELETE FROM transfers")
		if err != nil {
			t.Fatalf("Failed to delete from transfers")
		}
		_, err = db.Exec("DELETE FROM requested_cards")
		if err != nil {
			t.Fatalf("Failed to delete from requested_cards")
		}
		_, err = db.Exec("DELETE FROM requests")
		if err != nil {
			t.Fatalf("Failed to delete from requests")
		}
		_, err = db.Exec("DELETE FROM cards")
		if err != nil {
			t.Fatalf("Failed to delete from cards")
		}
		_, err = db.Exec("DELETE FROM users")
		if err != nil {
			t.Fatalf("Failed to delete from users")
		}
	}

	var user *User
	ret := t.Run("TestSQLAddUser", func(t *testing.T) {
		user, err = AddUser(context.Background(), db, "user", "user@domain.com")
		if err != nil {
			t.Fatalf("Failed to add user: %s", err.Error())
		}
	})
	if !ret {
		t.FailNow()
	}

	ret = t.Run("TestSQLGetUserByID", func(t *testing.T) {
		_, err = GetUserByID(context.Background(), db, user.ID)
		if err != nil {
			t.Fatalf("Failed to get user by ID: %s", err.Error())
		}
	})
	if !ret {
		t.FailNow()
	}

	ret = t.Run("TestSQLGetUserByUsername", func(t *testing.T) {
		_, err = GetUserByUsername(context.Background(), db, user.Username)
		if err != nil {
			t.Fatalf("Failed to get user by username: %s", err.Error())
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
