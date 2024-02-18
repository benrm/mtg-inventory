package sql

import (
	"context"
	"database/sql"
	"fmt"
)

// GetUserByID gets a User by its ID
func GetUserByID(ctx context.Context, db *sql.DB, id int64) (_ *User, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error adding user: %w", err)
		}
	}()

	queryStmt, err := db.PrepareContext(ctx, "SELECT username, email FROM users WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select on users: %w", err)
	}

	var username, email string
	row := queryStmt.QueryRow(id)
	err = row.Scan(&username, &email)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row from select on users: %w", err)
	}

	return &User{
		ID:       id,
		Username: username,
		Email:    email,
	}, nil
}

// GetUserByUsername gets a User by its username
func GetUserByUsername(ctx context.Context, db *sql.DB, username string) (_ *User, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error adding user: %w", err)
		}
	}()

	queryStmt, err := db.PrepareContext(ctx, "SELECT id, email FROM users WHERE username = ?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select on users: %w", err)
	}

	var id int64
	var email string
	row := queryStmt.QueryRow(username)
	err = row.Scan(&id, &email)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row from select on users: %w", err)
	}

	return &User{
		ID:       id,
		Username: username,
		Email:    email,
	}, nil
}

// AddUser adds a User given its username and email
func AddUser(ctx context.Context, db *sql.DB, username, email string) (*User, error) {
	tx, err := db.BeginTx(ctx, nil)
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

	result, err := insertStmt.Exec(username, email)
	if err != nil {
		return nil, fmt.Errorf("failed to insert on users: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit insert on users: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id from insert on users: %w", err)
	}

	return &User{
		ID:       id,
		Username: username,
		Email:    email,
	}, nil
}
