package strategies

import (
	"fmt"
	"testing"
)

// TestGetPriceDrop tests getPriceDrop function
func TestGetPriceDrop(t *testing.T) {
	slugs := []string{"avalanche-2", "binancecoin", "bitcoin", "cardano", "chainlink", "ethereum", "hedera-hashgraph", "polkadot", "ripple", "solana", "sui", "stellar", "the-open-network", "tron", "uniswap"}

	for _, slug := range slugs {
		_, err := getPriceDrop(slug)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestGetPriceDropErr tests getPriceDrop function and expects an error
func TestGetPriceDropErr(t *testing.T) {
	invalidSlug := "@@@"
	_, err := getPriceDrop(invalidSlug)
	if err == nil {
		t.Fatal(err)
	}

	if err.Error() != fmt.Sprintf(priceDropNotFoundErr, invalidSlug) {
		t.Fatal(err)
	}
}
