package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	USDSymbol       = "USDT"
	coinNotFoundErr = "coin symbol \"%s\" not found"
	noCoinsListErr  = "no coins list in the JSON response"
)

var coingeckoSlugs = map[string]string{
	"AAVE": "aave",
	"ADA":  "cardano",
	"ARB":  "arbitrum",
	"AVAX": "avalanche",
	"BNB":  "binancecoin",
	"BTC":  "bitcoin",
	"DOT":  "polkadot",
	"ETH":  "ethereum",
	"LINK": "chainlink",
	"NEAR": "near",
	"OP":   "optimism",
	"POL":  "polygon",
	"SOL":  "solana",
	"SUI":  "sui",
	"UNI":  "uniswap",
	"XRP":  "ripple",
}

type StrategyData struct {
	MinimumBalance    float64
	ExpectedPriceDrop float64
	AssetWeights      map[string]float64
	LastClose         map[string]float64
	LastHigh          map[string]float64
	AssetStake        map[string]float64
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
		Volume:            make(map[string]float64),
		// Last ATH
		ATHTest: map[string]float64{
			"BTCUSDT":  112032.45,
			"ETHUSDT":  4867.51,
			"SOLUSDT":  293.31,
			"BNBUSDT":  794.18,
			"LINKUSDT": 52.7,
			"ARBUSDT":  2.4,
			"AAVEUSDT": 661.69,
			"ADAUSDT":  3.1065,
			"UNIUSDT":  44.92,
			"AVAXUSDT": 146.22,
			"POLUSDT":  2.92,
			"NEARUSDT": 20.42,
			"SUIUSDT":  10.0,
			"TONUSDT":  10.0,
			"XRPUSDT":  3.65,
			"DOTUSDT":  55.127,
		},
		Slugs: slugs,
	}, nil
}

// Get the slug (coin ID) values for each of the assets in the map
func getSlugs(assetWeights map[string]float64) (map[string]string, error) {
	resp, err := resty.New().R().
		Get("https://api.coingecko.com/api/v3/coins/list")
	if err != nil {
		return nil, err
	}
	var jsonResp []map[string]string
	json.Unmarshal(resp.Body(), &jsonResp)
	if len(jsonResp) == 0 {
		return nil, errors.New(noCoinsListErr)
	}

	slugs := make(map[string]string)

	for pair := range assetWeights {
		symbol := strings.Split(pair, USDSymbol)[0]

		if slug, ok := coingeckoSlugs[symbol]; ok {
			slugs[pair] = slug
		} else {
			return nil, fmt.Errorf(coinNotFoundErr, symbol)
		}
	}

	return slugs, nil
}
