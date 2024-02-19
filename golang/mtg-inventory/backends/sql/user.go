package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	inventory "github.com/benrm/mtg-inventory/golang/mtg-inventory"
)

func getUserByUsername(ctx context.Context, sp stmtPreparer, username string) (_ *fullUser, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting user by username %s: %w", username, err)
		}
	}()

	queryStmt, err := sp.PrepareContext(ctx, "SELECT id, email FROM users WHERE username = ?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select on users: %w", err)
	}
	defer queryStmt.Close()

	var id int64
	var email string
	row := queryStmt.QueryRow(username)
	err = row.Scan(&id, &email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, inventory.ErrUserNoExist
		}
		return nil, fmt.Errorf("failed to scan row from select on users: %w", err)
	}

	return &fullUser{
		ID: id,
		User: &inventory.User{
			Username: username,
			Email:    email,
		},
	}, nil
}

// GetUserByUsername gets a User by its username
func (b *Backend) GetUserByUsername(ctx context.Context, username string) (_ *inventory.User, err error) {
	user, err := getUserByUsername(ctx, b.DB, username)
	if err != nil {
		return user.User, nil
	}
	return nil, err
}

// AddUser adds a User given its username and email
func (b *Backend) AddUser(ctx context.Context, username, email string) (*inventory.User, error) {
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

	insertStmt, err := tx.PrepareContext(ctx, "INSERT INTO users (username, email) VALUES (?, ?)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare insert on users: %w", err)
	}
	defer insertStmt.Close()

	_, err = insertStmt.ExecContext(ctx, username, email)
	if err != nil {
		return nil, fmt.Errorf("failed to insert on users: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit insert on users: %w", err)
	}

	return &inventory.User{
		Username: username,
		Email:    email,
	}, nil
}
