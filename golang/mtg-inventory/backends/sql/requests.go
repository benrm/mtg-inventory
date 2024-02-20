package sql

import (
	"context"
	"fmt"
	"time"

	inventory "github.com/benrm/mtg-inventory/golang/mtg-inventory"
)

// RequestCards creates a Request from the provided rows of RequestedCards
func (b *Backend) RequestCards(ctx context.Context, requestorUsername string, rows []*inventory.RequestedCards) (_ *inventory.Request, err error) {
	tx, err := b.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error requesting cards: %w", err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("error requesting cards: %w, unable to rollback: %s", err, rollbackErr)
			} else {
				err = fmt.Errorf("error requesting cards: %w", err)
			}
		}
	}()

	requestor, err := getUserByUsername(ctx, tx, requestorUsername)
	if err != nil {
		return nil, fmt.Errorf("error getting requestor id: %w", err)
	}

	insertRequestStmt, err := tx.PrepareContext(ctx, `INSERT INTO requests (requestor, opened)
VALUES (?, ?)
`)
	if err != nil {
		return nil, fmt.Errorf("error preparing request: %w", err)
	}

	now := time.Now()

	result, err := insertRequestStmt.ExecContext(ctx, requestor.ID, now)
	if err != nil {
		return nil, fmt.Errorf("error inserting request: %w", err)
	}

	requestID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting new request ID: %w", err)
	}

	upsertRequestedCardsStmt, err := tx.PrepareContext(ctx, `INSERT INTO requested_cards (request_id, name, oracle_id, quantity)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE quantity = quantity + ?
`)
	if err != nil {
		return nil, fmt.Errorf("error preparing request cards: %w", err)
	}

	for _, cards := range rows {
		_, err := upsertRequestedCardsStmt.ExecContext(ctx, requestID, cards.Name, cards.OracleID, cards.Quantity, cards.Quantity)
		if err != nil {
			return nil, fmt.Errorf("error inserting requested cards: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("error commiting cards request: %w", err)
	}

	request := &inventory.Request{
		ID:        requestID,
		Requestor: requestorUsername,
		Opened:    now,
		Cards:     rows,
	}

	return request, nil
}
