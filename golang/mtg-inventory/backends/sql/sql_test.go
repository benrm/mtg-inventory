package sql

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"testing"

	inventory "github.com/benrm/mtg-inventory/golang/mtg-inventory"
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

	b := &Backend{
		DB: db,
	}

	user1, err := b.AddUser(context.Background(), "user1", "user1@domain.com")
	if err != nil {
		t.Fatalf("Failed to add user: %s", err.Error())
	}

	_, err = b.GetUserByUsername(context.Background(), user1.Username)
	if err != nil {
		t.Fatalf("Failed to get user by username: %s", err.Error())
	}

	fakeCard := &inventory.Card{
		Name:       "fake-card-name",
		OracleID:   "fake-oracle-ID",
		ScryfallID: "fake-scryfall-ID",
		Foil:       false,
	}

	fakeCardRow := &inventory.CardRow{
		Quantity: 1,
		Card:     fakeCard,
		Owner:    user1,
		Keeper:   user1,
	}

	err = b.AddCards(context.Background(), []*inventory.CardRow{
		fakeCardRow,
	})
	if err != nil {
		t.Fatalf("Failed to insert cards: %s", err.Error())
	}

	err = b.AddCards(context.Background(), []*inventory.CardRow{
		fakeCardRow,
	})
	if err != nil {
		t.Fatalf("Failed to update cards: %s", err.Error())
	}

	_, err = b.GetCardsByOracleID(context.Background(), fakeCard.OracleID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get cards by oracle ID: %s", err.Error())
	}

	_, err = b.GetCardsByOwner(context.Background(), user1, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get cards by owner: %s", err.Error())
	}

	_, err = b.GetCardsByKeeper(context.Background(), user1, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get cards by keeper: %s", err.Error())
	}

	user2, err := b.AddUser(context.Background(), "user2", "user2@domain.com")
	if err != nil {
		t.Fatalf("Failed to add user: %s", err.Error())
	}

	fakeTransferRow := &inventory.TransferredCards{
		Quantity: 1,
		Card:     fakeCard,
		Owner:    user1,
	}

	_, err = b.TransferCards(context.Background(), user2, user1, nil, []*inventory.TransferredCards{
		fakeTransferRow,
	})
	if err != nil {
		t.Fatalf("Failed to transfer cards: %s", err.Error())
	}
}
