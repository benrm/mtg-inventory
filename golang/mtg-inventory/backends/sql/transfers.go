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

	selectStmt, err := b.DB.PrepareContext(ctx, `SELECT transfers.id, transfers.request_id, from_users.slack_id, transfers.opened, transfers.closed, SUM(tc.quantity)
FROM transfers
LEFT JOIN users from_users ON transfers.from_user = from_users.id
LEFT JOIN users to_users ON transfers.to_user = to_users.id
LEFT JOIN transferred_cards tc ON tc.transfer_id = transfers.id
WHERE to_users.slack_id = ?
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

	selectStmt, err := b.DB.PrepareContext(ctx, `SELECT transfers.id, transfers.request_id, to_users.slack_id, transfers.opened, transfers.closed, SUM(tc.quantity)
FROM transfers
LEFT JOIN users from_users ON transfers.from_user = from_users.id
LEFT JOIN users to_users ON transfers.to_user = to_users.id
LEFT JOIN transferred_cards tc ON tc.transfer_id = transfers.id
WHERE from_users.slack_id = ?
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

	selectStmt, err := b.DB.PrepareContext(ctx, `SELECT transfers.id, to_users.slack_id, from_users.slack_id, transfers.opened, transfers.closed, SUM(tc.quantity)
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

	selectTransferStmt, err := b.DB.PrepareContext(ctx, `SELECT transfers.request_id, to_users.slack_id, from_users.slack_id, opened, closed
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

	selectCardsStmt, err := b.DB.PrepareContext(ctx, `SELECT tc.quantity, tc.name, tc.scryfall_id, tc.foil, owners.slack_id
FROM transferred_cards AS tc
LEFT JOIN users owners ON owners.id = tc.owner
WHERE tc.transfer_id = ?
ORDER BY name, owners.slack_id
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
		return nil, fmt.Errorf("error opening transfer: %w", err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("error opening transfer: %w, unable to rollback: %s", err, rollbackErr)
			} else {
				err = fmt.Errorf("error opening transfer: %w", err)
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
WHERE to_users.slack_id = ? AND from_users.slack_id = ?
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
WHERE cards.scryfall_id = ? AND cards.foil = ? AND owners.slack_id = ? AND keepers.slack_id = ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select for cards: %w", err)
	}
	defer selectQuantityStmt.Close()

	updateCardStmt, err := tx.PrepareContext(ctx, `UPDATE cards
LEFT JOIN users owners ON owners.id = cards.owner
LEFT JOIN users keepers ON keepers.id = cards.keeper
SET quantity = quantity - ?
WHERE cards.scryfall_id = ? AND cards.foil = ? AND owners.slack_id = ? AND keepers.slack_id = ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare update for cards: %w", err)
	}
	defer updateCardStmt.Close()

	upsertCardStmt, err := tx.PrepareContext(ctx, `INSERT INTO cards (quantity, name, oracle_id, scryfall_id, foil, owner, keeper)
SELECT ?, ?, ?, ?, ?, owners.id, keepers.id
FROM users owners, users keepers
WHERE owners.slack_id = ? AND keepers.slack_id = ?
ON DUPLICATE KEY UPDATE quantity = quantity + ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare upsert for cards: %w", err)
	}
	defer upsertCardStmt.Close()

	upsertTransferCardStmt, err := tx.PrepareContext(ctx, `INSERT INTO transferred_cards (transfer_id, quantity, name, scryfall_id, foil, owner)
SELECT ?, ?, ?, ?, ?, users.id
FROM users
WHERE users.slack_id = ?
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

// CancelTransfer cancels a transfer, deleting it from the database
func (b *Backend) CancelTransfer(ctx context.Context, id int64) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error canceling transfer \"%d\": %w", id, err)
		}
	}()

	deleteStmt, err := b.DB.PrepareContext(ctx, `DELETE FROM transfers WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("error preparing transfer delete: %w", err)
	}
	defer deleteStmt.Close()

	result, err := deleteStmt.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("error executing transfer delete: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected <= 0 {
		return inventory.ErrTransferNoExist
	}

	return nil
}

// CloseTransfer sets the closed time on a Transfer and reassigns the keeper of
// the cards
func (b *Backend) CloseTransfer(ctx context.Context, id int64) (err error) {
	tx, err := b.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error closing transfer \"%d\": %w", id, err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("error closing transfer \"%d\": %w, unable to rollback: %s", id, err, rollbackErr)
			} else {
				err = fmt.Errorf("error closing transfer \"%d\": %w", id, err)
			}
		}
	}()

	selectToUserStmt, err := tx.PrepareContext(ctx, `SELECT to_users.id, to_users.slack_id
FROM transfers
LEFT JOIN users to_users ON transfers.to_user = users.id
WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("error preparing select for to user: %w", err)
	}
	defer selectToUserStmt.Close()

	var toUserID int64
	var toUser string
	queryRow := selectToUserStmt.QueryRowContext(ctx, id)
	err = queryRow.Scan(&toUserID, &toUser)
	if err != nil {
		return fmt.Errorf("error querying to user: %w", err)
	}

	selectCards, err := tx.PrepareContext(ctx, `SELECT cards.name, cards.oracle_id, cards.scryfall_id, cards.foil, cards.quantity, tc.quantity, owners.slack_id, owners.id, from_users.slack_id, from_users.id
FROM transfers
LEFT JOIN transferred_cards tc ON transfers.id = tc.transfer_id
LEFT JOIN cards ON tc.scryfall_id = cards.scryfall_id AND tc.foil = cards.FOIL AND tc.owner = cards.owner AND transfers.from_user = cards.keeper
LEFT JOIN users owners ON owners.id = cards.owner
LEFT JOIN users from_users ON transfers.from_user = users.id
WHERE transfer_id = ?
`)
	if err != nil {
		return fmt.Errorf("error preparing select: %w", err)
	}
	defer selectCards.Close()

	rows, err := selectCards.QueryContext(ctx, id)
	if err != nil {
		return fmt.Errorf("error selecting: %w", err)
	}

	type transferredCards struct {
		name             string
		oracleID         string
		scryfallID       string
		foil             bool
		actualQuantity   uint
		transferQuantity uint
		owner            string
		ownerID          int64
		fromUser         string
		fromUserID       int64
	}

	transferRows := make([]*transferredCards, 0)
	for rows.Next() {
		var tc transferredCards
		err = rows.Scan(&tc.name, &tc.oracleID, &tc.scryfallID, &tc.foil, &tc.actualQuantity, &tc.transferQuantity, &tc.owner, &tc.ownerID, &tc.fromUser, &tc.fromUserID)
		if err != nil {
			return fmt.Errorf("error scanning on select on transferred_cards: %w", err)
		}
		if tc.actualQuantity < tc.transferQuantity {
			return &inventory.RowError{
				Err: inventory.ErrTooFewCards,
				Row: &inventory.TransferredCards{
					Quantity: tc.transferQuantity,
					Card: &inventory.Card{
						ScryfallID: tc.scryfallID,
						Foil:       tc.foil,
					},
					Owner: tc.owner,
				},
			}
		}
		transferRows = append(transferRows, &tc)
	}
	err = rows.Err()
	if err != nil {
		return fmt.Errorf("error scanning on select on transferred_cards: %w", err)
	}

	removeStmt, err := tx.PrepareContext(ctx, `UPDATE cards
SET quantity = quantity - ?
WHERE scryfall_id = ? AND foil = ? AND owner = ? AND keeper = ?`)
	if err != nil {
		return fmt.Errorf("error preparing update statement on cards: %w", err)
	}
	defer removeStmt.Close()

	deleteStmt, err := tx.PrepareContext(ctx, `DELETE FROM cards
WHERE scryfall_id = ? AND foil = ? AND owner = ? AND keeper = ?`)
	if err != nil {
		return fmt.Errorf("error preparing delete statement on cards: %w", err)
	}
	defer deleteStmt.Close()

	upsertStmt, err := tx.PrepareContext(ctx, `INSERT INTO cards (quantity, name, oracle_id, scryfall_id, foil, owner, keeper)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE quantity = quantity + ?`)
	if err != nil {
		return fmt.Errorf("error preparing upsert statement on cards: %w", err)
	}
	defer upsertStmt.Close()

	for _, row := range transferRows {
		if row.actualQuantity == row.transferQuantity {
			_, err = deleteStmt.ExecContext(ctx, row.scryfallID, row.foil, row.ownerID, row.fromUserID)
			if err != nil {
				return fmt.Errorf("error deleting from cards: %w", err)
			}
		} else {
			_, err = removeStmt.ExecContext(ctx, row.transferQuantity, row.scryfallID, row.foil, row.ownerID, row.fromUserID)
			if err != nil {
				return fmt.Errorf("error removing quantity from cards: %w", err)
			}
		}
		_, err = upsertStmt.ExecContext(ctx, row.transferQuantity, row.name, row.oracleID, row.scryfallID, row.foil, row.ownerID, toUserID, row.transferQuantity)
		if err != nil {
			return fmt.Errorf("error upserting into cards: %w", err)
		}
	}

	closeStmt, err := tx.PrepareContext(ctx, `UPDATE transfers
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

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing: %w", err)
	}

	return nil
}
