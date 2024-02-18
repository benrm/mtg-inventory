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

func getPreferredCard(cards []*Card) *Card {
	if len(cards) == 0 {
		return nil
	}
	preferred := cards[0]
	for _, current := range cards[1:] {
		if current.Language == "en" && preferred.Language != "en" {
			preferred = current
		} else if current.ReleasedAt.Time.After(preferred.ReleasedAt.Time) {
			preferred = current
		} else if strings.Compare(current.CollectorNumber, preferred.CollectorNumber) < 0 {
			preferred = current
		}
	}
	return preferred
}

// ErrNotInCache should be returned when a card or cards is not in a Cache
var ErrNotInCache = errors.New("not in Scryfall cache")

// Cache represents a cache of Scryfall bulk data that can be used to look up
// cards
type Cache interface {
	GetCard(name, set, language, collectorNumber string) (*Card, error)
	GetCardByOracleID(string) (*Card, error)
	GetCardByScryfallID(string) (*Card, error)
}

type cardsWithDefault struct {
	Default *Card
	All     []*Card
}

type jsonCache struct {
	ByNameBySetByLang map[string]map[string]map[string]*cardsWithDefault
	OracleIDMap       map[string]*cardsWithDefault
	ScryfallIDMap     map[string]*Card
}

// NewCacheFromJSON creates a Cache from Scryfall bulk JSON data
func NewCacheFromJSON(reader io.Reader) (Cache, error) {
	decoder := json.NewDecoder(reader)
	_, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("error reading first token: %w", err)
	}

	cache := &jsonCache{
		ByNameBySetByLang: make(map[string]map[string]map[string]*cardsWithDefault),
		OracleIDMap:       make(map[string]*cardsWithDefault),
		ScryfallIDMap:     make(map[string]*Card),
	}

	for decoder.More() {
		var card Card
		err = decoder.Decode(&card)
		if err != nil {
			return nil, fmt.Errorf("error reading after %d bytes: %w", decoder.InputOffset(), err)
		}

		if _, exists := cache.ByNameBySetByLang[card.Name]; !exists {
			cache.ByNameBySetByLang[card.Name] = make(map[string]map[string]*cardsWithDefault)
		}
		if _, exists := cache.ByNameBySetByLang[card.Set]; !exists {
			cache.ByNameBySetByLang[card.Name][card.Set] = make(map[string]*cardsWithDefault)
		}
		if _, exists := cache.ByNameBySetByLang[card.Name][card.Set][card.Language]; !exists {
			cache.ByNameBySetByLang[card.Name][card.Set][card.Language] = &cardsWithDefault{
				All: make([]*Card, 0),
			}
		}
		cache.ByNameBySetByLang[card.Name][card.Set][card.Language].All = append(cache.ByNameBySetByLang[card.Name][card.Set][card.Language].All, &card)

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
		if _, exists := cache.OracleIDMap[oracleID]; !exists {
			cache.OracleIDMap[oracleID] = &cardsWithDefault{
				All: make([]*Card, 0),
			}
		}
		cache.OracleIDMap[oracleID].All = append(cache.OracleIDMap[oracleID].All, &card)

		cache.ScryfallIDMap[card.ID] = &card
	}

	for _, bySetByLang := range cache.ByNameBySetByLang {
		for _, byLang := range bySetByLang {
			for _, withDefault := range byLang {
				withDefault.Default = getPreferredCard(withDefault.All)
			}
		}
	}

	for _, byOracleID := range cache.OracleIDMap {
		byOracleID.Default = getPreferredCard(byOracleID.All)
	}

	return cache, nil
}

func (jc *jsonCache) GetCard(name, set, language, collectorNumber string) (*Card, error) {
	if bySetByLang, exists := jc.ByNameBySetByLang[name]; exists {
		if byLang, exists := bySetByLang[set]; exists {
			if withDefault, exists := byLang[language]; exists {
				if collectorNumber == "" {
					return withDefault.Default, nil
				}
				for _, card := range withDefault.All {
					if collectorNumber == card.CollectorNumber {
						return card, nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("didn't find %q|%q|%q|%q: %w", name, set, language, collectorNumber, ErrNotInCache)
}

func (jc *jsonCache) GetCardByOracleID(oracleID string) (*Card, error) {
	if oracleSlice, exists := jc.OracleIDMap[oracleID]; exists {
		return oracleSlice.Default, nil
	}
	return nil, fmt.Errorf("didn't find oracle ID %q: %w", oracleID, ErrNotInCache)
}

func (jc *jsonCache) GetCardByScryfallID(scryfallID string) (*Card, error) {
	if card, exists := jc.ScryfallIDMap[scryfallID]; exists {
		return card, nil
	}
	return nil, fmt.Errorf("didn't find scryfall ID %q: %w", scryfallID, ErrNotInCache)
}
