package backtesting

import (
	"context"
	"fmt"
	"github.com/seguidor777/portfel/internal/models"

	"github.com/rodrigo-brito/ninjabot"
	"github.com/rodrigo-brito/ninjabot/exchange"
	"github.com/rodrigo-brito/ninjabot/plot"
	"github.com/rodrigo-brito/ninjabot/storage"
	"github.com/seguidor777/portfel/internal/strategies"

	log "github.com/sirupsen/logrus"
)

func Run(config *models.Config) {
	var (
		ctx   = context.Background()
		pairs = make([]string, 0, len(config.AssetWeights))
	)

	for pair, _ := range config.AssetWeights {
		pairs = append(pairs, pair)
	}

	settings := ninjabot.Settings{
		Pairs: pairs,
	}

	strategy := strategies.NewDiamondHands(config.MinimumBalance, config.AssetWeights)
	pairFeed := make([]exchange.PairFeed, 0, len(config.AssetWeights))

	for pair, _ := range config.AssetWeights {
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
		"USDT",
		exchange.WithPaperAsset("USDT", 500),
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

	// Print bot results
	bot.Summary()
	totalEquity := 0.0
	fmt.Printf("REAL ASSETS VALUE\n")

	for pair, _ := range strategy.AssetWeights {
		asset, _, err := wallet.Position(pair)
		if err != nil {
			log.Fatal(err)
		}

		assetValue := asset * strategy.LastClose[pair]
		fmt.Printf("%s = %.2f USDT\n", pair, assetValue)
		totalEquity += assetValue
	}

	_, quote, err := wallet.Position("BTCUSDT") // Any pair, we just get the available balance
	if err != nil {
		log.Fatal(err)
	}

	totalEquity += quote
	fmt.Printf("TOTAL EQUITY = %.2f USDT\n--------------\n", totalEquity)

	// Display candlesticks chart in browser
	err = chart.Start()
	if err != nil {
		log.Fatal(err)
	}
}
