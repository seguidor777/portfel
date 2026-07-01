package strategies

import (
	"testing"

	"github.com/seguidor777/portfel/internal/localkv"
)

var databasePath = "../../user_data"

// TestNewMarketDataFetcher tests the MarketDataFetcher with real API and KV cache
func TestNewMarketDataFetcher(t *testing.T) {
	kv, err := localkv.NewLocalKV(&databasePath)
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()
	defer kv.RemoveDB()

	slugs := map[string]string{
		"AAVEUSDT": "aave",
		"ADAUSDT":  "cardano",
		"ARBUSDT":  "arbitrum",
		"AVAXUSDT": "avalanche-2",
		"BNBUSDT":  "binancecoin",
		"BTCUSDT":  "bitcoin",
		"DOTUSDT":  "polkadot",
		"ETHUSDT":  "ethereum",
		"LINKUSDT": "chainlink",
		"NEARUSDT": "near",
		"OPUSDT":   "optimism",
		"SOLUSDT":  "solana",
		"SUIUSDT":  "sui",
		"UNIUSDT":  "uniswap",
		"XRPUSDT":  "ripple",
	}

	fetcher := NewMarketDataFetcher(kv, slugs)

	// First call — fetches from API and caches
	data, err := fetcher("bitcoin")
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := data["ath"]; !ok {
		t.Fatal("expected 'ath' field in market data")
	}

	// Second call — should be served from KV cache
	data2, err := fetcher("ethereum")
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := data2["ath"]; !ok {
		t.Fatal("expected 'ath' field in cached market data")
	}
}

// TestNewMarketDataFetcherErr tests the MarketDataFetcher with an invalid slug
func TestNewMarketDataFetcherErr(t *testing.T) {
	kv, err := localkv.NewLocalKV(&databasePath)
	if err != nil {
		t.Fatal(err)
	}
	defer kv.Close()
	defer kv.RemoveDB()

	slugs := map[string]string{
		"INVALID": "@@@",
	}

	fetcher := NewMarketDataFetcher(kv, slugs)

	_, err = fetcher("@@@")
	if err == nil {
		t.Fatal("expected an error for invalid slug")
	}

	if err.Error() != noMarketDataErr {
		t.Fatalf("Expected error '%s' but got '%s'", noMarketDataErr, err.Error())
	}
}
