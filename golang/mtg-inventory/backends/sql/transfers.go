package sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	inventory "github.com/benrm/mtg-inventory/golang/mtg-inventory"
)

// GetTransfersByToUser returns Transfers based on their ToUser
func (b *Backend) GetTransfersByToUser(_ context.Context, _ string, _, _ uint) (_ []*inventory.Transfer, _ error) {
	return nil, inventory.ErrUnimplemented
}

// GetTransfersByFromUser returns Transfers based on their FromUser
func (b *Backend) GetTransfersByFromUser(_ context.Context, _ string, _, _ uint) (_ []*inventory.Transfer, _ error) {
	return nil, inventory.ErrUnimplemented
}

// GetTransfersByRequestID returns Transfers based on their RequestID
func (b *Backend) GetTransfersByRequestID(_ context.Context, _ int64, _, _ uint) (_ []*inventory.Transfer, _ error) {
	return nil, inventory.ErrUnimplemented
}

// GetTransferByID returns a Transfer based on its ID
func (b *Backend) GetTransferByID(_ context.Context, _ int64) (_ *inventory.Transfer, _ error) {
	return nil, inventory.ErrUnimplemented
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
