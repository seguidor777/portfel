package strategies

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"os"
)

const (
	noMarketDataErr      = "no market data in the JSON response"
	notFoundErr          = "not found"
	priceDropNotFoundErr = "price drop not found for %s"
)

var counter uint32

// Verifies if given day is within the provided days
func dayIn(day int, days []int) bool {
	for _, d := range days {
		if day == d {
			return true
		}
	}

	return false
}

// getMarketData fetches the market_data block for the given coin slug from CoinGecko
func getMarketData(slug string) (map[string]interface{}, error) {
	resp, err := resty.New().R().
		SetHeader("x-cg-demo-api-key", os.Getenv("COINGECKO_API_KEY")).
		Get(fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s?localization=false&tickers=false&community_data=false&developer_data=false&sparkline=false", slug))
	if err != nil {
		return nil, err
	}

	var jsonResp map[string]interface{}
	json.Unmarshal(resp.Body(), &jsonResp)
	if len(jsonResp) == 0 {
		return nil, errors.New(noMarketDataErr)
	}

	marketData := jsonResp["market_data"]
	if marketData == nil {
		return nil, fmt.Errorf(priceDropNotFoundErr, slug)
	}

	return marketData.(map[string]interface{}), nil
}

