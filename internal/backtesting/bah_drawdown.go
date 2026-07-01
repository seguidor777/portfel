package backtesting

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"

	"github.com/seguidor777/portfel/internal/models"
)

// BuyAndHoldDrawdown computes the maximum drawdown of a static buy-and-hold
// portfolio built from the same CSV files used in the backtest.
//
// The portfolio is initialised at the first available candle: each pair
// receives (assetWeight * walletAmount) USDT worth of coins at that day's
// close price. From that point the portfolio value is tracked daily as the
// weighted sum of mark-to-market positions. The function returns the worst
// peak-to-trough percentage decline observed over the full period.
func BuyAndHoldDrawdown(config *models.Config, timeframe string, walletAmt float64) (float64, error) {
	// Map: pair -> (timestamp -> close price)
	priceSeries := make(map[string]map[int64]float64, len(config.AssetWeights))
	allTimestamps := make(map[int64]struct{})

	for pair := range config.AssetWeights {
		prices, err := readClosePrices(fmt.Sprintf("testdata/%s-%s.csv", pair, timeframe))
		if err != nil {
			return 0, fmt.Errorf("reading %s: %w", pair, err)
		}
		priceSeries[pair] = prices
		for ts := range prices {
			allTimestamps[ts] = struct{}{}
		}
	}

	// Sort timestamps ascending
	timestamps := make([]int64, 0, len(allTimestamps))
	for ts := range allTimestamps {
		timestamps = append(timestamps, ts)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })

	// Determine the first timestamp where ALL pairs have data (common start)
	firstCommon := int64(0)
	for _, ts := range timestamps {
		allPresent := true
		for pair := range config.AssetWeights {
			if _, ok := priceSeries[pair][ts]; !ok {
				allPresent = false
				break
			}
		}
		if allPresent {
			firstCommon = ts
			break
		}
	}
	if firstCommon == 0 {
		return 0, fmt.Errorf("no common start timestamp found across all pairs")
	}

	// Calculate initial holdings: how many units of each asset we buy on day 0
	holdings := make(map[string]float64, len(config.AssetWeights))
	for pair, weight := range config.AssetWeights {
		alloc := weight * walletAmt
		holdings[pair] = alloc / priceSeries[pair][firstCommon]
	}

	// Walk timestamps from firstCommon, compute portfolio value, track drawdown
	peak := 0.0
	maxDrawdown := 0.0

	for _, ts := range timestamps {
		if ts < firstCommon {
			continue
		}

		// Portfolio value at this timestamp (use last known price if a pair is missing a candle)
		portfolioVal := 0.0
		for pair, qty := range holdings {
			if price, ok := priceSeries[pair][ts]; ok {
				portfolioVal += qty * price
			}
		}
		if portfolioVal == 0 {
			continue
		}

		if portfolioVal > peak {
			peak = portfolioVal
		}

		drawdown := (portfolioVal - peak) / peak * 100
		if drawdown < maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return math.Abs(maxDrawdown), nil
}

// readClosePrices parses a ninjabot-style CSV (time,open,close,low,high,volume)
// and returns a map of unix-timestamp -> close price.
func readClosePrices(path string) (map[int64]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	prices := make(map[int64]float64, len(records))
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 3 {
			continue
		}
		ts, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			continue
		}
		close, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			continue
		}
		prices[ts] = close
	}
	return prices, nil
}
