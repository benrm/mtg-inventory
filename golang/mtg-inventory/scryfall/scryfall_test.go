package scryfall

import (
	"errors"
	"os"
	"testing"
)

func TestJSONCache(t *testing.T) {
	scryfallBulkData, err := os.Open("./testdata/all-cards.json")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skip("Download Scryfall bulk data and install in ./testdata/all-cards.json for full testing")
		} else {
			t.Fatalf("Error opening Scryfall bulk data file: %s", err.Error())
		}
	}

	cache, err := NewCacheFromJSON(scryfallBulkData)
	if err != nil {
		t.Fatalf("Error loading JSON cache: %s", err.Error())
	}

	_, err = cache.GetCards("Primeval Titan", "mm2", "en", "156")
	if err != nil {
		t.Fatalf("Error retrieving 'Primeval Titan' from JSON cache: %s", err.Error())
	}
}
