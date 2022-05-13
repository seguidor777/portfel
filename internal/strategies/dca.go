package strategies

import (
	"github.com/rodrigo-brito/ninjabot"
	"github.com/rodrigo-brito/ninjabot/model"
	"github.com/rodrigo-brito/ninjabot/service"
	log "github.com/sirupsen/logrus"
	"math"
)

type DCA struct {
	MinimumBalance float64
	AssetWeights   map[string]float64 `json:"asset_weight,omitempty"`
	LastClose      map[string]float64 `json:"last_close,omitempty"`
}

func NewDCA(minimumBalance float64, assetWeights map[string]float64) *DCA {
	b := &DCA{
		MinimumBalance: minimumBalance,
		AssetWeights:   assetWeights,
		LastClose:      make(map[string]float64),
	}

	return b
}

func (d DCA) Timeframe() string {
	return "1d"
}

func (d DCA) WarmupPeriod() int {
	return 1
}

func (d DCA) Indicators(df *model.Dataframe) {
	d.LastClose[df.Pair] = df.Close.Last(0)
}

func (d DCA) OnCandle(df *model.Dataframe, broker service.Broker) {
	// Invest on fridays
	if dayIn(int(df.LastUpdate.Weekday()), []int{5}) {
		week := (df.LastUpdate.Day()-1)/7 + 1

		// Do not count these weeks
		if dayIn(week, []int{1, 3, 5}) {
			return
		}

		// Return total base coins
		_, quotePosition, err := broker.Position(df.Pair)
		if err != nil {
			log.Error(err)
			return
		}

		deposit := 500.0 // Simulate deposit
		asset := math.Floor(d.AssetWeights[df.Pair]*deposit*100) / 100

		if asset > quotePosition {
			log.Errorf("free cash not enough, CASH = %.2f BUSD", quotePosition)
			return
		}

		_, err = broker.CreateOrderMarketQuote(ninjabot.SideTypeBuy, df.Pair, asset)
		if err != nil {
			log.Error(err)
		}
	}
}

func dayIn(day int, days []int) bool {
	for _, d := range days {
		if day == d {
			return true
		}
	}
	return false
}
