package sql

import (
	"context"
	"fmt"

	inventory "github.com/benrm/mtg-inventory/golang/mtg-inventory"
)

// AddUserIfNotExist adds a User given its username and email
func (b *Backend) AddUserIfNotExist(ctx context.Context, slackID string) (*inventory.User, error) {
	tx, err := b.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error adding user: %w", err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("error adding user: %w, unable to rollback: %s", err, rollbackErr)
			} else {
				err = fmt.Errorf("error adding user: %w", err)
			}
		}
	}()

	selectStmt, err := tx.PrepareContext(ctx, "SELECT COUNT(*) FROM users WHERE slack_id = ?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select on users: %w", err)
	}

	row := selectStmt.QueryRowContext(ctx, slackID)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to select on users: %w", err)
	}
	if count > 0 {
		return &inventory.User{
			SlackID: slackID,
		}, nil
	}

	insertStmt, err := tx.PrepareContext(ctx, "INSERT INTO users (slack_id) VALUES (?)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare insert on users: %w", err)
	}
	defer insertStmt.Close()

	_, err = insertStmt.ExecContext(ctx, slackID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert on users: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit insert on users: %w", err)
	}

	return &inventory.User{
		SlackID: slackID,
	}, nil
}
