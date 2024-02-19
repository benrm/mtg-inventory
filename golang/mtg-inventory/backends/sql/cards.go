package sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	inventory "github.com/benrm/mtg-inventory/golang/mtg-inventory"
)

// GetCardsByOracleID gets cards based on their Oracle ID
func (b *Backend) GetCardsByOracleID(ctx context.Context, oracleID string, limit, offset int) (_ []*inventory.CardRow, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting cards by oracle ID: %w", err)
		}
	}()

	queryStmt, err := b.DB.PrepareContext(ctx, `SELECT cards.quantity, cards.name, cards.scryfall_id, cards.foil, owners.username, keepers.username
FROM cards
LEFT JOIN users owners ON cards.owner = owners.id
LEFT JOIN users keepers ON cards.keeper = keepers.id
WHERE cards.oracle_id = ?
ORDER BY cards.name
LIMIT ? OFFSET ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select for cards: %w", err)
	}
	defer queryStmt.Close()

	queryRows, err := queryStmt.QueryContext(ctx, oracleID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to select for cards: %w", err)
	}

	cardRows := make([]*inventory.CardRow, 0)
	for queryRows.Next() {
		var quantity int
		var cardName, scryfallID, ownerUsername, keeperUsername string
		var foil bool
		err = queryRows.Scan(&quantity, &cardName, &scryfallID, &foil, &ownerUsername, &keeperUsername)
		if err != nil {
			return nil, fmt.Errorf("failed to scan select on cards: %w", err)
		}
		cardRow := &inventory.CardRow{
			Quantity: quantity,
			Card: &inventory.Card{
				Name:       cardName,
				OracleID:   oracleID,
				ScryfallID: scryfallID,
				Foil:       foil,
			},
			Owner:  ownerUsername,
			Keeper: keeperUsername,
		}
		cardRows = append(cardRows, cardRow)
	}
	err = queryRows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to get next row on select on cards: %w", err)
	}

	return cardRows, nil
}

// GetCardsByOwner gets cards based on their owner
func (b *Backend) GetCardsByOwner(ctx context.Context, ownerUsername string, limit, offset int) (_ []*inventory.CardRow, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting cards by owner: %w", err)
		}
	}()

	queryStmt, err := b.DB.PrepareContext(ctx, `SELECT cards.quantity, cards.name, cards.oracle_id, cards.scryfall_id, cards.foil, keepers.username
	FROM cards
	LEFT JOIN users owners ON cards.owner = owners.id
	LEFT JOIN users keepers ON cards.keeper = keepers.id
	WHERE owners.username = ?
	ORDER BY cards.name
	LIMIT ? OFFSET ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select for cards: %w", err)
	}
	defer queryStmt.Close()

	queryRows, err := queryStmt.QueryContext(ctx, ownerUsername, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to select for cards: %w", err)
	}

	cardRows := make([]*inventory.CardRow, 0)
	for queryRows.Next() {
		var quantity int
		var cardName, oracleID, scryfallID, keeperUsername string
		var foil bool
		err = queryRows.Scan(&quantity, &cardName, &oracleID, &scryfallID, &foil, &keeperUsername)
		if err != nil {
			return nil, fmt.Errorf("failed to scan select on cards: %w", err)
		}
		cardRow := &inventory.CardRow{
			Quantity: quantity,
			Card: &inventory.Card{
				Name:       cardName,
				OracleID:   oracleID,
				ScryfallID: scryfallID,
				Foil:       foil,
			},
			Owner:  ownerUsername,
			Keeper: keeperUsername,
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
func (b *Backend) GetCardsByKeeper(ctx context.Context, keeperUsername string, limit, offset int) (_ []*inventory.CardRow, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting cards by keeper: %w", err)
		}
	}()

	queryStmt, err := b.DB.PrepareContext(ctx, `SELECT cards.quantity, cards.name, cards.oracle_id, cards.scryfall_id, cards.foil, owners.username
	FROM cards
	LEFT JOIN users owners ON cards.owner = owners.id
	LEFT JOIN users keepers ON cards.keeper = keepers.id
	WHERE keepers.username = ?
	ORDER BY cards.name
	LIMIT ? OFFSET ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select for cards: %w", err)
	}
	defer queryStmt.Close()

	queryRows, err := queryStmt.QueryContext(ctx, keeperUsername, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to select for cards: %w", err)
	}

	cardRows := make([]*inventory.CardRow, 0)
	for queryRows.Next() {
		var quantity int
		var cardName, oracleID, scryfallID, ownerUsername string
		var foil bool
		err = queryRows.Scan(&quantity, &cardName, &oracleID, &scryfallID, &foil, &ownerUsername)
		if err != nil {
			return nil, fmt.Errorf("failed to scan select on cards: %w", err)
		}
		cardRow := &inventory.CardRow{
			Quantity: quantity,
			Card: &inventory.Card{
				Name:       cardName,
				OracleID:   oracleID,
				ScryfallID: scryfallID,
				Foil:       foil,
			},
			Owner:  ownerUsername,
			Keeper: keeperUsername,
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
func (b *Backend) AddCards(ctx context.Context, cardRows []*inventory.CardRow) (err error) {
	tx, err := b.DB.BeginTx(ctx, nil)
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

	upsertStmt, err := tx.PrepareContext(ctx, `INSERT INTO cards (quantity, name, oracle_id, scryfall_id, foil, owner, keeper)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE quantity = quantity + ?
`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert on cards: %w", err)
	}
	defer upsertStmt.Close()

	userMap := make(map[string]int64)

	for _, cardRow := range cardRows {
		if _, exists := userMap[cardRow.Owner]; !exists {
			user, err := getUserByUsername(ctx, tx, cardRow.Owner)
			if err != nil {
				return fmt.Errorf("failed to add cards: %w", err)
			}
			userMap[cardRow.Owner] = user.ID
		}
		if _, exists := userMap[cardRow.Keeper]; !exists {
			user, err := getUserByUsername(ctx, tx, cardRow.Keeper)
			if err != nil {
				return fmt.Errorf("failed to add cards: %w", err)
			}
			userMap[cardRow.Keeper] = user.ID
		}
		_, err := upsertStmt.ExecContext(ctx,
			cardRow.Quantity,
			cardRow.Card.Name,
			cardRow.Card.OracleID,
			cardRow.Card.ScryfallID,
			cardRow.Card.Foil,
			userMap[cardRow.Owner],
			userMap[cardRow.Keeper],
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
func (b *Backend) TransferCards(ctx context.Context, toUser, fromUser string, requestIDIn *int64, transferRows []*inventory.TransferredCards) (_ *inventory.Transfer, err error) {
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

	userMap := make(map[string]int64)

	user, err := getUserByUsername(ctx, tx, toUser)
	if err != nil {
		return nil, fmt.Errorf("error getting to_user info: %w", err)
	}
	userMap[toUser] = user.ID
	if toUser != fromUser {
		user, err := getUserByUsername(ctx, tx, fromUser)
		if err != nil {
			return nil, fmt.Errorf("error getting from_user info: %w", err)
		}
		userMap[fromUser] = user.ID
	}

	now := time.Now()

	var requestID sql.NullInt64
	if requestIDIn != nil {
		requestID.Int64 = *requestIDIn
		requestID.Valid = true
	}
	insertTransferStmt, err := tx.PrepareContext(ctx, `INSERT INTO transfers (to_user, from_user, request_id, created)
VALUES (?, ?, ?, ?)
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare insert transfer: %w", err)
	}
	defer insertTransferStmt.Close()

	result, err := insertTransferStmt.ExecContext(ctx, userMap[toUser], userMap[fromUser], requestID, now)
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
		Created:   now,
		Cards:     transferRows,
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

	upsertCardStmt, err := tx.PrepareContext(ctx, `INSERT INTO cards (quantity, name, oracle_id, scryfall_id, foil, owner, keeper)
VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE quantity = quantity + ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare upsert for cards: %w", err)
	}
	defer upsertCardStmt.Close()

	upsertTransferCardStmt, err := tx.PrepareContext(ctx, `INSERT INTO transferred_cards (transfer_id, quantity, name, scryfall_id, foil, owner)
VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE quantity = quantity + ?
`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare upsert for transferred_cards: %w", err)
	}
	defer upsertTransferCardStmt.Close()

	for _, transferRow := range transferRows {
		if _, exists := userMap[transferRow.Owner]; !exists {
			owner, err := getUserByUsername(ctx, tx, transferRow.Owner)
			if err != nil {
				return nil, fmt.Errorf("failed to get owner: %w", err)
			}
			userMap[transferRow.Owner] = owner.ID
		}
		row := selectQuantityStmt.QueryRow(
			transferRow.Card.ScryfallID,
			transferRow.Card.Foil,
			userMap[transferRow.Owner],
			userMap[fromUser],
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

		_, err = updateCardStmt.ExecContext(ctx, transferRow.Quantity, transferRow.Card.ScryfallID, transferRow.Card.Foil, userMap[transferRow.Owner], userMap[fromUser])
		if err != nil {
			return nil, fmt.Errorf("failed to update cards: %w", err)
		}

		_, err = upsertCardStmt.ExecContext(ctx, transferRow.Quantity, transferRow.Card.Name, transferRow.Card.OracleID, transferRow.Card.ScryfallID, transferRow.Card.Foil, userMap[transferRow.Owner], userMap[toUser], transferRow.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert cards: %w", err)
		}

		_, err = upsertTransferCardStmt.ExecContext(ctx, transfer.ID, transferRow.Quantity, transferRow.Card.Name, transferRow.Card.ScryfallID, transferRow.Card.Foil, userMap[transferRow.Owner], transferRow.Quantity)
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
