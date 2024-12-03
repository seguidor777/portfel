package models

import (
	"fmt"
	"testing"
)

var assetWeights = map[string]float64{
	"BTCUSDT":  0.5,
	"ADAUSDT":  0.1,
	"ETHUSDT":  0.07,
	"SOLUSDT":  0.05,
	"BNBUSDT":  0.05,
	"XRPUSDT":  0.03,
	"DOTUSDT":  0.025,
	"UNIUSDT":  0.025,
	"AVAXUSDT": 0.025,
	"LINKUSDT": 0.025,
	"TRXUSDT":  0.02,
	"TONUSDT":  0.02,
	"HBARUSDT": 0.02,
	"XLMUSDT":  0.02,
	"SUIUSDT":  0.02,
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
