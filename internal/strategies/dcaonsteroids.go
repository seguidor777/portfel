package strategies

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/rodrigo-brito/ninjabot"
	"github.com/rodrigo-brito/ninjabot/indicator"
	"github.com/rodrigo-brito/ninjabot/model"
	"github.com/rodrigo-brito/ninjabot/service"
	"github.com/rodrigo-brito/ninjabot/strategy"
	"github.com/seguidor777/portfel/internal/localkv"
	"github.com/seguidor777/portfel/internal/models"
	log "github.com/sirupsen/logrus"
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
	return 21
}

func (d DCAOnSteroids) Indicators(df *model.Dataframe) []strategy.ChartIndicator {
	d.D.LastClose[df.Pair] = df.Close.Last(0)
	d.D.LastHigh[df.Pair] = df.High.Last(0)
	sma20 := indicator.SMA(df.Close, 20)

	// Ensure the inner map for this pair is initialized
	if d.D.Metadata[df.Pair] == nil {
		d.D.Metadata[df.Pair] = make(map[string]float64)
	}

	// Store only the latest SMA value for use in OnCandle
	if len(sma20) > 0 {
		d.D.Metadata[df.Pair]["sma20"] = sma20[len(sma20)-1]
	}

	return []strategy.ChartIndicator{
		{
			Overlay:   true,
			GroupName: "MA's",
			Time:      df.Time,
			Metrics: []strategy.IndicatorMetric{
				{
					Values: sma20,
					Name:   "EMA 20",
					Color:  "red",
					Style:  strategy.StyleLine,
				},
			},
		},
	}
}

func (d *DCAOnSteroids) OnCandle(df *model.Dataframe, broker service.Broker) {
	atomic.AddUint32(&counter, 1)
	defer d.resetAssetStake()
	account, err := broker.Account()
	if err != nil {
		log.Error(err)
		return
	}

	// Sell 100% when price reaches ATH, regardless of balance or accumulation state
	if d.D.LastHigh[df.Pair] >= d.D.ATHTest[df.Pair] {
		d.D.ATHTest[df.Pair] = d.D.LastHigh[df.Pair]
		assetSymbol := strings.TrimSuffix(df.Pair, models.USDSymbol)
		assetBalance, _ := account.Balance(assetSymbol, models.USDSymbol)
		if assetBalance.Free > 0 {
			_, err = broker.CreateOrderMarket(ninjabot.SideTypeSell, df.Pair, assetBalance.Free)
			if err != nil {
				log.Errorf("Cannot sell %s at ATH: %v", df.Pair, err)
			} else {
				proceeds := assetBalance.Free * d.D.LastClose[df.Pair]
				d.D.SellProceeds[df.Pair] += proceeds
				log.Warnf("Sold 100%% of %s at price %.2f (proceeds: %.2f USDT)", df.Pair, d.D.LastClose[df.Pair], proceeds)
				if err := d.kv.Set(fmt.Sprintf("%s-acc", df.Pair), "0.0"); err != nil {
					log.Error(err)
				}
			}
		}
		return
	}

	accVal, err := d.kv.Get(fmt.Sprintf("%s-acc", df.Pair))
	if err != nil {
		if err.Error() != notFoundErr {
			log.Error(err)
		}

		accVal = "0.0"
	}

	acc, _ := strconv.ParseFloat(accVal, 64)
	quotePosition := 0.0

	if int(df.LastUpdate.Day()) == 9 {
		quotePosition = d.D.MinimumBalance
	}

	if quotePosition == 0.0 && acc == 0.0 {
		return
	}

	// Calculate ATH
	if d.D.LastHigh[df.Pair] > d.D.ATHTest[df.Pair] {
		d.D.ATHTest[df.Pair] = d.D.LastHigh[df.Pair]
	}

	athDrop := math.Abs((d.D.ATHTest[df.Pair] - d.D.LastClose[df.Pair]) / d.D.ATHTest[df.Pair])

	if quotePosition >= d.D.MinimumBalance {
		// Each pair gets its proportional share of the injection amount.
		// Using weight × quotePosition ensures total spending = MinimumBalance,
		// regardless of the order pairs are processed.
		newStake := math.Floor(d.D.AssetWeights[df.Pair]*quotePosition*100) / 100
		d.D.AssetStake[df.Pair] = newStake
		acc += newStake
	}

	if athDrop < d.D.ExpectedPriceDrop {
		if err := d.kv.Set(fmt.Sprintf("%s-acc", df.Pair), fmt.Sprintf("%f", acc)); err != nil {
			log.Error(err)
			return
		}

		log.Warnf("%.2f USD accumulated for %s", acc, df.Pair)
		return
	}

	// Check available balance before placing the order
	_, quoteBalance := account.Balance("", models.USDSymbol)
	if acc > quoteBalance.Free {
		log.Warnf("Insufficient funds for %s: need %.2f USDT but have %.2f USDT, keeping accumulation", df.Pair, acc, quoteBalance.Free)
		if err := d.kv.Set(fmt.Sprintf("%s-acc", df.Pair), fmt.Sprintf("%f", acc)); err != nil {
			log.Error(err)
		}
		return
	}

	// Buy more coins
	_, err = broker.CreateOrderMarketQuote(ninjabot.SideTypeBuy, df.Pair, acc)
	if err != nil {
		log.Errorf("Cannot create order for %s (%.2f USDT): %v", df.Pair, acc, err)
		return
	}

	d.D.Volume[df.Pair] += acc

	// Reset accumulation
	if err := d.kv.Set(fmt.Sprintf("%s-acc", df.Pair), "0.0"); err != nil {
		log.Error(err)
	}
}

func (d *DCAOnSteroids) resetAssetStake() {
	// If diversification has been completed then reset stakes
	if int(atomic.LoadUint32(&counter)) == len(d.D.AssetWeights) {
		for pair := range d.D.AssetStake {
			d.D.AssetStake[pair] = 0.0
		}

		atomic.CompareAndSwapUint32(&counter, uint32(len(d.D.AssetWeights)), 0)
		log.Warnln("Asset stakes have been reset")
	}
}
