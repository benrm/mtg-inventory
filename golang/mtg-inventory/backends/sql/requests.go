package sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	inventory "github.com/benrm/mtg-inventory/golang/mtg-inventory"
)

// GetRequestsByRequestor gets a number of requests made by the requestor up to
// the provided limit
func (b *Backend) GetRequestsByRequestor(ctx context.Context, requestorUsername string, limit, offset int) (_ []*inventory.Request, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting requests from %q: %w", requestorUsername, err)
		}
	}()

	selectStmt, err := b.DB.PrepareContext(ctx, `SELECT requests.id, requests.opened, requests.closed, SUM(rc.quantity)
FROM requests
LEFT JOIN requested_cards rc ON requests.id = rc.request_id
LEFT JOIN users ON requests.requestor = users.id
WHERE users.username = ?
GROUP BY requests.id
ORDER BY requests.opened
LIMIT ?
OFFSET ?
`)
	if err != nil {
		return nil, fmt.Errorf("error preparing select query: %w", err)
	}

	rows, err := selectStmt.QueryContext(ctx, requestorUsername, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error executing select query: %w", err)
	}

	requests := make([]*inventory.Request, 0)
	for rows.Next() {
		var id int64
		var opened time.Time
		var closed sql.NullTime
		var quantity int
		err = rows.Scan(&id, &opened, &closed, &quantity)
		if err != nil {
			return nil, fmt.Errorf("error scanning row of select: %w", err)
		}
		request := &inventory.Request{
			ID:        id,
			Requestor: requestorUsername,
			Opened:    opened,
			Quantity:  quantity,
		}
		if closed.Valid {
			request.Closed = &closed.Time
		}
		requests = append(requests, request)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error getting next row of select: %w", err)
	}

	return requests, nil
}

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

	insertRequestStmt, err := tx.PrepareContext(ctx, `INSERT INTO requests (requestor, opened)
SELECT users.id, ?
FROM users
WHERE users.username = ?
`)
	if err != nil {
		return nil, fmt.Errorf("error preparing request: %w", err)
	}

	now := time.Now()

	result, err := insertRequestStmt.ExecContext(ctx, now, requestorUsername)
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
