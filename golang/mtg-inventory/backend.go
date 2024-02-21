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

	GetRequestsByRequestor(ctx context.Context, requestor string, limit, offset int) ([]*Request, error)
	GetRequestByID(ctx context.Context, id int64, limit, offset int) (*Request, error)
	RequestCards(ctx context.Context, requestor string, rows []*RequestedCards) (*Request, error)
	CloseRequest(ctx context.Context, id int64) error

	TransferCards(ctx context.Context, toUser, fromUser string, request *int64, rows []*TransferredCards) (*Transfer, error)

	GetUserByUsername(ctx context.Context, username string) (*User, error)
	AddUser(ctx context.Context, username, email string) (*User, error)
}
