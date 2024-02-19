package sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// GetCardsByOwner gets cards based on their owner
func GetCardsByOwner(ctx context.Context, db *sql.DB, owner *User, limit, offset int) (_ []*CardRow, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting cards by owner: %w", err)
		}
	}()

	queryStmt, err := db.PrepareContext(ctx, `SELECT cards.quantity, cards.english_name, cards.oracle_id, cards.scryfall_id, cards.foil, keepers.id, keepers.username, keepers.email
	FROM cards
	LEFT JOIN users keepers ON cards.keeper = keepers.id
	WHERE cards.owner = ?
	ORDER BY cards.english_name
	LIMIT ? OFFSET ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select for cards: %w", err)
	}
	defer queryStmt.Close()

	queryRows, err := queryStmt.QueryContext(ctx, owner.ID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to select for cards: %w", err)
	}

	cardRows := make([]*CardRow, 0)
	for queryRows.Next() {
		var quantity int
		var keeperID int64
		var englishName, oracleID, scryfallID, keeperUsername, keeperEmail string
		var foil bool
		err = queryRows.Scan(&quantity, &englishName, &oracleID, &scryfallID, &foil, &keeperID, &keeperUsername, &keeperEmail)
		if err != nil {
			return nil, fmt.Errorf("failed to scan select on cards: %w", err)
		}
		cardRow := &CardRow{
			Quantity: quantity,
			Card: &Card{
				EnglishName: englishName,
				OracleID:    oracleID,
				ScryfallID:  scryfallID,
				Foil:        foil,
			},
			Owner: owner,
			Keeper: &User{
				ID:       keeperID,
				Username: keeperUsername,
				Email:    keeperEmail,
			},
		}
		cardRows = append(cardRows, cardRow)
	}
	err = queryRows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to get next row on select on cards: %w", err)
	}

	return cardRows, nil
}

// GetCardsByKeeper gets cards based on their keeper
func GetCardsByKeeper(ctx context.Context, db *sql.DB, keeper *User, limit, offset int) (_ []*CardRow, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting cards by keeper: %w", err)
		}
	}()

	queryStmt, err := db.PrepareContext(ctx, `SELECT cards.quantity, cards.english_name, cards.oracle_id, cards.scryfall_id, cards.foil, owners.id, owners.username, owners.email
	FROM cards
	LEFT JOIN users owners ON cards.owner = owners.id
	WHERE cards.keeper = ?
	ORDER BY cards.english_name
	LIMIT ? OFFSET ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select for cards: %w", err)
	}
	defer queryStmt.Close()

	queryRows, err := queryStmt.QueryContext(ctx, keeper.ID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to select for cards: %w", err)
	}

	cardRows := make([]*CardRow, 0)
	for queryRows.Next() {
		var quantity int
		var ownerID int64
		var englishName, oracleID, scryfallID, ownerUsername, ownerEmail string
		var foil bool
		err = queryRows.Scan(&quantity, &englishName, &oracleID, &scryfallID, &foil, &ownerID, &ownerUsername, &ownerEmail)
		if err != nil {
			return nil, fmt.Errorf("failed to scan select on cards: %w", err)
		}
		cardRow := &CardRow{
			Quantity: quantity,
			Card: &Card{
				EnglishName: englishName,
				OracleID:    oracleID,
				ScryfallID:  scryfallID,
				Foil:        foil,
			},
			Owner: &User{
				ID:       ownerID,
				Username: ownerUsername,
				Email:    ownerEmail,
			},
			Keeper: keeper,
		}
		cardRows = append(cardRows, cardRow)
	}
	err = queryRows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to get next row on select on cards: %w", err)
	}

	return cardRows, nil
}

// AddCards adds cards given a slice of them
func AddCards(ctx context.Context, db *sql.DB, cardRows []*CardRow) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error adding cards: %w", err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("error adding cards: %w, unable to rollback: %s", err, rollbackErr)
			} else {
				err = fmt.Errorf("error adding cards: %w", err)
			}
		}
	}()

	upsertStmt, err := tx.PrepareContext(ctx, `INSERT INTO cards (quantity, english_name, oracle_id, scryfall_id, foil, owner, keeper)
VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE quantity = quantity + ?
`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert on cards: %w", err)
	}
	defer upsertStmt.Close()

	for _, cardRow := range cardRows {
		_, err := upsertStmt.ExecContext(ctx,
			cardRow.Quantity,
			cardRow.Card.EnglishName,
			cardRow.Card.OracleID,
			cardRow.Card.ScryfallID,
			cardRow.Card.Foil,
			cardRow.Owner.ID,
			cardRow.Keeper.ID,
			cardRow.Quantity,
		)
		if err != nil {
			return fmt.Errorf("failed to execute insert on cards: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit inserts on cards: %w", err)
	}

	return nil
}

