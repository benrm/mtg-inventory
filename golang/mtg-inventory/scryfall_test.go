package inventory

import (
	"errors"
	"os"
	"testing"
)

func TestJSONScryfallCache(t *testing.T) {
	scryfallBulkData, err := os.Open("./testdata/all-cards.json")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skip("Download Scryfall bulk data and install in ./testdata/all-cards.json for full testing")
		} else {
			t.Fatalf("Error opening Scryfall bulk data file: %s", err.Error())
		}
	}

	scryfallCache, err := NewScryfallCacheFromJSON(scryfallBulkData)
	if err != nil {
		t.Fatalf("Error loading JSON Scryfall cache: %s", err.Error())
	}

	_, err = scryfallCache.GetCards("Primeval Titan", "mm2", "en", "156")
	if err != nil {
		t.Fatalf("Error retrieving 'Primeval Titan' from JSON Scryfall cache: %s", err.Error())
	}
}
