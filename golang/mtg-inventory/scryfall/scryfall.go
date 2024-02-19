/*
Package scryfall contains code used to interact with Scryfall data, generally
through a local cache of the bulk data.
*/
package scryfall

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// Date represents a date as represented in Scryfall
type Date struct {
	Value string
	Time  time.Time
}

// UnmarshalJSON implements json.Unmarshaler
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

func isPreferredCard(a, b *Card) bool {
	// Prefer English
	if a.Language != b.Language && (b.Language == "en" || a.Language != "en") {
		return a.Language == "en"
	}
	// Prefer newer
	if !a.ReleasedAt.Time.Equal(b.ReleasedAt.Time) {
		return a.ReleasedAt.Time.After(b.ReleasedAt.Time)
	}
	// Prefer smaller collector number
	if a.CollectorNumber != b.CollectorNumber {
		return strings.Compare(a.CollectorNumber, b.CollectorNumber) < 0
	}
	// Prefer alphabetical by set
	if a.Set != b.Set {
		return strings.Compare(a.Set, b.Set) < 0
	}
	// Default first
	return true
}

func getPreferredCard(a, b *Card) *Card {
	if isPreferredCard(a, b) {
		return a
	}
	return b
}

var (
	// ErrNotInCache should be returned when a card or cards is not in a Cache
	ErrNotInCache = errors.New("not in Scryfall cache")

	// ErrMultipleCacheHits should be returned when a search would unavoidably return multiple results
	ErrMultipleCacheHits = errors.New("multiple Scryfall cache hits")
)

// Cache represents a cache of Scryfall bulk data that can be used to look up
// cards
type Cache interface {
	GetCard(name, set, language, collectorNumber string) (*Card, error)
	GetCardByName(string) (*Card, error)
	GetCardByOracleID(string) (*Card, error)
	GetCardByScryfallID(string) (*Card, error)
}

type cardsWithDefault struct {
	Default            *Card
	CollectorNumberMap map[string][]*Card
}

type cardKey struct {
	Name     string
	Set      string
	Language string
}

type jsonCache struct {
	KeyMap          map[cardKey]*cardsWithDefault
	OracleIDMap     map[string]*Card
	ScryfallIDMap   map[string]*Card
	NameToOracleMap map[string]map[string]*Card
}

// NewCacheFromJSON creates a Cache from Scryfall bulk JSON data
func NewCacheFromJSON(reader io.Reader) (Cache, error) {
	decoder := json.NewDecoder(reader)
	_, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("error reading first token: %w", err)
	}

	cache := &jsonCache{
		KeyMap:          make(map[cardKey]*cardsWithDefault),
		OracleIDMap:     make(map[string]*Card),
		ScryfallIDMap:   make(map[string]*Card),
		NameToOracleMap: make(map[string]map[string]*Card),
	}

	for decoder.More() {
		var card Card
		err = decoder.Decode(&card)
		if err != nil {
			return nil, fmt.Errorf("error reading after %d bytes: %w", decoder.InputOffset(), err)
		}

		key := cardKey{
			Name:     card.Name,
			Set:      card.Set,
			Language: card.Language,
		}
		if _, exists := cache.KeyMap[key]; !exists {
			cache.KeyMap[key] = &cardsWithDefault{
				Default:            &card,
				CollectorNumberMap: make(map[string][]*Card),
			}
		} else {
			cache.KeyMap[key].Default = getPreferredCard(cache.KeyMap[key].Default, &card)
		}
		if _, exists := cache.KeyMap[key].CollectorNumberMap[card.CollectorNumber]; !exists {
			cache.KeyMap[key].CollectorNumberMap[card.CollectorNumber] = make([]*Card, 0)
		}
		cache.KeyMap[key].CollectorNumberMap[card.CollectorNumber] = append(cache.KeyMap[key].CollectorNumberMap[card.CollectorNumber], &card)

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
		if current, exists := cache.OracleIDMap[oracleID]; !exists {
			cache.OracleIDMap[oracleID] = &card
		} else {
			cache.OracleIDMap[oracleID] = getPreferredCard(current, &card)
		}

		cache.ScryfallIDMap[card.ID] = &card

		if _, exists := cache.NameToOracleMap[card.Name]; !exists {
			cache.NameToOracleMap[card.Name] = make(map[string]*Card)
		}
		if current, exists := cache.NameToOracleMap[card.Name][card.OracleID]; !exists {
			cache.NameToOracleMap[card.Name][card.OracleID] = &card
		} else {
			cache.NameToOracleMap[card.Name][card.OracleID] = getPreferredCard(current, &card)
		}
	}

	return cache, nil
}

func (jc *jsonCache) GetCard(name, set, language, collectorNumber string) (*Card, error) {
	key := cardKey{
		Name:     name,
		Set:      set,
		Language: language,
	}
	if withDefault, exists := jc.KeyMap[key]; exists {
		if collectorNumber == "" {
			return withDefault.Default, nil
		}
		if cards, exists := jc.KeyMap[key].CollectorNumberMap[collectorNumber]; exists {
			return cards[0], nil
		}
	}
	return nil, fmt.Errorf("didn't find %q|%q|%q|%q: %w", name, set, language, collectorNumber, ErrNotInCache)
}

func (jc *jsonCache) GetCardByName(name string) (*Card, error) {
	if oracleMap, exists := jc.NameToOracleMap[name]; exists {
		if len(oracleMap) == 1 {
			for _, card := range oracleMap {
				return card, nil
			}
		} else if len(oracleMap) > 1 {
			return nil, fmt.Errorf("found multiple cards named %q: %w", name, ErrMultipleCacheHits)
		}
	}
	return nil, fmt.Errorf("didn't find name %q: %w", name, ErrNotInCache)
}

func (jc *jsonCache) GetCardByOracleID(oracleID string) (*Card, error) {
	if card, exists := jc.OracleIDMap[oracleID]; exists {
		return card, nil
	}
	return nil, fmt.Errorf("didn't find oracle ID %q: %w", oracleID, ErrNotInCache)
}

func (jc *jsonCache) GetCardByScryfallID(scryfallID string) (*Card, error) {
	if card, exists := jc.ScryfallIDMap[scryfallID]; exists {
		return card, nil
	}
	return nil, fmt.Errorf("didn't find scryfall ID %q: %w", scryfallID, ErrNotInCache)
}
