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

type DCAOnSteroids struct {
	D *models.StrategyData
}

func NewDCAOnSteroids(config *models.Config) (*DCAOnSteroids, error) {
	data, err := models.NewStrategyData(config)
	if err != nil {
		return nil, err
	}

	d := &DCAOnSteroids{
		D: data,
	}

	return d, nil
}

func (d DCAOnSteroids) Timeframe() string {
	return "1d"
}

func (d DCAOnSteroids) WarmupPeriod() int {
	return 1
}

func (d DCAOnSteroids) Indicators(df *model.Dataframe) {
	d.D.LastClose[df.Pair] = df.Close.Last(0)
	d.D.LastHigh[df.Pair] = df.High.Last(0)
}

func (d DCAOnSteroids) OnCandle(df *model.Dataframe, broker service.Broker) {
	// Trade on fridays
	if !dayIn(int(df.LastUpdate.Weekday()), []int{5}) {
		return
	}

	week := (df.LastUpdate.Day()-1)/7 + 1

	// Do not count these weeks
	if dayIn(week, []int{1, 3, 5}) {
		return
	}

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

	// Calculate ATH
	if d.D.LastHigh[df.Pair] > d.D.ATHTest[df.Pair] {
		d.D.ATHTest[df.Pair] = d.D.LastHigh[df.Pair]
	}

	athDelta := d.D.ATHTest[df.Pair] - d.D.LastClose[df.Pair]

	if athDelta/d.D.ATHTest[df.Pair] < d.D.ExpectedPriceDrop {
		return
	}

	deposit := 500.0 // Simulate deposit
	asset := math.Floor(d.D.AssetWeights[df.Pair]*deposit*100) / 100

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

	d.D.Volume[df.Pair] += d.D.Accumulation[df.Pair]
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
