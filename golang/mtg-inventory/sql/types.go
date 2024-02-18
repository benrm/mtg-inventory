package sql

import "time"

type User struct {
	ID       int64
	Username string
	Email    string
}

type Card struct {
	EnglishName string
	OracleID    string
	ScryfallID  string
	Foil        bool
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
	RequestID   int64
	EnglishName string
	OracleID    string
	Quantity    int
}

type Transfer struct {
	ID        int64
	RequestID int64
	ToUser    *User
	FromUser  *User
	Created   time.Time
	Executed  time.Time
	Cards     []*TransferRow
}

type TransferRow struct {
	Quantity int
	Card     *Card
	Owner    *User
}
