package sql

import (
	"context"
	"database/sql"
	"errors"
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

	if limit == 0 {
		limit = inventory.DefaultListLimit
	} else if limit > inventory.MaxListLimit {
		limit = inventory.MaxListLimit
	}

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
	defer selectStmt.Close()

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

// GetRequestByID returns a request and associated requested cards given an ID
func (b *Backend) GetRequestByID(ctx context.Context, id int64, limit, offset int) (_ *inventory.Request, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting request \"%d\": %w", id, err)
		}
	}()

	if limit == 0 {
		limit = inventory.DefaultListLimit
	} else if limit > inventory.MaxListLimit {
		limit = inventory.MaxListLimit
	}

	selectRequestStmt, err := b.DB.PrepareContext(ctx, `SELECT users.username, requests.opened, requests.closed
FROM requests
LEFT JOIN users ON requests.requestor = users.id
WHERE requests.id = ?
`)
	if err != nil {
		return nil, fmt.Errorf("error preparing select for request: %w", err)
	}
	defer selectRequestStmt.Close()

	row := selectRequestStmt.QueryRowContext(ctx, id)
	var requestor string
	var opened time.Time
	var closed sql.NullTime
	err = row.Scan(&requestor, &opened, &closed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, inventory.ErrRequestNoExist
		}
		return nil, fmt.Errorf("error scanning row for request: %w", err)
	}

	request := &inventory.Request{
		ID:        id,
		Requestor: requestor,
		Opened:    opened,
		Cards:     make([]*inventory.RequestedCards, 0),
	}
	if closed.Valid {
		request.Closed = &closed.Time
	}

	selectCardsStmt, err := b.DB.PrepareContext(ctx, `SELECT quantity, name, oracle_id
FROM requested_cards
WHERE request_id = ?
ORDER BY name
LIMIT ?
OFFSET ?
`)
	if err != nil {
		return nil, fmt.Errorf("error preparing select for cards: %w", err)
	}
	defer selectCardsStmt.Close()

	rows, err := selectCardsStmt.QueryContext(ctx, id, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error executing select for cards: %w", err)
	}
	for rows.Next() {
		var quantity int
		var name, oracleID string
		err = rows.Scan(&id, &name, &oracleID)
		if err != nil {
			return nil, fmt.Errorf("error scanning row for cards: %w", err)
		}

		cards := &inventory.RequestedCards{
			Quantity: quantity,
			Name:     name,
			OracleID: oracleID,
		}
		request.Quantity += quantity
		request.Cards = append(request.Cards, cards)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error getting next row of cards: %w", err)
	}

	return request, nil
}

// RequestCards creates a Request from the provided rows of RequestedCards
func (b *Backend) RequestCards(ctx context.Context, requestorUsername string, rows []*inventory.RequestedCards) (_ *inventory.Request, err error) {
	if len(rows) > inventory.RowUploadLimit {
		return nil, inventory.ErrTooManyRows
	}
	for _, row := range rows {
		if row.Quantity <= 0 {
			return nil, &inventory.RowError{
				Err: inventory.ErrZeroOrFewerCards,
				Row: row,
			}
		}
	}

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

// CloseRequest sets the closed time on a Request
func (b *Backend) CloseRequest(ctx context.Context, id int64) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error closing request \"%d\": %w", id, err)
		}
	}()

	closeStmt, err := b.DB.PrepareContext(ctx, `UPDATE requests
SET closed = NOW()
WHERE id = ?
`)
	if err != nil {
		return fmt.Errorf("error preparing update to close request: %w", err)
	}
	defer closeStmt.Close()

	result, err := closeStmt.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("error updating request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting result: %w", err)
	}

	if rowsAffected <= 0 {
		return inventory.ErrRequestNoExist
	}

	return nil
}
