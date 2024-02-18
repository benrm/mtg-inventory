package sql

import (
	"context"
	"database/sql"
	"fmt"
)

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
		return nil, err
	}

	queryRows, err := queryStmt.Query(owner.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	cardRows := make([]*CardRow, 0)
	for queryRows.Next() {
		var quantity int
		var keeperID int64
		var englishName, oracleID, scryfallID, keeperUsername, keeperEmail string
		var foil bool
		err = queryRows.Scan(&quantity, &englishName, &oracleID, &scryfallID, &foil, &keeperID, &keeperUsername, &keeperEmail)
		if err != nil {
			return nil, err
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

	return cardRows, nil
}

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
		return nil, err
	}

	queryRows, err := queryStmt.Query(keeper.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	cardRows := make([]*CardRow, 0)
	for queryRows.Next() {
		var quantity int
		var ownerID int64
		var englishName, oracleID, scryfallID, ownerUsername, ownerEmail string
		var foil bool
		err = queryRows.Scan(&quantity, &englishName, &oracleID, &scryfallID, &foil, &ownerID, &ownerUsername, &ownerEmail)
		if err != nil {
			return nil, err
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

	return cardRows, nil
}

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

	insertStmt, err := tx.PrepareContext(ctx, `INSERT INTO cards (quantity, english_name, oracle_id, scryfall_id, foil, owner, keeper)
VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE quantity = quantity + ?
`)
	if err != nil {
		return err
	}

	for _, cardRow := range cardRows {
		_, err := insertStmt.Exec(
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
			return err
		}
	}

	return tx.Commit()
}
