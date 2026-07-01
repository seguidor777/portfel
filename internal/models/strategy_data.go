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
	"AVAX": "avalanche-2",
	"BNB":  "binancecoin",
	"BTC":  "bitcoin",
	"DOT":  "polkadot",
	"ETH":  "ethereum",
	"LINK": "chainlink",
	"NEAR": "near",
	"OP":   "optimism",
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
	SellProceeds      map[string]float64 // cumulative USDT received from ATH sells
	ATHTest           map[string]float64
	Slugs             map[string]string
	Metadata          map[string]map[string]float64
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
		SellProceeds:      make(map[string]float64),
		// ATH as of 2022-06-14 (start of the backtest window).
		// Coins from the 2021 bull run use their actual all-time highs.
		// ARB (launched Mar 2023) and SUI (launched May 2023) use their
		// first-day candle high since they had no ATH before this date.
		ATHTest: map[string]float64{
			"BTCUSDT":  69044.77, // Nov 10, 2021
			"ETHUSDT":  4891.70,  // Nov 10, 2021
			"SOLUSDT":  260.06,   // Nov 6, 2021
			"BNBUSDT":  686.31,   // May 10, 2021
			"LINKUSDT": 52.70,    // Aug 16, 2020
			"ARBUSDT":  1.60,     // first available candle (launched Mar 2023)
			"AAVEUSDT": 661.69,   // May 18, 2021
			"ADAUSDT":  3.10,     // Sep 2, 2021
			"UNIUSDT":  44.92,    // May 3, 2021
			"AVAXUSDT": 146.22,   // Nov 21, 2021
			"NEARUSDT": 20.42,    // Jan 16, 2022
			"SUIUSDT":  2.00,     // first available candle (launched May 2023)
			"XRPUSDT":  3.84,     // Jan 7, 2018
			"DOTUSDT":  55.09,    // Nov 4, 2021
			"OPUSDT":   10.0,     // placeholder (OP data not in this window)
		},
		Slugs:    slugs,
		Metadata: make(map[string]map[string]float64),
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
