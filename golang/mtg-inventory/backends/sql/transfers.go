package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	inventory "github.com/benrm/mtg-inventory/golang/mtg-inventory"
)

// GetTransfersByToUser returns Transfers based on their ToUser
func (b *Backend) GetTransfersByToUser(ctx context.Context, toUser string, limit, offset uint) (_ []*inventory.Transfer, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting transfers to %q: %w", toUser, err)
		}
	}()

	if limit == 0 {
		limit = inventory.DefaultListLimit
	} else if limit > inventory.MaxListLimit {
		limit = inventory.MaxListLimit
	}

	selectStmt, err := b.DB.PrepareContext(ctx, `SELECT transfers.id, transfers.request_id, from_users.username, transfers.opened, transfers.closed, SUM(tc.quantity)
FROM transfers
LEFT JOIN users from_users ON transfers.from_user = from_users.id
LEFT JOIN users to_users ON transfers.to_user = to_users.id
LEFT JOIN transferred_cards tc ON tc.transfer_id = transfers.id
WHERE to_users.username = ?
GROUP BY transfers.id
ORDER BY transfers.opened
LIMIT ?
OFFSET ?
`)
	if err != nil {
		return nil, fmt.Errorf("error preparing select query: %w", err)
	}
	defer selectStmt.Close()

	rows, err := selectStmt.QueryContext(ctx, toUser, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error executing select query: %w", err)
	}

	transfers := make([]*inventory.Transfer, 0)
	for rows.Next() {
		var id int64
		var requestID sql.NullInt64
		var fromUser string
		var opened time.Time
		var closed sql.NullTime
		var quantity uint
		err = rows.Scan(&id, &requestID, &fromUser, &opened, &closed, &quantity)
		if err != nil {
			return nil, fmt.Errorf("error scanning row of select: %w", err)
		}
		transfer := &inventory.Transfer{
			ID:       id,
			ToUser:   toUser,
			FromUser: fromUser,
			Opened:   opened,
			Quantity: quantity,
		}
		if requestID.Valid {
			transfer.RequestID = &requestID.Int64
		}
		if closed.Valid {
			transfer.Closed = &closed.Time
		}
		transfers = append(transfers, transfer)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error getting next row of select: %w", err)
	}

	return transfers, nil
}

// GetTransfersByFromUser returns Transfers based on their FromUser
func (b *Backend) GetTransfersByFromUser(ctx context.Context, fromUser string, limit, offset uint) (_ []*inventory.Transfer, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting transfers from %q: %w", fromUser, err)
		}
	}()

	if limit == 0 {
		limit = inventory.DefaultListLimit
	} else if limit > inventory.MaxListLimit {
		limit = inventory.MaxListLimit
	}

	selectStmt, err := b.DB.PrepareContext(ctx, `SELECT transfers.id, transfers.request_id, to_users.username, transfers.opened, transfers.closed, SUM(tc.quantity)
FROM transfers
LEFT JOIN users from_users ON transfers.from_user = from_users.id
LEFT JOIN users to_users ON transfers.to_user = to_users.id
LEFT JOIN transferred_cards tc ON tc.transfer_id = transfers.id
WHERE from_users.username = ?
GROUP BY transfers.id
ORDER BY transfers.opened
LIMIT ?
OFFSET ?
`)
	if err != nil {
		return nil, fmt.Errorf("error preparing select query: %w", err)
	}
	defer selectStmt.Close()

	rows, err := selectStmt.QueryContext(ctx, fromUser, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error executing select query: %w", err)
	}

	transfers := make([]*inventory.Transfer, 0)
	for rows.Next() {
		var id int64
		var requestID sql.NullInt64
		var toUser string
		var opened time.Time
		var closed sql.NullTime
		var quantity uint
		err = rows.Scan(&id, &requestID, &toUser, &opened, &closed, &quantity)
		if err != nil {
			return nil, fmt.Errorf("error scanning row of select: %w", err)
		}
		transfer := &inventory.Transfer{
			ID:       id,
			ToUser:   toUser,
			FromUser: fromUser,
			Opened:   opened,
			Quantity: quantity,
		}
		if requestID.Valid {
			transfer.RequestID = &requestID.Int64
		}
		if closed.Valid {
			transfer.Closed = &closed.Time
		}
		transfers = append(transfers, transfer)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error getting next row of select: %w", err)
	}

	return transfers, nil
}

