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

	cache, err := NewJSONCache(scryfallBulkData)
	if err != nil {
		t.Fatalf("Error loading JSON cache: %s", err.Error())
	}

	_, err = cache.GetCard("Primeval Titan", "mm2", "en", "")
	if err != nil {
		t.Fatalf("Error retrieving 'Primeval Titan' from JSON cache with name, set, and language: %s", err.Error())
	}

	_, err = cache.GetCard("Primeval Titan", "mm2", "en", "156")
	if err != nil {
		t.Fatalf("Error retrieving 'Primeval Titan' from JSON cache with name, set, language, and collector number: %s", err.Error())
	}

	_, err = cache.GetCardByName("Primeval Titan")
	if err != nil {
		t.Fatalf("Error retrieving 'Primeval Titan' from JSON cache with name: %s", err.Error())
	}

	_, err = cache.GetCardByName("Very Cryptic Command")
	if err == nil {
		t.Fatalf("No error retrieving 'Very Cryptic Command' from JSON cache with name")
	} else if !errors.Is(err, ErrMultipleCacheHits) {
		t.Fatalf("Unexpected error retrieving 'Very Cryptic Command' from JSON cache with name: %s", err.Error())
	}

	_, err = cache.GetCardByOracleID("ae83ef2c-960f-4c5b-97cc-52465c687c18")
	if err != nil {
		t.Fatalf("Error retrieving 'Primeval Titan' from JSON cache with oracle ID: %s", err.Error())
	}

	_, err = cache.GetCardByID("eea2bf31-4320-4605-ab5b-6b32472b82fa")
	if err != nil {
		t.Fatalf("Error retrieving 'Primeval Titan' from JSON cache with Scryfall ID: %s", err.Error())
	}
}
