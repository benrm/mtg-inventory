package inventory

import (
	"context"
)

// Backend describes an object that maintains state about a Magic: the
// Gathering inventory
type Backend interface {
	GetCardsByOracleID(ctx context.Context, oracleID string) ([]*CardRow, error)
	GetCardsByOwner(ctx context.Context, owner *User, limit, offset int) ([]*CardRow, error)
	GetCardsByKeeper(ctx context.Context, keeper *User, limit, offset int) ([]*CardRow, error)
	AddCards(ctx context.Context, cardRows []*CardRow) error
	TransferCards(ctx context.Context, toUser, fromUser *User, request *Request, rows []*TransferredCards) (*Transfer, error)

	GetUserByID(ctx context.Context, id int64) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	AddUser(ctx context.Context, username, email string) (*User, error)
}
