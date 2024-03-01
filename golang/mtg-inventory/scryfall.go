package inventory

import (
	"encoding/json"
	"fmt"
	"time"
)

// ScryfallDate represents a date as represented in Scryfall
type ScryfallDate struct {
	Value string
	Time  time.Time
}

// UnmarshalJSON implements json.Unmarshaler
func (sd *ScryfallDate) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return fmt.Errorf("error unmarshaling date %q into string: %w", b, err)
	}
	t, err := time.Parse(time.DateOnly, str)
	if err != nil {
		return fmt.Errorf("error parsing string %q as date: %w", str, err)
	}
	*sd = ScryfallDate{
		Value: str,
		Time:  t,
	}
	return nil
}

// ScryfallCardFace represents one of the faces of a Card
type ScryfallCardFace struct {
	Name     string `json:"name"`
	OracleID string `json:"oracle_id"`
}

// ScryfallCard represents a card object retrieved from Scryfall
type ScryfallCard struct {
	// Core Card Fields
	ID       string `json:"id"`
	Language string `json:"lang"`
	OracleID string `json:"oracle_id"`

	// Gameplay fields
	Name string `json:"name"`

	// Print fields
	CollectorNumber string       `json:"collector_number"`
	ReleasedAt      ScryfallDate `json:"released_at"`
	Set             string       `json:"set"`

	// Card Face Objects
	CardFaces []ScryfallCardFace `json:"card_faces"`
}

// Scryfall describes the interface with something that returns Scryfall data,
// whether it's a cache or the REST API.
type Scryfall interface {
	GetCard(name, set, language, collectorNumber string) (*ScryfallCard, error)
	GetCardByName(string) (*ScryfallCard, error)
	GetCardByOracleID(string) (*ScryfallCard, error)
	GetCardByID(string) (*ScryfallCard, error)
}