// GetTransfersByRequestID returns Transfers based on their RequestID
func (b *Backend) GetTransfersByRequestID(ctx context.Context, requestID int64, limit, offset uint) (_ []*inventory.Transfer, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting transfers with request ID \"%d\": %w", requestID, err)
		}
	}()

	if limit == 0 {
		limit = inventory.DefaultListLimit
	} else if limit > inventory.MaxListLimit {
		limit = inventory.MaxListLimit
	}

	selectStmt, err := b.DB.PrepareContext(ctx, `SELECT transfers.id, to_users.username, from_users.username, transfers.opened, transfers.closed, SUM(tc.quantity)
FROM transfers
LEFT JOIN users from_users ON transfers.from_user = from_users.id
LEFT JOIN users to_users ON transfers.to_user = to_users.id
LEFT JOIN transferred_cards tc ON tc.transfer_id = transfers.id
WHERE transfers.request_id = ?
GROUP BY transfers.id
ORDER BY transfers.opened
LIMIT ?
OFFSET ?
`)
	if err != nil {
		return nil, fmt.Errorf("error preparing select query: %w", err)
	}
	defer selectStmt.Close()

	rows, err := selectStmt.QueryContext(ctx, requestID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error executing select query: %w", err)
	}

	transfers := make([]*inventory.Transfer, 0)
	for rows.Next() {
		var id int64
		var toUser, fromUser string
		var opened time.Time
		var closed sql.NullTime
		var quantity uint
		err = rows.Scan(&id, &toUser, &fromUser, &opened, &closed, &quantity)
		if err != nil {
			return nil, fmt.Errorf("error scanning row of select: %w", err)
		}
		transfer := &inventory.Transfer{
			ID:        id,
			RequestID: &requestID,
			ToUser:    toUser,
			FromUser:  fromUser,
			Opened:    opened,
			Quantity:  quantity,
		}
		if closed.Valid {
			transfer.Closed = &closed.Time
		}
		transfers = append(transfers, transfer)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error getting next row of select: %w", err)
	}

	return transfers, nil
}

