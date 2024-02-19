package inventory

import "time"

// User represents a user in the users table
type User struct {
	ID       int64  `json:"-"`
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
	Quantity int   `json:"quantity"`
	Card     *Card `json:"card"`
	Owner    *User `json:"owner"`
	Keeper   *User `json:"keeper"`
}

// Request represents a row in the requests table
type Request struct {
	ID        int64     `json:"id"`
	Requestor *User     `json:"requestor"`
	Opened    time.Time `json:"opened"`
	Closed    time.Time `json:"closed"`
}

// RequestedCards represents a row in the requested_cards table
type RequestedCards struct {
	RequestID int64  `json:"request_id"`
	Name      string `json:"name"`
	OracleID  string `json:"oracle_id"`
	Quantity  int    `json:"quantity"`
}

// Transfer represents a row in the transfers table
type Transfer struct {
	ID        int64               `json:"id"`
	RequestID int64               `json:"request_id"`
	ToUser    *User               `json:"to_user"`
	FromUser  *User               `json:"from_user"`
	Created   time.Time           `json:"created"`
	Executed  time.Time           `json:"executed"`
	Cards     []*TransferredCards `json:"cards"`
}

// TransferredCards represents a row in the transferred_cards table
type TransferredCards struct {
	Quantity int   `json:"quantity"`
	Card     *Card `json:"card"`
	Owner    *User `json:"owner"`
}

// HTTPError is the type used to marshal errors into JSON
type HTTPError struct {
	Error string `json:"error"`
}
