package sql

import "time"

type User struct {
	ID       int64
	Username string
	Email    string
}

type Card struct {
	OracleID   string
	ScryfallID string
	Foil       bool
}

type CardRow struct {
	Quantity int
	Card     *Card
	Owner    *User
	Keeper   *User
}

type Request struct {
	ID        int64
	Requestor *User
	Opened    time.Time
	Closed    time.Time
}

type RequestedCards struct {
	RequestID int64
	OracleID  string
	Quantity  int
}

type Transfer struct {
	ID        int64
	RequestID int64
	ToUser    *User
	FromUser  *User
	Created   time.Time
	Executed  time.Time
}

type TransferredCards struct {
	TransferID int64
	Quantity   int
	ScryfallID string
	Foil       bool
}
