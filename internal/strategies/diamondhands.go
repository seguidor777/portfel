package strategies

import (
	"github.com/rodrigo-brito/ninjabot"
	"github.com/rodrigo-brito/ninjabot/model"
	"github.com/rodrigo-brito/ninjabot/service"
	"github.com/seguidor777/portfel/internal/models"
	log "github.com/sirupsen/logrus"
	"math"
	"sync/atomic"
)

type DiamondHands struct {
	D *models.StrategyData
}

func NewDiamondHands(config *models.Config) (*DiamondHands, error) {
	data, err := models.NewStrategyData(config)
	if err != nil {
		return nil, err
	}

	d := &DiamondHands{
		D: data,
	}

	return d, nil
}

func (d DiamondHands) Timeframe() string {
	return "1d"
}

func (d DiamondHands) WarmupPeriod() int {
	return 1
}

func (d DiamondHands) Indicators(df *model.Dataframe) {
	d.D.LastClose[df.Pair] = df.Close.Last(0)
}

func (d DiamondHands) OnCandle(df *model.Dataframe, broker service.Broker) {
	// Trade as long there is available balance
	account, err := broker.Account()
	if err != nil {
		log.Error(err)
		return
	}

	balance := account.Balance(models.USDSymbol)
	quotePosition := balance.Free

	for _, stake := range d.D.AssetStake {
		quotePosition += stake
	}

	if quotePosition < d.D.MinimumBalance {
		log.Error("The balance is below the minimum")
		return
	}

	priceDrop, err := getPriceDrop(df.Pair)
	if err != nil {
		log.Error(err)
		return
	}

	if math.Abs(priceDrop/100) < d.D.ExpectedPriceDrop {
		return
	}

	// Round to 2 decimals
	asset := math.Floor(d.D.AssetWeights[df.Pair]*quotePosition*100) / 100

	if d.D.Accumulation[df.Pair] > quotePosition {
		log.Errorf("free cash not enough, CASH = %.2f BUSD", quotePosition)
		return
	}

	// Buy more coins
	d.D.Accumulation[df.Pair] += asset
	_, err = broker.CreateOrderMarketQuote(ninjabot.SideTypeBuy, df.Pair, d.D.Accumulation[df.Pair])
	if err != nil {
		log.Error(err)
		return
	}

	d.D.Accumulation[df.Pair] = 0.0 // Reset accumulation
	atomic.AddUint32(&counter, 1)

	// If diversification has been completed then reset stakes
	if int(atomic.LoadUint32(&counter)) == len(d.D.AssetWeights) {
		for pair := range d.D.AssetStake {
			d.D.AssetStake[pair] = 0.0
		}

		atomic.CompareAndSwapUint32(&counter, counter, 0)
		return
	}

	// Save asset stake for further calculation
	d.D.AssetStake[df.Pair] = asset
}
