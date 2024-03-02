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
			t.Fatalf("Failed to delete from transferred_cards: %s", err.Error())
		}
		_, err = db.Exec("DELETE FROM transfers")
		if err != nil {
			t.Fatalf("Failed to delete from transfers: %s", err.Error())
		}
		_, err = db.Exec("DELETE FROM requested_cards")
		if err != nil {
			t.Fatalf("Failed to delete from requested_cards: %s", err.Error())
		}
		_, err = db.Exec("DELETE FROM requests")
		if err != nil {
			t.Fatalf("Failed to delete from requests: %s", err.Error())
		}
		_, err = db.Exec("DELETE FROM cards")
		if err != nil {
			t.Fatalf("Failed to delete from cards: %s", err.Error())
		}
		_, err = db.Exec("DELETE FROM users")
		if err != nil {
			t.Fatalf("Failed to delete from users: %s", err.Error())
		}
	}

	b := NewBackend(db)

	user1, err := b.AddUserIfNotExist(context.Background(), "user1")
	if err != nil {
		t.Fatalf("Failed to add user: %s", err.Error())
	}

	fakeCard1 := &inventory.Card{
		Name:       "fake-card-name-1",
		OracleID:   "fake-oracle-ID-1",
		ScryfallID: "fake-scryfall-ID-1",
		Foil:       false,
	}

	fakeCardRow1 := &inventory.CardRow{
		Quantity: 1,
		Card:     fakeCard1,
		Owner:    user1.SlackID,
		Keeper:   user1.SlackID,
	}

	fakeCard2 := &inventory.Card{
		Name:       "fake-card-name-2",
		OracleID:   "fake-oracle-ID-2",
		ScryfallID: "fake-scryfall-ID-2",
		Foil:       false,
	}

	fakeCardRow2 := &inventory.CardRow{
		Quantity: 1,
		Card:     fakeCard2,
		Owner:    user1.SlackID,
		Keeper:   user1.SlackID,
	}

	err = b.AddCards(context.Background(), []*inventory.CardRow{
		fakeCardRow1,
		fakeCardRow2,
	})
	if err != nil {
		t.Fatalf("Failed to insert cards: %s", err.Error())
	}

	err = b.AddCards(context.Background(), []*inventory.CardRow{
		fakeCardRow1,
	})
	if err != nil {
		t.Fatalf("Failed to update cards: %s", err.Error())
	}

	err = b.ModifyCardQuantity(context.Background(), user1.SlackID, user1.SlackID, fakeCard1.ScryfallID, false, 7)
	if err != nil {
		t.Fatalf("Failed to update card quantity: %s", err.Error())
	}

	err = b.ModifyCardQuantity(context.Background(), user1.SlackID, user1.SlackID, fakeCard2.ScryfallID, false, 0)
	if err != nil {
		t.Fatalf("Failed to update card quantity: %s", err.Error())
	}

	_, err = b.GetCardsByOracleID(context.Background(), fakeCard1.OracleID, inventory.DefaultListLimit, 0)
	if err != nil {
		t.Fatalf("Failed to get cards by oracle ID: %s", err.Error())
	}

	_, err = b.GetCardsByOwner(context.Background(), user1.SlackID, inventory.DefaultListLimit, 0)
	if err != nil {
		t.Fatalf("Failed to get cards by owner: %s", err.Error())
	}

	_, err = b.GetCardsByKeeper(context.Background(), user1.SlackID, inventory.DefaultListLimit, 0)
	if err != nil {
		t.Fatalf("Failed to get cards by keeper: %s", err.Error())
	}

	user2, err := b.AddUserIfNotExist(context.Background(), "user2")
	if err != nil {
		t.Fatalf("Failed to add user: %s", err.Error())
	}

	request, err := b.OpenRequest(context.Background(), user1.SlackID, []*inventory.RequestedCards{
		{
			Name:     "fake-card-name-2",
			OracleID: "fake-oracle-ID-2",
			Quantity: 7,
		},
	})
	if err != nil {
		t.Fatalf("Failed to request cards: %s", err.Error())
	}

	_, err = b.GetRequestsByRequestor(context.Background(), user1.SlackID, inventory.DefaultListLimit, 0)
	if err != nil {
		t.Fatalf("Failed to get requests: %s", err.Error())
	}

	_, err = b.GetRequestByID(context.Background(), request.ID, inventory.DefaultListLimit, 0)
	if err != nil {
		t.Fatalf("Failed to get request by ID: %s", err.Error())
	}

	err = b.CloseRequest(context.Background(), request.ID)
	if err != nil {
		t.Fatalf("Failed to close request: %s", err.Error())
	}

	fakeTransferRow := &inventory.TransferredCards{
		Quantity: 1,
		Card:     fakeCard1,
		Owner:    user1.SlackID,
	}

	transfer, err := b.OpenTransfer(context.Background(), user2.SlackID, user1.SlackID, &request.ID, []*inventory.TransferredCards{
		fakeTransferRow,
	})
	if err != nil {
		t.Fatalf("Failed to transfer cards: %s", err.Error())
	}

	_, err = b.GetTransfersByToUser(context.Background(), user2.SlackID, inventory.DefaultListLimit, 0)
	if err != nil {
		t.Fatalf("Failed to get transfer by to user: %s", err.Error())
	}

	_, err = b.GetTransfersByFromUser(context.Background(), user1.SlackID, inventory.DefaultListLimit, 0)
	if err != nil {
		t.Fatalf("Failed to get transfer by from user: %s", err.Error())
	}

	_, err = b.GetTransfersByRequestID(context.Background(), request.ID, inventory.DefaultListLimit, 0)
	if err != nil {
		t.Fatalf("Failed to get transfer by request ID: %s", err.Error())
	}

	_, err = b.GetTransferByID(context.Background(), transfer.ID, inventory.DefaultListLimit, 0)
	if err != nil {
		t.Fatalf("Failed to get transfer by ID: %s", err.Error())
	}

	err = b.CancelTransfer(context.Background(), transfer.ID)
	if err != nil {
		t.Fatalf("Failed to close transfer: %s", err.Error())
	}
}
