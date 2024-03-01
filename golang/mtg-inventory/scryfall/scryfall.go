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

	inventory "github.com/benrm/mtg-inventory/golang/mtg-inventory"
)

func isPreferredCard(a, b *inventory.ScryfallCard) bool {
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

func getPreferredCard(a, b *inventory.ScryfallCard) *inventory.ScryfallCard {
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

type cardsWithDefault struct {
	Default            *inventory.ScryfallCard
	CollectorNumberMap map[string][]*inventory.ScryfallCard
}

type cardKey struct {
	Name     string
	Set      string
	Language string
}

type jsonCache struct {
	KeyMap          map[cardKey]*cardsWithDefault
	OracleIDMap     map[string]*inventory.ScryfallCard
	ScryfallIDMap   map[string]*inventory.ScryfallCard
	NameToOracleMap map[string]map[string]*inventory.ScryfallCard
}

// NewCacheFromJSON creates a Cache from Scryfall bulk JSON data
func NewCacheFromJSON(reader io.Reader) (inventory.Scryfall, error) {
	decoder := json.NewDecoder(reader)
	_, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("error reading first token: %w", err)
	}

	cache := &jsonCache{
		KeyMap:          make(map[cardKey]*cardsWithDefault),
		OracleIDMap:     make(map[string]*inventory.ScryfallCard),
		ScryfallIDMap:   make(map[string]*inventory.ScryfallCard),
		NameToOracleMap: make(map[string]map[string]*inventory.ScryfallCard),
	}

	for decoder.More() {
		var card inventory.ScryfallCard
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
				CollectorNumberMap: make(map[string][]*inventory.ScryfallCard),
			}
		} else {
			cache.KeyMap[key].Default = getPreferredCard(cache.KeyMap[key].Default, &card)
		}
		if _, exists := cache.KeyMap[key].CollectorNumberMap[card.CollectorNumber]; !exists {
			cache.KeyMap[key].CollectorNumberMap[card.CollectorNumber] = make([]*inventory.ScryfallCard, 0)
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
			cache.NameToOracleMap[card.Name] = make(map[string]*inventory.ScryfallCard)
		}
		if current, exists := cache.NameToOracleMap[card.Name][card.OracleID]; !exists {
			cache.NameToOracleMap[card.Name][card.OracleID] = &card
		} else {
			cache.NameToOracleMap[card.Name][card.OracleID] = getPreferredCard(current, &card)
		}
	}

	return cache, nil
}

func (jc *jsonCache) GetCard(name, set, language, collectorNumber string) (*inventory.ScryfallCard, error) {
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

func (jc *jsonCache) GetCardByName(name string) (*inventory.ScryfallCard, error) {
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

func (jc *jsonCache) GetCardByOracleID(oracleID string) (*inventory.ScryfallCard, error) {
	if card, exists := jc.OracleIDMap[oracleID]; exists {
		return card, nil
	}
	return nil, fmt.Errorf("didn't find oracle ID %q: %w", oracleID, ErrNotInCache)
}

func (jc *jsonCache) GetCardByID(scryfallID string) (*inventory.ScryfallCard, error) {
	if card, exists := jc.ScryfallIDMap[scryfallID]; exists {
		return card, nil
	}
	return nil, fmt.Errorf("didn't find scryfall ID %q: %w", scryfallID, ErrNotInCache)
}
