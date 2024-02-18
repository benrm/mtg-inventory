package inventory

import "time"

type User struct {
	ID       int
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
	ID        int
	Requestor *User
	Opened    time.Time
	Closed    time.Time
}

type RequestedCards struct {
	RequestID int
	OracleID  string
	Quantity  int
}

type Transfer struct {
	ID        int
	RequestID int
	ToUser    *User
	FromUser  *User
	Created   time.Time
	Executed  time.Time
}

type TransferredCards struct {
	TransferID int
	Quantity   int
	ScryfallID string
	Foil       bool
}
