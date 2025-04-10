package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strings"
)

const (
	USDSymbol       = "USDT"
	coinNotFoundErr = "coin symbol \"%s\" not found"
	noCoinsListErr  = "no coins list in the JSON response"
)

var coingeckoSlugs = map[string]string{
	"ADA":  "cardano",
	"AVAX": "avalanche-2",
	"BNB":  "binancecoin",
	"BTC":  "bitcoin",
	"DOT":  "polkadot",
	"ETH":  "ethereum",
	"HBAR": "hedera-hashgraph",
	"LINK": "chainlink",
	"SOL":  "solana",
	"SUI":  "sui",
	"TON":  "the-open-network",
	"TRX":  "tron",
	"UNI":  "uniswap",
	"XLM":  "stellar",
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
			"BTCUSDT":  68972.0,
			"ADAUSDT":  3.1016,
			"ETHUSDT":  4886.0,
			"SOLUSDT":  259.0,
			"BNBUSDT":  692.2,
			"XRPUSDT":  1.9706,
			"DOTUSDT":  55.0,
			"UNIUSDT":  44.357,
			"AVAXUSDT": 146.76,
			"LINKUSDT": 53.08,
			"TRXUSDT":  0.1803,
			"TONUSDT":  10.0,
			"HBARUSDT": 0.57512,
			"XLMUSDT":  0.797,
			"SUIUSDT":  10.0,
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
