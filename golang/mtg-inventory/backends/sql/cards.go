package sql

import (
	"context"
	"fmt"

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
SELECT ?, ?, ?, ?, ?, owners.id, keepers.id
FROM users owners, users keepers
WHERE owners.username = ? AND keepers.username = ?
ON DUPLICATE KEY UPDATE quantity = quantity + ?
`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert on cards: %w", err)
	}
	defer upsertStmt.Close()

	for _, cardRow := range cardRows {
		_, err := upsertStmt.ExecContext(ctx,
			cardRow.Quantity,
			cardRow.Card.Name,
			cardRow.Card.OracleID,
			cardRow.Card.ScryfallID,
			cardRow.Card.Foil,
			cardRow.Owner,
			cardRow.Keeper,
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
