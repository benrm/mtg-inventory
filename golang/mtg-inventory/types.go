package inventory

import "time"

// User represents a user in the users table
type User struct {
	ID       int64
	Username string
	Email    string
}

// Card represents a Card
type Card struct {
	Name       string
	OracleID   string
	ScryfallID string
	Foil       bool
}

// CardRow represents a row in the cards table
type CardRow struct {
	Quantity int
	Card     *Card
	Owner    *User
	Keeper   *User
}

// Request represents a row in the requests table
type Request struct {
	ID        int64
	Requestor *User
	Opened    time.Time
	Closed    time.Time
}

// RequestedCards represents a row in the requested_cards table
type RequestedCards struct {
	RequestID int64
	Name      string
	OracleID  string
	Quantity  int
}

// Transfer represents a row in the transfers table
type Transfer struct {
	ID        int64
	RequestID int64
	ToUser    *User
	FromUser  *User
	Created   time.Time
	Executed  time.Time
	Cards     []*TransferredCards
}

// TransferredCards represents a row in the transferred_cards table
type TransferredCards struct {
	Quantity int
	Card     *Card
	Owner    *User
}
