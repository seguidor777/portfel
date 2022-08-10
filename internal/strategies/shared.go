package strategies

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
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

// Gets the price drop since the ATH
func getPriceDrop(slug string) (float64, error) {
	resp, err := resty.New().R().
		Get(fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s?localization=false&tickers=false&community_data=false&developer_data=false&sparkline=false", slug))
	if err != nil {
		return 0.0, err
	}

	var jsonResp map[string]interface{}
	json.Unmarshal(resp.Body(), &jsonResp)
	if len(jsonResp) == 0 {
		return 0.0, errors.New(noMarketDataErr)
	}

	marketData := jsonResp["market_data"]
	if marketData == nil {
		return 0.0, fmt.Errorf(priceDropNotFoundErr, slug)
	}

	return marketData.(map[string]interface{})["ath_change_percentage"].(map[string]interface{})["usd"].(float64), nil
}
