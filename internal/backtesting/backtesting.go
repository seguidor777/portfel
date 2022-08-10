package backtesting

import (
	"context"
	"fmt"
	"github.com/seguidor777/portfel/internal/localkv"
	"github.com/seguidor777/portfel/internal/models"

	"github.com/rodrigo-brito/ninjabot"
	"github.com/rodrigo-brito/ninjabot/exchange"
	"github.com/rodrigo-brito/ninjabot/plot"
	"github.com/rodrigo-brito/ninjabot/storage"
	"github.com/seguidor777/portfel/internal/strategies"

	log "github.com/sirupsen/logrus"
)

const walletAmount = 12000

// TODO: Pass name of strategy and call it from a switch
func Run(config *models.Config, databasePath *string) {
	var (
		ctx   = context.Background()
		pairs = make([]string, 0, len(config.AssetWeights))
	)

	for pair := range config.AssetWeights {
		pairs = append(pairs, pair)
	}

	settings := ninjabot.Settings{
		Pairs: pairs,
	}

	// initialize local KV store for strategies
	kv, err := localkv.NewLocalKV(*databasePath)
	if err != nil {
		log.Fatal(err)
	}

	strategy, err := strategies.NewDCAOnSteroids(config, kv)
	if err != nil {
		log.Fatal(err)
	}

	pairFeed := make([]exchange.PairFeed, 0, len(config.AssetWeights))

	for pair := range config.AssetWeights {
		pairFeed = append(pairFeed, exchange.PairFeed{
			Pair:      pair,
			File:      fmt.Sprintf("testdata/%s-%s.csv", pair, strategy.Timeframe()),
			Timeframe: strategy.Timeframe(),
		})
	}

	csvFeed, err := exchange.NewCSVFeed(strategy.Timeframe(), pairFeed...)
	if err != nil {
		log.Fatal(err)
	}

	storage, err := storage.FromMemory()
	if err != nil {
		log.Fatal(err)
	}

	wallet := exchange.NewPaperWallet(
		ctx,
		"BUSD",
		exchange.WithPaperAsset("BUSD", walletAmount),
		exchange.WithDataFeed(csvFeed),
	)

	chart, err := plot.NewChart(plot.WithPaperWallet(wallet))
	if err != nil {
		log.Fatal(err)
	}

	bot, err := ninjabot.NewBot(
		ctx,
		settings,
		wallet,
		strategy,
		ninjabot.WithBacktest(wallet),
		ninjabot.WithStorage(storage),
		ninjabot.WithCandleSubscription(chart),
		ninjabot.WithOrderSubscription(chart),
		ninjabot.WithLogLevel(log.WarnLevel),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = bot.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}

	kv.RemoveDB()

	// Print bot results
	bot.Summary()
	totalEquity := 0.0
	fmt.Printf("REAL ASSETS VALUE\n")

	for pair := range strategy.D.AssetWeights {
		asset, _, err := wallet.Position(pair)
		if err != nil {
			log.Fatal(err)
		}

		assetValue := asset * strategy.D.LastClose[pair]
		volume := strategy.D.Volume[pair]
		profitPerc := (assetValue - volume) / volume * 100
		fmt.Printf("%s = %.2f BUSD, Asset Qty = %f, Profit = %.2f%%\n", pair, assetValue, asset, profitPerc)
		totalEquity += assetValue
	}

	totalVolume := 0.0

	for _, volume := range strategy.D.Volume {
		totalVolume += volume
	}

	totalProfit := totalEquity - totalVolume
	totalProfitPerc := totalProfit / totalVolume * 100
	fmt.Printf("TOTAL EQUITY = %.2f BUSD, Profit = %.2f = %.2f%%\n--------------\n", totalEquity, totalProfit, totalProfitPerc)

	// Display candlesticks chart in browser
	err = chart.Start()
	if err != nil {
		log.Fatal(err)
	}
}