// GetTransferByID returns a Transfer based on its ID
func (b *Backend) GetTransferByID(ctx context.Context, id int64, limit, offset uint) (_ *inventory.Transfer, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting transfer \"%d\": %w", id, err)
		}
	}()

	if limit == 0 {
		limit = inventory.DefaultListLimit
	} else if limit > inventory.MaxListLimit {
		limit = inventory.MaxListLimit
	}

	selectTransferStmt, err := b.DB.PrepareContext(ctx, `SELECT transfers.request_id, to_users.username, from_users.username, opened, closed
FROM transfers
LEFT JOIN users to_users ON to_users.id = transfers.to_user
LEFT JOIN users from_users ON from_users.id = transfers.from_user
WHERE transfers.id = ?
`)
	if err != nil {
		return nil, fmt.Errorf("error preparing select for transfer: %w", err)
	}
	defer selectTransferStmt.Close()

	row := selectTransferStmt.QueryRowContext(ctx, id)
	var requestID sql.NullInt64
	var toUser, fromUser string
	var opened time.Time
	var closed sql.NullTime
	err = row.Scan(&requestID, &toUser, &fromUser, &opened, &closed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, inventory.ErrTransferNoExist
		}
		return nil, fmt.Errorf("error scanning row for transfer: %w", err)
	}

	transfer := &inventory.Transfer{
		ID:       id,
		ToUser:   toUser,
		FromUser: fromUser,
		Opened:   opened,
		Cards:    make([]*inventory.TransferredCards, 0),
	}
	if requestID.Valid {
		transfer.RequestID = &requestID.Int64
	}
	if closed.Valid {
		transfer.Closed = &closed.Time
	}

	selectCardsStmt, err := b.DB.PrepareContext(ctx, `SELECT tc.quantity, tc.name, tc.scryfall_id, tc.foil, owners.username
FROM transferred_cards AS tc
LEFT JOIN users owners ON owners.id = tc.owner
WHERE tc.transfer_id = ?
ORDER BY name, owners.username
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
		var quantity uint
		var name, scryfallID, owner string
		var foil bool
		err = rows.Scan(&quantity, &name, &scryfallID, &foil, &owner)
		if err != nil {
			return nil, fmt.Errorf("error scanning row for cards: %w", err)
		}
		transferredCardsRow := &inventory.TransferredCards{
			Quantity: quantity,
			Card: &inventory.Card{
				Name:       name,
				ScryfallID: scryfallID,
				Foil:       foil,
			},
			Owner: owner,
		}
		transfer.Cards = append(transfer.Cards, transferredCardsRow)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error getting next row of cards: %w", err)
	}

	return transfer, nil
}

// OpenTransfer creates a transfer
func (b *Backend) OpenTransfer(ctx context.Context, toUser, fromUser string, requestIDIn *int64, transferRows []*inventory.TransferredCards) (_ *inventory.Transfer, err error) {
	if len(transferRows) > inventory.RowUploadLimit {
		return nil, inventory.ErrTooManyRows
	}
	for _, row := range transferRows {
		if row.Quantity == 0 {
			return nil, &inventory.RowError{
				Err: inventory.ErrZeroCards,
				Row: row,
			}
		}
	}

	tx, err := b.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error transferring cards: %w", err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("error transferring cards: %w, unable to rollback: %s", err, rollbackErr)
			} else {
				err = fmt.Errorf("error transferring cards: %w", err)
			}
		}
	}()

	now := time.Now()

	var requestID sql.NullInt64
	if requestIDIn != nil {
		requestID.Int64 = *requestIDIn
		requestID.Valid = true
	}
	insertTransferStmt, err := tx.PrepareContext(ctx, `INSERT INTO transfers (to_user, from_user, request_id, opened)
SELECT to_users.id, from_users.id, ?, ?
FROM users to_users, users from_users
WHERE to_users.username = ? AND from_users.username = ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare insert transfer: %w", err)
	}
	defer insertTransferStmt.Close()

	result, err := insertTransferStmt.ExecContext(ctx, requestID, now, toUser, fromUser)
	if err != nil {
		return nil, fmt.Errorf("failed to insert transfer: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert: %w", err)
	}
	transfer := &inventory.Transfer{
		ID:        id,
		RequestID: requestIDIn,
		ToUser:    toUser,
		FromUser:  fromUser,
		Opened:    now,
		Cards:     transferRows,
	}

	selectQuantityStmt, err := tx.PrepareContext(ctx, `SELECT quantity
FROM cards
LEFT JOIN users owners ON owners.id = cards.owner
LEFT JOIN users keepers ON keepers.id = cards.keeper
WHERE cards.scryfall_id = ? AND cards.foil = ? AND owners.username = ? AND keepers.username = ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select for cards: %w", err)
	}
	defer selectQuantityStmt.Close()

	updateCardStmt, err := tx.PrepareContext(ctx, `UPDATE cards
LEFT JOIN users owners ON owners.id = cards.owner
LEFT JOIN users keepers ON keepers.id = cards.keeper
SET quantity = quantity - ?
WHERE cards.scryfall_id = ? AND cards.foil = ? AND owners.username = ? AND keepers.username = ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare update for cards: %w", err)
	}
	defer updateCardStmt.Close()

	upsertCardStmt, err := tx.PrepareContext(ctx, `INSERT INTO cards (quantity, name, oracle_id, scryfall_id, foil, owner, keeper)
SELECT ?, ?, ?, ?, ?, owners.id, keepers.id
FROM users owners, users keepers
WHERE owners.username = ? AND keepers.username = ?
ON DUPLICATE KEY UPDATE quantity = quantity + ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare upsert for cards: %w", err)
	}
	defer upsertCardStmt.Close()

	upsertTransferCardStmt, err := tx.PrepareContext(ctx, `INSERT INTO transferred_cards (transfer_id, quantity, name, scryfall_id, foil, owner)
SELECT ?, ?, ?, ?, ?, users.id
FROM users
WHERE users.username = ?
ON DUPLICATE KEY UPDATE quantity = quantity + ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare upsert for transferred_cards: %w", err)
	}
	defer upsertTransferCardStmt.Close()

	for _, transferRow := range transferRows {
		row := selectQuantityStmt.QueryRow(
			transferRow.Card.ScryfallID,
			transferRow.Card.Foil,
			transferRow.Owner,
			fromUser,
		)
		var quantity uint
		err = row.Scan(&quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to scan select row: %s", err)
		}
		if quantity < transferRow.Quantity {
			return nil, fmt.Errorf("too few cards to transfer %d copies of %s",
				transferRow.Quantity, transferRow.Card.ScryfallID)
		}

		_, err = updateCardStmt.ExecContext(ctx, transferRow.Quantity, transferRow.Card.ScryfallID, transferRow.Card.Foil, transferRow.Owner, fromUser)
		if err != nil {
			return nil, fmt.Errorf("failed to update cards: %w", err)
		}

		_, err = upsertCardStmt.ExecContext(ctx, transferRow.Quantity, transferRow.Card.Name, transferRow.Card.OracleID, transferRow.Card.ScryfallID, transferRow.Card.Foil, transferRow.Owner, toUser, transferRow.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert cards: %w", err)
		}

		_, err = upsertTransferCardStmt.ExecContext(ctx, transfer.ID, transferRow.Quantity, transferRow.Card.Name, transferRow.Card.ScryfallID, transferRow.Card.Foil, transferRow.Owner, transferRow.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert transferred_cards: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit inserts and updates on multiple tables: %w", err)
	}

	return transfer, nil
}

// CloseTransfer sets the closed time on a Transfer
func (b *Backend) CloseTransfer(ctx context.Context, id int64) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error closing transfer \"%d\": %w", id, err)
		}
	}()

	closeStmt, err := b.DB.PrepareContext(ctx, `UPDATE transfers
SET closed = NOW()
WHERE id = ?
`)
	if err != nil {
		return fmt.Errorf("error preparing update to close transfer: %w", err)
	}
	defer closeStmt.Close()

	result, err := closeStmt.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("error updating transfer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting transfer: %w", err)
	}

	if rowsAffected <= 0 {
		return inventory.ErrTransferNoExist
	}

	return nil
}
