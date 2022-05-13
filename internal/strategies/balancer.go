package strategies

import (
	"math"

	"github.com/rodrigo-brito/ninjabot"
	"github.com/rodrigo-brito/ninjabot/model"
	"github.com/rodrigo-brito/ninjabot/service"
	log "github.com/sirupsen/logrus"
)

type Balancer struct {
	MinimumBalance float64
	AssetWeights   map[string]float64 `json:"asset_weight,omitempty"`
	LastClose      map[string]float64
}

type Weight struct {
	Pair   string
	Weight float64
}

func NewBalancer(minimumBalance float64, assetWeights map[string]float64) *Balancer {
	b := &Balancer{
		MinimumBalance: minimumBalance,
		AssetWeights:   assetWeights,
		LastClose:      make(map[string]float64),
	}

	return b
}

func (b Balancer) Timeframe() string {
	return "1d"
}

func (b Balancer) WarmupPeriod() int {
	return 1
}

func (b Balancer) Indicators(df *model.Dataframe) {
	b.LastClose[df.Pair] = df.Close.Last(0)
}

func (b Balancer) CalculatePositionAdjustment(df *ninjabot.Dataframe, broker service.Broker) (expect, diff float64, err error) {
	totalEquity := 0.0

	for p, _ := range b.AssetWeights {
		asset, _, err := broker.Position(p)
		if err != nil {
			return 0, 0, err
		}

		totalEquity += asset * b.LastClose[p]
	}

	asset, _, err := broker.Position(df.Pair)

	if err != nil {
		return 0, 0, err
	}

	quote := 500.0       // Simulate deposit
	totalEquity += quote // include free cash to calculate the total equity
	targetSize := b.AssetWeights[df.Pair] * totalEquity

	return targetSize, asset*b.LastClose[df.Pair] - targetSize, nil
}

func (b Balancer) OnCandle(df *model.Dataframe, broker service.Broker) {
	if dayIn(int(df.LastUpdate.Weekday()), []int{5}) {
		week := (df.LastUpdate.Day()-1)/7 + 1

		// Do not count these weeks
		if dayIn(week, []int{1, 3, 5}) {
			return
		}

		_, quotePosition, err := broker.Position(df.Pair)
		if err != nil {
			log.Error(err)
			return
		}

		expected, diff, err := b.CalculatePositionAdjustment(df, broker)
		if err != nil {
			log.Error(err)
			return
		}

		// avoid small operations
		if math.Abs(diff)/expected < 0.01 || math.Abs(diff) < 10 {
			return
		}

		if diff > 0 {
			// Sell excess of coins
			_, err = broker.CreateOrderMarketQuote(ninjabot.SideTypeSell, df.Pair, diff)
			if err != nil {
				log.Error(err)
				return
			}
		}

		if diff > quotePosition {
			log.Errorf("free cash not enough, DIFF = %.2f BUSD, CASH = %.2f BUSD", diff, quotePosition)
			return
		}

		// Buy more coins
		_, err = broker.CreateOrderMarketQuote(ninjabot.SideTypeBuy, df.Pair, -diff)
	}
}
