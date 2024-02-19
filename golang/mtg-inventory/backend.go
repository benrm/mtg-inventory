package inventory

import (
	"context"
)

// Backend describes an object that maintains state about a Magic: the
// Gathering inventory
type Backend interface {
	GetCardsByOracleID(ctx context.Context, oracleID string, limit, offset int) ([]*CardRow, error)
	GetCardsByOwner(ctx context.Context, owner string, limit, offset int) ([]*CardRow, error)
	GetCardsByKeeper(ctx context.Context, keeper string, limit, offset int) ([]*CardRow, error)
	AddCards(ctx context.Context, cardRows []*CardRow) error
	TransferCards(ctx context.Context, toUser, fromUser string, request *int64, rows []*TransferredCards) (*Transfer, error)

	GetUserByUsername(ctx context.Context, username string) (*User, error)
	AddUser(ctx context.Context, username, email string) (*User, error)
}
