package strategies

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/seguidor777/portfel/internal/localkv"
)

const (
	athKVKey        = "ath"
	noMarketDataErr = "no market data in the JSON response"
	notFoundErr     = "not found"
)

var counter uint32

// MarketDataFetcher fetches market data for a single slug, using the KV cache
// backed by the full Slugs set.
type MarketDataFetcher func(slug string) (map[string]float64, error)

// Verifies if given day is within the provided days
func dayIn(day int, days []int) bool {
	for _, d := range days {
		if day == d {
			return true
		}
	}

	return false
}

// buildCacheKey builds the composite cache key from sorted slugs and today's date.
// Format: "aave,avalanche-2,binancecoin-06-17-2026"
func buildCacheKey(slugs []string) string {
	sort.Strings(slugs)

	return fmt.Sprintf("%s-%s", strings.Join(slugs, ","), time.Now().Format("01-02-2006"))
}

// NewMarketDataFetcher creates a MarketDataFetcher that caches market data
// in the KV store. All slugs are fetched in a single API call on the first
// request of the day, and subsequent requests are served from cache.
//
// KV structure:
//
//	Key: "ath"
//	Value: {
//	  "aave,arbitrum,...,uniswap-06-17-2026": {
//	    "bitcoin": {"ath": 69044.77, "ath_change_percentage": -42.5, ...},
//	    "ethereum": {"ath": 4891.70, ...},
//	    ...
//	  }
//	}
func NewMarketDataFetcher(kv *localkv.LocalKV, slugs map[string]string) MarketDataFetcher {
	return func(slug string) (map[string]float64, error) {
		allSlugs := make([]string, 0, len(slugs))
		for _, s := range slugs {
			allSlugs = append(allSlugs, s)
		}

		cacheKey := buildCacheKey(allSlugs)

		// Try to read from KV cache
		raw, err := kv.Get(athKVKey)
		if err == nil {
			var cached map[string]map[string]map[string]float64
			if json.Unmarshal([]byte(raw), &cached) == nil {
				if entry, ok := cached[cacheKey]; ok {
					if data, ok := entry[slug]; ok {
						return data, nil
					}
				}
			}
		}

		// Cache miss — fetch all slugs from CoinGecko in a single API call
		resp, err := resty.New().R().
			SetHeader("x-cg-demo-api-key", os.Getenv("COINGECKO_API_KEY")).
			Get(fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=%s", strings.Join(allSlugs, ",")))
		if err != nil {
			return nil, err
		}

		var jsonResp []map[string]any
		json.Unmarshal(resp.Body(), &jsonResp)
		if len(jsonResp) == 0 {
			return nil, errors.New(noMarketDataErr)
		}

		// Build the per-slug map
		marketData := make(map[string]map[string]float64, len(jsonResp))
		for _, coin := range jsonResp {
			id, _ := coin["id"].(string)
			marketData[id] = map[string]float64{
				"ath":                   coin["ath"].(float64),
				"ath_change_percentage": coin["ath_change_percentage"].(float64),
			}
		}

		// Store in KV under the "ath" key with only the current cache key
		// (replaces any stale entries from previous days)
		toStore := map[string]map[string]map[string]float64{
			cacheKey: marketData,
		}

		encoded, err := json.Marshal(toStore)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal market data for KV: %w", err)
		}

		if err := kv.Set(athKVKey, string(encoded)); err != nil {
			return nil, fmt.Errorf("cannot store market data in KV: %w", err)
		}

		// Return the requested market data
		data, ok := marketData[slug]
		if !ok {
			return nil, errors.New(notFoundErr)
		}

		return data, nil
	}
}
