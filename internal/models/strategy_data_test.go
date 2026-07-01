package models

import (
	"fmt"
	"testing"
)

var assetWeights = map[string]float64{
	"BTCUSDT": 0.29,
	"ETHUSDT": 0.25,
	"SOLUSDT": 0.08,
	"BNBUSDT": 0.05,
	"LINKUSDT": 0.05,
	"ARBUSDT": 0.05,
	"AAVEUSDT": 0.04,
	"OPUSDT": 0.03,
	"UNIUSDT": 0.03,
	"AVAXUSDT": 0.03,
	"NEARUSDT": 0.02,
	"SUIUSDT": 0.02,
	"XRPUSDT": 0.02,
	"ADAUSDT": 0.02,
	"DOTUSDT": 0.02,
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
	delete(assetWeights, "BTCUSDT")
	assetWeights["@@@USDT"] = 0.5
	_, err := getSlugs(assetWeights)

	if err == nil {
		t.Fatal(err)
	}

	if err.Error() != fmt.Sprintf(coinNotFoundErr, "@@@") {
		t.Fatal(err)
	}
}
