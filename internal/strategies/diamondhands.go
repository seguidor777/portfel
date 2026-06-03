package strategies

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/rodrigo-brito/ninjabot"
	"github.com/rodrigo-brito/ninjabot/model"
	"github.com/rodrigo-brito/ninjabot/service"
	"github.com/rodrigo-brito/ninjabot/strategy"
	"github.com/seguidor777/portfel/internal/localkv"
	"github.com/seguidor777/portfel/internal/models"
	log "github.com/sirupsen/logrus"
)

type DiamondHands struct {
	D  *models.StrategyData
	kv *localkv.LocalKV
}

// NewDiamondHands is used in trade real and dry-run. Sells 100% of a position when price reaches ATH.
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
	d.D.LastHigh[df.Pair] = df.High.Last(0)

	return []strategy.ChartIndicator{}
}

func (d *DiamondHands) OnCandle(df *model.Dataframe, broker service.Broker) {
	atomic.AddUint32(&counter, 1)
	defer d.resetAssetStake()

	// Trade as long there is available balance
	account, err := broker.Account()
	if err != nil {
		log.Error(err)
		return
	}

	_, quoteBalance := account.Balance("", models.USDSymbol)

	// Single API call: fetches both ATH price and price drop percentage
	marketData, err := getMarketData(d.D.Slugs[df.Pair])
	if err != nil {
		log.Error(err)
		return
	}

	ath := marketData["ath"].(map[string]interface{})["usd"].(float64)
	priceDrop := marketData["ath_change_percentage"].(map[string]interface{})["usd"].(float64)

	// Sell 100% when close reaches ATH, regardless of balance or accumulation state
	if d.D.LastHigh[df.Pair] >= ath {
		assetSymbol := strings.TrimSuffix(df.Pair, models.USDSymbol)
		assetBalance, _ := account.Balance(assetSymbol, models.USDSymbol)
		if assetBalance.Free > 0 {
			_, err = broker.CreateOrderMarket(ninjabot.SideTypeSell, df.Pair, assetBalance.Free)
			if err != nil {
				log.Errorf("Cannot sell %s at ATH: %v", df.Pair, err)
			} else {
				log.Warnf("Sold 100%% of %s at ATH price %.2f", df.Pair, ath)
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

	// Total balance = free balance + asset stakes already placed - accumulation.
	// This is the single source of truth used for both the minimum balance check
	// and as the base for calculating each asset's new stake.
	var placedStakes float64
	for _, stake := range d.D.AssetStake {
		placedStakes += stake
	}

	totalBalance := quoteBalance.Free + placedStakes - acc

	// Trade as long as there is available balance (>= configured minimum).
	// If balance is below minimum and there is no pending accumulation, skip.
	if totalBalance < d.D.MinimumBalance && acc == 0.0 {
		log.Errorf("Balance is below the minimum and there is no accumulation for %s", df.Pair)
		return
	}

	if totalBalance >= d.D.MinimumBalance {
		// Round to 2 decimals
		d.D.AssetStake[df.Pair] = math.Floor(d.D.AssetWeights[df.Pair]*totalBalance*100) / 100
		acc += d.D.AssetStake[df.Pair]
	}

	if math.Abs(priceDrop/100) < d.D.ExpectedPriceDrop {
		if err := d.kv.Set(fmt.Sprintf("%s-acc", df.Pair), fmt.Sprintf("%f", acc)); err != nil {
			log.Error(err)
			return
		}

		log.Warnf("%.2f USD accumulated for %s", acc, df.Pair)
		return
	}

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
		log.Error(err)
		return
	}

	// Reset accumulation
	if err := d.kv.Set(fmt.Sprintf("%s-acc", df.Pair), "0.0"); err != nil {
		log.Error(err)
	}
}

func (d *DiamondHands) resetAssetStake() {
	// If diversification has been completed then reset stakes
	if int(atomic.LoadUint32(&counter)) == len(d.D.AssetWeights) {
		for pair := range d.D.AssetStake {
			d.D.AssetStake[pair] = 0.0
		}

		atomic.CompareAndSwapUint32(&counter, uint32(len(d.D.AssetWeights)), 0)
		log.Warnln("Asset stakes have been reset")
	}
}
