package strategies

import (
	"github.com/rodrigo-brito/ninjabot"
	"github.com/rodrigo-brito/ninjabot/model"
	"github.com/rodrigo-brito/ninjabot/service"
	log "github.com/sirupsen/logrus"
	"math"
	"sync/atomic"
)

const usdSymbol = "BUSD"

var counter uint32

type DiamondHands struct {
	MinimumBalance float64            `json:"minimum_balance"`
	AssetWeights   map[string]float64 `json:"asset_weight"`
	LastClose      map[string]float64 `json:"last_close,omitempty"`
	AssetStake     map[string]float64 `json:"asset_stake,omitempty"`
}

func NewDiamondHands(minimumBalance float64, assetWeights map[string]float64) *DiamondHands {
	d := &DiamondHands{
		MinimumBalance: minimumBalance,
		AssetWeights:   assetWeights,
		LastClose:      make(map[string]float64),
		AssetStake:     make(map[string]float64),
	}

	return d
}

func (d DiamondHands) Timeframe() string {
	return "1d"
}

func (d DiamondHands) WarmupPeriod() int {
	return 1
}

func (d DiamondHands) Indicators(df *model.Dataframe) {
	d.LastClose[df.Pair] = df.Close.Last(0)
}

func (d DiamondHands) OnCandle(df *model.Dataframe, broker service.Broker) {
	// Trade as long there is available balance
	account, err := broker.Account()
	if err != nil {
		log.Error(err)
		return
	}

	balance := account.Balance(usdSymbol)
	quotePosition := balance.Free

	for _, stake := range d.AssetStake {
		quotePosition += stake
	}

	if quotePosition >= d.MinimumBalance {
		// Round to 2 decimals
		asset := math.Floor(d.AssetWeights[df.Pair]*quotePosition*100) / 100

		// Buy more coins
		_, err = broker.CreateOrderMarketQuote(ninjabot.SideTypeBuy, df.Pair, asset)
		if err != nil {
			log.Error(err)
			return
		}

		atomic.AddUint32(&counter, 1)

		// If diversification has been completed then reset stakes
		if int(atomic.LoadUint32(&counter)) == len(d.AssetWeights) {
			for pair := range d.AssetStake {
				d.AssetStake[pair] = 0.0
			}

			atomic.CompareAndSwapUint32(&counter, counter, 0)
			return
		}

		// Save asset stake for further calculation
		d.AssetStake[df.Pair] = asset
	}
}
