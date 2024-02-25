package inventory

import (
	"context"
)

// Backend describes an object that maintains state about a Magic: the
// Gathering inventory
type Backend interface {
	GetCardsByOracleID(ctx context.Context, oracleID string, limit, offset uint) ([]*CardRow, error)
	GetCardsByOwner(ctx context.Context, owner string, limit, offset uint) ([]*CardRow, error)
	GetCardsByKeeper(ctx context.Context, keeper string, limit, offset uint) ([]*CardRow, error)
	AddCards(ctx context.Context, cardRows []*CardRow) error
	ModifyCardQuantity(ctx context.Context, owner, keeper, scryfallID string, foil bool, quantity uint) error

	GetRequestsByRequestor(ctx context.Context, requestor string, limit, offset uint) ([]*Request, error)
	GetRequestByID(ctx context.Context, id int64, limit, offset uint) (*Request, error)
	OpenRequest(ctx context.Context, requestor string, rows []*RequestedCards) (*Request, error)
	CloseRequest(ctx context.Context, id int64) error

	GetTransfersByToUser(ctx context.Context, toUser string, limit, offset uint) ([]*Transfer, error)
	GetTransfersByFromUser(ctx context.Context, fromUser string, limit, offset uint) ([]*Transfer, error)
	GetTransfersByRequestID(ctx context.Context, requestID int64, limit, offset uint) ([]*Transfer, error)
	GetTransferByID(ctx context.Context, id int64, limit, offset uint) (*Transfer, error)
	OpenTransfer(ctx context.Context, toUser, fromUser string, request *int64, rows []*TransferredCards) (*Transfer, error)
	CloseTransfer(ctx context.Context, id int64) error

	GetUserByUsername(ctx context.Context, username string) (*User, error)
	AddUser(ctx context.Context, username, email string) (*User, error)
}
