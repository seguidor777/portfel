package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/resty.v0"
	"strings"
)

const (
	USDSymbol       = "BUSD"
	coinNotFoundErr = "coin symbol \"%s\" not found"
	noCoinsListErr  = "no coins list in the JSON response"
)

var slugsBlacklist = [...]string{"san-diego-coin"}

type StrategyData struct {
	MinimumBalance    float64
	ExpectedPriceDrop float64
	AssetWeights      map[string]float64
	LastClose         map[string]float64
	LastHigh          map[string]float64
	AssetStake        map[string]float64
	Accumulation      map[string]float64
	Volume            map[string]float64
	ATHTest           map[string]float64
	Slugs             map[string]string
}

func NewStrategyData(config *Config) (*StrategyData, error) {
	slugs, err := getSlugs(config.AssetWeights)
	if err != nil {
		return nil, err
	}

	return &StrategyData{
		MinimumBalance:    config.MinimumBalance,
		ExpectedPriceDrop: config.ExpectedPriceDrop,
		AssetWeights:      config.AssetWeights,
		LastClose:         make(map[string]float64),
		LastHigh:          make(map[string]float64),
		AssetStake:        make(map[string]float64),
		Accumulation:      make(map[string]float64),
		Volume:            make(map[string]float64),
		ATHTest: map[string]float64{
			"BTCBUSD":   64637.0,
			"ADABUSD":   2.4728,
			"ETHBUSD":   4372.0,
			"SOLBUSD":   57.74,
			"BNBBUSD":   691.9,
			"DOTBUSD":   49.78,
			"AVAXBUSD":  59.74,
			"LINKBUSD":  53.08,
			"FTMBUSD":   0,
			"MATICBUSD": 2.695,
			"ROSEBUSD":  0.24855,
			"MANABUSD":  1.658,
			"SANDBUSD":  0.8936,
			"GALABUSD":  0,
			"AUDIOBUSD": 4.996,
		},
		// ATH on apr 14th
		// ATHTest: map[string]float64{
		//	"BTCBUSD":   68972.0,
		//	"ADABUSD":   3.1016,
		//	"ETHBUSD":   4886.0,
		//	"SOLBUSD":   259.0,
		//	"BNBBUSD":   668.2,
		//	"DOTBUSD":   55.0,
		//	"AVAXBUSD":  146.76,
		//	"LINKBUSD":  53.08,
		//	"FTMBUSD":   3.4937,
		//	"MATICBUSD": 2.918,
		//	"ROSEBUSD":  0.5981,
		//	"MANABUSD":  5.9118,
		//	"SANDBUSD":  8.4765,
		//	"GALABUSD":  0.84222,
		//	"AUDIOBUSD": 4.996,
		//},
		Slugs: slugs,
	}, nil
}

// Get the slug (coin ID) values for each of the assets in the map
func getSlugs(assetWeights map[string]float64) (map[string]string, error) {
	resp, err := resty.R().
		Get("https://api.coingecko.com/api/v3/coins/list")
	if err != nil {
		return nil, err
	}
	var jsonResp []map[string]string
	json.Unmarshal(resp.Body, &jsonResp)
	if len(jsonResp) == 0 {
		return nil, errors.New(noCoinsListErr)
	}

	slugs := make(map[string]string)

	for pair := range assetWeights {
		symbol := strings.Split(pair, USDSymbol)[0]
		slug, err := getSlug(strings.ToLower(symbol), jsonResp)
		if err != nil {
			return slugs, err
		}

		slugs[pair] = slug
	}

	return slugs, nil
}

// Gets the slug for the given symbol
// Discards any slug if it has the binance-peg prefix or is blacklisted
func getSlug(symbol string, jsonResp []map[string]string) (string, error) {
	for _, coin := range jsonResp {
		if symbol == coin["symbol"] && !strings.HasPrefix(coin["id"], "binance-peg") && !isBlacklisted(coin["id"]) {
			return coin["id"], nil
		}
	}

	return "", fmt.Errorf(coinNotFoundErr, symbol)
}

// Verifies if a given slug is blacklisted
func isBlacklisted(slug string) bool {
	for _, s := range slugsBlacklist {
		if slug == s {
			return true
		}
	}

	return false
}
