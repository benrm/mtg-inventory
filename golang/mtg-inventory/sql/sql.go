package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

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

	queryStmt, err := tx.PrepareContext(ctx, "SELECT quantity FROM cards WHERE scryfall_id = ? AND foil = ? AND owner = ? AND keeper = ?")
	if err != nil {
		return err
	}

	insertStmt, err := tx.PrepareContext(ctx, "INSERT INTO cards (quantity, oracle_id, scryfall_id, foil, owner, keeper) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	updateStmt, err := tx.PrepareContext(ctx, "UPDATE cards SET quantity = ? WHERE scryfall_id = ? AND foil = ? AND OWNER = ? AND KEEPER = ?")
	if err != nil {
		return err
	}

	for _, cardRow := range cardRows {
		queryRow := queryStmt.QueryRow(
			cardRow.Card.ScryfallID,
			cardRow.Card.Foil,
			cardRow.Owner.ID,
			cardRow.Keeper.ID,
		)
		var quantity int
		err = queryRow.Scan(&quantity)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			_, err := insertStmt.Exec(
				cardRow.Quantity,
				cardRow.Card.OracleID,
				cardRow.Card.ScryfallID,
				cardRow.Card.Foil,
				cardRow.Owner.ID,
				cardRow.Keeper.ID,
			)
			if err != nil {
				return err
			}
		} else {
			if err != nil {
				return err
			}
			_, err := updateStmt.Exec(
				quantity+cardRow.Quantity,
				cardRow.Card.ScryfallID,
				cardRow.Card.Foil,
				cardRow.Owner.ID,
				cardRow.Keeper.ID,
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
