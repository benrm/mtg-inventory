package sql

import (
	"context"
	"database/sql"
	"fmt"
)

func GetUserByID(ctx context.Context, db *sql.DB, id int64) (_ *User, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error adding user: %w", err)
		}
	}()

	queryStmt, err := db.PrepareContext(ctx, "SELECT username, email FROM users WHERE id = ?")
	if err != nil {
		return nil, err
	}

	var username, email string
	row := queryStmt.QueryRow(id)
	err = row.Scan(&username, &email)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       id,
		Username: username,
		Email:    email,
	}, nil
}

func GetUserByUsername(ctx context.Context, db *sql.DB, username string) (_ *User, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error adding user: %w", err)
		}
	}()

	queryStmt, err := db.PrepareContext(ctx, "SELECT id, email FROM users WHERE username = ?")
	if err != nil {
		return nil, err
	}

	var id int64
	var email string
	row := queryStmt.QueryRow(username)
	err = row.Scan(&id, &email)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       id,
		Username: username,
		Email:    email,
	}, nil
}

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
		return nil, err
	}

	result, err := insertStmt.Exec(username, email)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       id,
		Username: username,
		Email:    email,
	}, nil
}
