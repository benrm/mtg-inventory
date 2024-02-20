package inventory

import "time"

// User represents a user in the users table
type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Card represents a Card
type Card struct {
	Name       string `json:"name"`
	OracleID   string `json:"oracle_id"`
	ScryfallID string `json:"scryfall_id"`
	Foil       bool   `json:"foil"`
}

// CardRow represents a row in the cards table
type CardRow struct {
	Quantity int    `json:"quantity"`
	Card     *Card  `json:"card"`
	Owner    string `json:"owner"`
	Keeper   string `json:"keeper"`
}

// Request represents a row in the requests table
type Request struct {
	ID        int64             `json:"id"`
	Requestor string            `json:"requestor"`
	Opened    time.Time         `json:"opened"`
	Closed    time.Time         `json:"closed"`
	Cards     []*RequestedCards `json:"cards"`
}

// RequestedCards represents a row in the requested_cards table
type RequestedCards struct {
	Quantity int    `json:"quantity"`
	Name     string `json:"name"`
	OracleID string `json:"oracle_id"`
}

// Transfer represents a row in the transfers table
type Transfer struct {
	ID        int64               `json:"id"`
	RequestID *int64              `json:"request_id"`
	ToUser    string              `json:"to_user"`
	FromUser  string              `json:"from_user"`
	Created   time.Time           `json:"created"`
	Executed  time.Time           `json:"executed"`
	Cards     []*TransferredCards `json:"cards"`
}

// TransferredCards represents a row in the transferred_cards table
type TransferredCards struct {
	Quantity int    `json:"quantity"`
	Card     *Card  `json:"card"`
	Owner    string `json:"owner"`
}

// HTTPError is the type used to marshal errors into JSON
type HTTPError struct {
	Error string `json:"error"`
}