// TransferCards transfers cards between two users
func TransferCards(ctx context.Context, db *sql.DB, toUser *User, fromUser *User, request *Request, transferRows []*TransferredCards) (_ *Transfer, err error) {
	tx, err := db.BeginTx(ctx, nil)
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

	var result sql.Result
	if request == nil {
		insertTransferStmt, err := tx.PrepareContext(ctx, `INSERT INTO transfers (to_user, from_user, created)
VALUES (?, ?, ?)
`)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare insert transfer without request id: %w", err)
		}
		defer insertTransferStmt.Close()

		result, err = insertTransferStmt.ExecContext(ctx, toUser.ID, fromUser.ID, now)
		if err != nil {
			return nil, fmt.Errorf("failed to insert transfer without request id: %w", err)
		}
	} else {
		insertTransferStmt, err := tx.PrepareContext(ctx, `INSERT INTO transfers (to_user, from_user, request_id, created)
VALUES (?, ?, ?, ?)
`)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare insert transfer with request id: %w", err)
		}
		defer insertTransferStmt.Close()

		result, err = insertTransferStmt.ExecContext(ctx, toUser.ID, fromUser.ID, request.ID, now)
		if err != nil {
			return nil, fmt.Errorf("failed to insert transfer with request id: %w", err)
		}
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id of transfer: %w", err)
	}
	transfer := &Transfer{
		ID:       id,
		ToUser:   toUser,
		FromUser: fromUser,
		Created:  now,
		Cards:    transferRows,
	}
	if request != nil {
		transfer.RequestID = request.ID
	}

	selectQuantityStmt, err := tx.PrepareContext(ctx, `SELECT quantity
FROM cards
WHERE scryfall_id = ? AND foil = ? AND owner = ? AND keeper =  ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select for cards: %w", err)
	}
	defer selectQuantityStmt.Close()

	updateCardStmt, err := tx.PrepareContext(ctx, `UPDATE cards
SET quantity = quantity - ?
WHERE scryfall_id = ? AND foil = ? AND owner = ? AND keeper = ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare update for cards: %w", err)
	}
	defer updateCardStmt.Close()

	upsertCardStmt, err := tx.PrepareContext(ctx, `INSERT INTO cards (quantity, english_name, oracle_id, scryfall_id, foil, owner, keeper)
VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE quantity = quantity + ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare upsert for cards: %w", err)
	}
	defer upsertCardStmt.Close()

	upsertTransferCardStmt, err := tx.PrepareContext(ctx, `INSERT INTO transferred_cards (transfer_id, quantity, english_name, scryfall_id, foil, owner)
VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE quantity = quantity + ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare upsert for transferred_cards: %w", err)
	}
	defer upsertTransferCardStmt.Close()

	for _, transferRow := range transferRows {
		row := selectQuantityStmt.QueryRow(
			transferRow.Card.ScryfallID,
			transferRow.Card.Foil,
			transferRow.Owner.ID,
			fromUser.ID,
		)
		var quantity int
		err = row.Scan(&quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to scan select row: %s", err)
		}
		if quantity < transferRow.Quantity {
			return nil, fmt.Errorf("too few cards to transfer %d copies of %s",
				transferRow.Quantity, transferRow.Card.ScryfallID)
		}

		_, err = updateCardStmt.ExecContext(ctx, transferRow.Quantity, transferRow.Card.ScryfallID, transferRow.Card.Foil, transferRow.Owner.ID, fromUser.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to update cards: %w", err)
		}

		_, err = upsertCardStmt.ExecContext(ctx, transferRow.Quantity, transferRow.Card.EnglishName, transferRow.Card.OracleID, transferRow.Card.ScryfallID, transferRow.Card.Foil, transferRow.Owner.ID, toUser.ID, transferRow.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert cards: %w", err)
		}

		_, err = upsertTransferCardStmt.ExecContext(ctx, transfer.ID, transferRow.Quantity, transferRow.Card.EnglishName, transferRow.Card.ScryfallID, transferRow.Card.Foil, transferRow.Owner.ID, transferRow.Quantity)
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
