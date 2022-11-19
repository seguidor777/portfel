package strategies

import (
	"fmt"
	"github.com/rodrigo-brito/ninjabot"
	"github.com/rodrigo-brito/ninjabot/model"
	"github.com/rodrigo-brito/ninjabot/service"
	"github.com/rodrigo-brito/ninjabot/strategy"
	"github.com/seguidor777/portfel/internal/localkv"
	"github.com/seguidor777/portfel/internal/models"
	log "github.com/sirupsen/logrus"
	"math"
	"strconv"
	"sync/atomic"
)

type DCAOnSteroids struct {
	D  *models.StrategyData
	kv *localkv.LocalKV
}

// NewDCAOnSteroids is used for backtesting
func NewDCAOnSteroids(config *models.Config, kv *localkv.LocalKV) (*DCAOnSteroids, error) {
	data, err := models.NewStrategyData(config)
	if err != nil {
		return nil, err
	}

	d := &DCAOnSteroids{
		D:  data,
		kv: kv,
	}

	return d, nil
}

func (d DCAOnSteroids) Timeframe() string {
	return "1d"
}

func (d DCAOnSteroids) WarmupPeriod() int {
	return 1
}

func (d DCAOnSteroids) Indicators(df *model.Dataframe) []strategy.ChartIndicator {
	d.D.LastClose[df.Pair] = df.Close.Last(0)
	d.D.LastHigh[df.Pair] = df.High.Last(0)

	return []strategy.ChartIndicator{}
}

func (d DCAOnSteroids) OnCandle(df *model.Dataframe, broker service.Broker) {
	accVal, err := d.kv.Get(fmt.Sprintf("%s-acc", df.Pair))
	if err != nil {
		if err.Error() != notFoundErr {
			log.Error(err)
		}

		accVal = "0.0"
	}

	acc, _ := strconv.ParseFloat(accVal, 64)

	// If there is no accumulation
	if acc == 0.0 {
		// Trade on thursdays
		if !dayIn(int(df.LastUpdate.Weekday()), []int{4}) {
			return
		}

		week := (df.LastUpdate.Day()-1)/7 + 1

		// Do not count these weeks
		if dayIn(week, []int{1, 3, 5}) {
			return
		}
	}

	// Trade as long there is available balance
	account, err := broker.Account()
	if err != nil {
		log.Error(err)
		return
	}

	_, quoteBalance := account.Balance("", models.USDSymbol)
	quotePosition := quoteBalance.Free

	for _, stake := range d.D.AssetStake {
		quotePosition += stake
	}

	if quotePosition < d.D.MinimumBalance && acc == 0.0 {
		log.Errorf("The balance is below the minimum and there is no accumulation for %s", df.Pair)
		return
	}

	deposit := d.D.MinimumBalance // Simulate deposit
	asset := math.Floor(d.D.AssetWeights[df.Pair]*deposit*100) / 100

	// Calculate ATH
	if d.D.LastHigh[df.Pair] > d.D.ATHTest[df.Pair] {
		d.D.ATHTest[df.Pair] = d.D.LastHigh[df.Pair]
	}

	athDelta := d.D.ATHTest[df.Pair] - d.D.LastClose[df.Pair]
	acc += asset

	if athDelta/d.D.ATHTest[df.Pair] < d.D.ExpectedPriceDrop {
		if err := d.kv.Set(fmt.Sprintf("%s-acc", df.Pair), fmt.Sprintf("%f", acc)); err != nil {
			log.Error(err)
			return
		}

		log.Warnf("%.2f USD accumulated for %s", acc, df.Pair)
		return
	}

	if acc > quotePosition {
		log.Errorf("free cash not enough, CASH = %.2f USDT", quotePosition)
		return
	}

	// Buy more coins
	_, err = broker.CreateOrderMarketQuote(ninjabot.SideTypeBuy, df.Pair, acc)
	if err != nil {
		log.Error(err)
		return
	}

	d.D.Volume[df.Pair] += acc

	// Reset accumulation
	if err := d.kv.Set(fmt.Sprintf("%s-acc", df.Pair), "0.0"); err != nil {
		log.Error(err)
		return
	}

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
