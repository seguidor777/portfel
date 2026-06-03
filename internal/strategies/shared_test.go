package strategies

import (
	"fmt"
	"testing"
)

// TestGetMarketData tests getMarketData function
func TestGetMarketData(t *testing.T) {
	slugs := []string{"avalanche-2", "binancecoin", "bitcoin", "cardano", "chainlink", "ethereum", "hedera-hashgraph", "polkadot", "ripple", "solana", "sui", "stellar", "the-open-network", "tron", "uniswap"}

	for _, slug := range slugs {
		_, err := getMarketData(slug)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestGetMarketDataErr tests getMarketData function and expects an error
func TestGetMarketDataErr(t *testing.T) {
	invalidSlug := "@@@"
	_, err := getMarketData(invalidSlug)
	if err == nil {
		t.Fatal(err)
	}

	if err.Error() != fmt.Sprintf(priceDropNotFoundErr, invalidSlug) {
		t.Fatal(err)
	}
}
