package inventory

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

// ScryfallCardFace represents one of the faces of a ScryfallCard
type ScryfallCardFace struct {
	Name     string `json:"name"`
	OracleID string `json:"oracle_id"`
}

// ScryfallCard represents a card object retrieved from Scryfall
type ScryfallCard struct {
	// Core Card Fields
	ID       string
	Language string
	OracleID string

	// Gameplay fields
	Name string

	// Print fields
	CollectorNumber string
	ReleasedAt      time.Time
	Set             string

	// Card Face Objects
	CardFaces []ScryfallCardFace
}

// UnmarshalJSON implements json.Unmarshaler
func (sc *ScryfallCard) UnmarshalJSON(b []byte) error {
	s := struct {
		// Core Card Fields
		ID       string `json:"id"`
		Language string `json:"lang"`
		OracleID string `json:"oracle_id"`

		// Gameplay fields
		Name string `json:"name"`

		// Print fields
		CollectorNumber string `json:"collector_number"`
		ReleasedAt      string `json:"released_at"`
		Set             string `json:"set"`

		// Card Face Objects
		CardFaces []ScryfallCardFace `json:"card_faces"`
	}{}

	err := json.Unmarshal(b, &s)
	if err != nil {
		return fmt.Errorf("error unmarshaling Scryfall card into anonymous struct: %w", err)
	}

	t, err := time.Parse("2006-01-02", s.ReleasedAt)
	if err != nil {
		return fmt.Errorf("error parsing released date from Scryfall card: %w", err)
	}

	*sc = ScryfallCard{
		ID:              s.ID,
		Language:        s.Language,
		OracleID:        s.OracleID,
		Name:            s.Name,
		CollectorNumber: s.CollectorNumber,
		ReleasedAt:      t,
		Set:             s.Set,
		CardFaces:       s.CardFaces,
	}
	return nil
}

// ErrNotInScryfallCache should be returned when a card or cards is not in a
// ScryfallCache
var ErrNotInScryfallCache = errors.New("not in Scryfall cache")

// ScryfallCache represents a cache of Scryfall bulk data that can be used to
// look up cards
type ScryfallCache interface {
	GetCards(name, set, language, collectorNumber string) ([]*ScryfallCard, error)
	GetCardsByOracleID(string) ([]*ScryfallCard, error)
	GetCardByScryfallID(string) (*ScryfallCard, error)
}

type jsonScryfallCache struct {
	nameMap       map[string]map[string]map[string]map[string][]*ScryfallCard
	oracleIDMap   map[string][]*ScryfallCard
	scryfallIDMap map[string]*ScryfallCard
}

func NewScryfallCacheFromJSON(reader io.Reader) (ScryfallCache, error) {
	decoder := json.NewDecoder(reader)
	_, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("error reading first token: %w", err)
	}

	cache := &jsonScryfallCache{
		nameMap:       make(map[string]map[string]map[string]map[string][]*ScryfallCard),
		oracleIDMap:   make(map[string][]*ScryfallCard),
		scryfallIDMap: make(map[string]*ScryfallCard),
	}

	for decoder.More() {
		var card ScryfallCard
		err = decoder.Decode(&card)
		if err != nil {
			return nil, fmt.Errorf("error reading after %d bytes: %w", decoder.InputOffset(), err)
		}

		if _, exists := cache.nameMap[card.Name]; !exists {
			cache.nameMap[card.Name] = make(map[string]map[string]map[string][]*ScryfallCard)
		}
		if _, exists := cache.nameMap[card.Name][card.Set]; !exists {
			cache.nameMap[card.Name][card.Set] = make(map[string]map[string][]*ScryfallCard)
		}
		if _, exists := cache.nameMap[card.Name][card.Set][card.Language]; !exists {
			cache.nameMap[card.Name][card.Set][card.Language] = make(map[string][]*ScryfallCard)
		}
		if _, exists := cache.nameMap[card.Name][card.Set][card.Language][card.CollectorNumber]; !exists {
			cache.nameMap[card.Name][card.Set][card.Language][card.CollectorNumber] = make([]*ScryfallCard, 0)
		}
		cache.nameMap[card.Name][card.Set][card.Language][card.CollectorNumber] = append(cache.nameMap[card.Name][card.Set][card.Language][card.CollectorNumber], &card)

		if _, exists := cache.oracleIDMap[card.OracleID]; !exists {
			cache.oracleIDMap[card.OracleID] = make([]*ScryfallCard, 0)
		}
		cache.oracleIDMap[card.OracleID] = append(cache.oracleIDMap[card.OracleID], &card)

		cache.scryfallIDMap[card.ID] = &card
	}

	return cache, nil
}

func (jsc *jsonScryfallCache) GetCards(name, set, language, collectorNumber string) ([]*ScryfallCard, error) {
	if setMap, exists := jsc.nameMap[name]; exists {
		if languageMap, exists := setMap[set]; exists {
			if collectorNumberMap, exists := languageMap[language]; exists {
				if cardSlice, exists := collectorNumberMap[collectorNumber]; exists {
					return cardSlice, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("didn't find %q|%q|%q|%q: %w", name, set, language, collectorNumber, ErrNotInScryfallCache)
}

func (jsc *jsonScryfallCache) GetCardsByOracleID(oracleID string) ([]*ScryfallCard, error) {
	if oracleSlice, exists := jsc.oracleIDMap[oracleID]; exists {
		return oracleSlice, nil
	}
	return nil, fmt.Errorf("didn't find oracle ID %q: %w", oracleID, ErrNotInScryfallCache)
}

func (jsc *jsonScryfallCache) GetCardByScryfallID(scryfallID string) (*ScryfallCard, error) {
	if card, exists := jsc.scryfallIDMap[scryfallID]; exists {
		return card, nil
	}
	return nil, fmt.Errorf("didn't find scryfall ID %q: %w", scryfallID, ErrNotInScryfallCache)
}
