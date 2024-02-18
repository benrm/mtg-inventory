package scryfall

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

// Date represents a date as represented in Scryfall
type Date struct {
	Value string
	Time  time.Time
}

func (sd *Date) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return fmt.Errorf("error unmarshaling date %q into string: %w", b, err)
	}
	t, err := time.Parse(time.DateOnly, str)
	if err != nil {
		return fmt.Errorf("error parsing string %q as date: %w", str, err)
	}
	*sd = Date{
		Value: str,
		Time:  t,
	}
	return nil
}

// CardFace represents one of the faces of a Card
type CardFace struct {
	Name     string `json:"name"`
	OracleID string `json:"oracle_id"`
}

// Card represents a card object retrieved from Scryfall
type Card struct {
	// Core Card Fields
	ID       string `json:"id"`
	Language string `json:"lang"`
	OracleID string `json:"oracle_id"`

	// Gameplay fields
	Name string `json:"name"`

	// Print fields
	CollectorNumber string `json:"collector_number"`
	ReleasedAt      Date   `json:"released_at"`
	Set             string `json:"set"`

	// Card Face Objects
	CardFaces []CardFace `json:"card_faces"`
}

// ErrNotInCache should be returned when a card or cards is not in a Cache
var ErrNotInCache = errors.New("not in Scryfall cache")

// Cache represents a cache of Scryfall bulk data that can be used to look up
// cards
type Cache interface {
	GetCards(name, set, language, collectorNumber string) ([]*Card, error)
	GetCardsByOracleID(string) ([]*Card, error)
	GetCardByScryfallID(string) (*Card, error)
}

type jsonCache struct {
	nameMap       map[string]map[string]map[string]map[string][]*Card
	oracleIDMap   map[string][]*Card
	scryfallIDMap map[string]*Card
}

func NewCacheFromJSON(reader io.Reader) (Cache, error) {
	decoder := json.NewDecoder(reader)
	_, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("error reading first token: %w", err)
	}

	cache := &jsonCache{
		nameMap:       make(map[string]map[string]map[string]map[string][]*Card),
		oracleIDMap:   make(map[string][]*Card),
		scryfallIDMap: make(map[string]*Card),
	}

	for decoder.More() {
		var card Card
		err = decoder.Decode(&card)
		if err != nil {
			return nil, fmt.Errorf("error reading after %d bytes: %w", decoder.InputOffset(), err)
		}

		if _, exists := cache.nameMap[card.Name]; !exists {
			cache.nameMap[card.Name] = make(map[string]map[string]map[string][]*Card)
		}
		if _, exists := cache.nameMap[card.Name][card.Set]; !exists {
			cache.nameMap[card.Name][card.Set] = make(map[string]map[string][]*Card)
		}
		if _, exists := cache.nameMap[card.Name][card.Set][card.Language]; !exists {
			cache.nameMap[card.Name][card.Set][card.Language] = make(map[string][]*Card)
		}
		if _, exists := cache.nameMap[card.Name][card.Set][card.Language][card.CollectorNumber]; !exists {
			cache.nameMap[card.Name][card.Set][card.Language][card.CollectorNumber] = make([]*Card, 0)
		}
		cache.nameMap[card.Name][card.Set][card.Language][card.CollectorNumber] = append(cache.nameMap[card.Name][card.Set][card.Language][card.CollectorNumber], &card)

		var oracleID string
		if card.OracleID == "" {
			if len(card.CardFaces) > 0 {
				oracleID = card.CardFaces[0].OracleID
			} else {
				return nil, fmt.Errorf("card with empty oracle ID after %d bytes", decoder.InputOffset())
			}
		} else {
			oracleID = card.OracleID
		}
		if _, exists := cache.oracleIDMap[oracleID]; !exists {
			cache.oracleIDMap[oracleID] = make([]*Card, 0)
		}
		cache.oracleIDMap[oracleID] = append(cache.oracleIDMap[oracleID], &card)

		cache.scryfallIDMap[card.ID] = &card
	}

	return cache, nil
}

func (jc *jsonCache) GetCards(name, set, language, collectorNumber string) ([]*Card, error) {
	if setMap, exists := jc.nameMap[name]; exists {
		if languageMap, exists := setMap[set]; exists {
			if collectorNumberMap, exists := languageMap[language]; exists {
				if cardSlice, exists := collectorNumberMap[collectorNumber]; exists {
					return cardSlice, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("didn't find %q|%q|%q|%q: %w", name, set, language, collectorNumber, ErrNotInCache)
}

func (jc *jsonCache) GetCardsByOracleID(oracleID string) ([]*Card, error) {
	if oracleSlice, exists := jc.oracleIDMap[oracleID]; exists {
		return oracleSlice, nil
	}
	return nil, fmt.Errorf("didn't find oracle ID %q: %w", oracleID, ErrNotInCache)
}

func (jc *jsonCache) GetCardByScryfallID(scryfallID string) (*Card, error) {
	if card, exists := jc.scryfallIDMap[scryfallID]; exists {
		return card, nil
	}
	return nil, fmt.Errorf("didn't find scryfall ID %q: %w", scryfallID, ErrNotInCache)
}
