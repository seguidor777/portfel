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

type DiamondHands struct {
	D  *models.StrategyData
	kv *localkv.LocalKV
}

// NewDiamondHands is used in trade real and dry-run. It never sells
func NewDiamondHands(config *models.Config, kv *localkv.LocalKV) (*DiamondHands, error) {
	data, err := models.NewStrategyData(config)
	if err != nil {
		return nil, err
	}

	d := &DiamondHands{
		D:  data,
		kv: kv,
	}

	return d, nil
}

func (d DiamondHands) Timeframe() string {
	return "1d"
}

func (d DiamondHands) WarmupPeriod() int {
	return 1
}

func (d DiamondHands) Indicators(df *model.Dataframe) []strategy.ChartIndicator {
	d.D.LastClose[df.Pair] = df.Close.Last(0)

	return []strategy.ChartIndicator{}
}

func (d DiamondHands) OnCandle(df *model.Dataframe, broker service.Broker) {
	// Trade as long there is available balance
	account, err := broker.Account()
	if err != nil {
		log.Error(err)
		return
	}

	_, quoteBalance := account.Balance("", models.USDSymbol)
	quotePosition := quoteBalance.Free

	accVal, err := d.kv.Get(fmt.Sprintf("%s-acc", df.Pair))
	if err != nil {
		if err.Error() != notFoundErr {
			log.Error(err)
		}

		accVal = "0.0"
	}

	acc, _ := strconv.ParseFloat(accVal, 64)

	for _, stake := range d.D.AssetStake {
		quotePosition += stake
	}

	if quotePosition < d.D.MinimumBalance && acc == 0.0 {
		log.Errorf("The balance is below the minimum and there is no accumulation for %s", df.Pair)
		return
	}

	priceDrop, err := getPriceDrop(d.D.Slugs[df.Pair])
	if err != nil {
		log.Error(err)
		return
	}

	if math.Abs(priceDrop/100) < d.D.ExpectedPriceDrop {
		if err := d.kv.Set(fmt.Sprintf("%s-acc", df.Pair), fmt.Sprintf("%f", acc)); err != nil {
			log.Error(err)
			return
		}

		log.Warnf("%.2f USD accumulated for %s", acc, df.Pair)
		return
	}

	// Round to 2 decimals
	asset := math.Floor(d.D.AssetWeights[df.Pair]*quotePosition*100) / 100
	acc += asset

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
