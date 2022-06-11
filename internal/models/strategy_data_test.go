package models

import (
	"fmt"
	"testing"
)

var assetWeights = map[string]float64{
	"BTCBUSD":   0.5,
	"ADABUSD":   0.1,
	"ETHBUSD":   0.07,
	"SOLBUSD":   0.05,
	"BNBBUSD":   0.05,
	"DOTBUSD":   0.03,
	"AVAXBUSD":  0.025,
	"LINKBUSD":  0.025,
	"FTMBUSD":   0.025,
	"MATICBUSD": 0.025,
	"ROSEBUSD":  0.02,
	"MANABUSD":  0.02,
	"SANDBUSD":  0.02,
	"GALABUSD":  0.02,
	"AUDIOBUSD": 0.02,
}

// TestGetSlugs tests getSlugs function
func TestGetSlugs(t *testing.T) {
	slugs, err := getSlugs(assetWeights)
	if err != nil {
		t.Fatal(err)
	}

	if len(slugs) != len(assetWeights) {
		t.Fatal("not all slugs were found")
	}
}

// TestGetSlugsErr tests getPriceDrop function and expects an error
func TestGetSlugsErr(t *testing.T) {
	delete(assetWeights, "BTCBUSD")
	assetWeights["@@@BUSD"] = 0.5
	_, err := getSlugs(assetWeights)

	if err == nil {
		t.Fatal(err)
	}

	if err.Error() != fmt.Sprintf(coinNotFoundErr, "@@@") {
		t.Fatal(err)
	}
}
