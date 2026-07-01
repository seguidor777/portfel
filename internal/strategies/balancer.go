package strategies

import (
	"math"

	"github.com/rodrigo-brito/ninjabot/strategy"
	"github.com/seguidor777/portfel/internal/models"

	"github.com/rodrigo-brito/ninjabot"
	"github.com/rodrigo-brito/ninjabot/model"
	"github.com/rodrigo-brito/ninjabot/service"
	log "github.com/sirupsen/logrus"
)

type Balancer struct {
	D              *models.StrategyData
	MinimumBalance float64
}

type Weight struct {
	Pair   string
	Weight float64
}

func NewBalancer(config *models.Config) *Balancer {
	b := &Balancer{
		MinimumBalance: config.MinimumBalance,
		D: &models.StrategyData{
			MinimumBalance: config.MinimumBalance,
			AssetWeights:   config.AssetWeights,
			LastClose:      make(map[string]float64),
			LastHigh:       make(map[string]float64),
			AssetStake:     make(map[string]float64),
			Volume:         make(map[string]float64),
			SellProceeds:   make(map[string]float64),
			ATHTest:        make(map[string]float64),
			Slugs:          make(map[string]string),
			Metadata:       make(map[string]map[string]float64),
		},
	}

	return b
}

func (b Balancer) Timeframe() string {
	return "1d"
}

func (b Balancer) WarmupPeriod() int {
	return 1
}

func (b Balancer) Indicators(df *model.Dataframe) []strategy.ChartIndicator {
	b.D.LastClose[df.Pair] = df.Close.Last(0)

	return []strategy.ChartIndicator{}
}

func (b Balancer) CalculatePositionAdjustment(df *ninjabot.Dataframe, broker service.Broker) (expect, diff float64, err error) {
	totalEquity := 0.0

	for p := range b.D.AssetWeights {
		asset, _, err := broker.Position(p)
		if err != nil {
			return 0, 0, err
		}

		totalEquity += asset * b.D.LastClose[p]
	}

	asset, _, err := broker.Position(df.Pair)

	if err != nil {
		return 0, 0, err
	}

	quote := 500.0       // Simulate deposit
	totalEquity += quote // include free cash to calculate the total equity
	targetSize := b.D.AssetWeights[df.Pair] * totalEquity

	return targetSize, asset*b.D.LastClose[df.Pair] - targetSize, nil
}

func (b Balancer) OnCandle(df *model.Dataframe, broker service.Broker) {
	week := (df.LastUpdate.Day()-1)/7 + 1

	if dayIn(int(df.LastUpdate.Weekday()), []int{5}) && week == 3 {
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
			// Sell excess of coins: convert quote excess to asset quantity
			assetQty := diff / b.D.LastClose[df.Pair]
			_, err = broker.CreateOrderMarket(ninjabot.SideTypeSell, df.Pair, assetQty)
			if err != nil {
				log.Error(err)
			}
			// After selling, skip the buy pass for this candle
			return
		}

		// diff < 0: need to buy -diff USDT worth of coins
		buyAmount := -diff

		// Check free USDT balance from the account
		account, err := broker.Account()
		if err != nil {
			log.Error(err)
			return
		}
		_, usdtBalance := account.Balance("", models.USDSymbol)

		if buyAmount > usdtBalance.Free {
			log.Errorf("free cash not enough, DIFF = %.2f USDT, CASH = %.2f USDT", buyAmount, usdtBalance.Free)
			return
		}

		// Buy more coins
		_, err = broker.CreateOrderMarketQuote(ninjabot.SideTypeBuy, df.Pair, buyAmount)
		if err != nil {
			log.Error(err)
			return
		}

		b.D.Volume[df.Pair] += buyAmount
	}
}
