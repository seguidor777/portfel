package paperwallet

import (
	"context"
	"github.com/seguidor777/portfel/internal/localkv"
	"github.com/seguidor777/portfel/internal/models"
	"os"
	"strconv"

	"github.com/rodrigo-brito/ninjabot"
	"github.com/rodrigo-brito/ninjabot/exchange"
	"github.com/rodrigo-brito/ninjabot/storage"
	"github.com/seguidor777/portfel/internal/strategies"

	log "github.com/sirupsen/logrus"
)

func Run(config *models.Config, databasePath *string) {
	var (
		ctx             = context.Background()
		telegramToken   = os.Getenv("TELEGRAM_TOKEN")
		telegramUser, _ = strconv.Atoi(os.Getenv("TELEGRAM_USER"))
		pairs           = make([]string, 0, len(config.AssetWeights))
	)

	for pair := range config.AssetWeights {
		pairs = append(pairs, pair)
	}

	settings := ninjabot.Settings{
		Pairs: pairs,
		Telegram: ninjabot.TelegramSettings{
			Enabled: true,
			Token:   telegramToken,
			Users:   []int{telegramUser},
		},
	}

	// Use binance for realtime data feed
	binance, err := exchange.NewBinance(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// creating a storage to save trades
	storage, err := storage.FromMemory()
	if err != nil {
		log.Fatal(err)
	}

	// creating a paper wallet to simulate an exchange waller for fake operations
	paperWallet := exchange.NewPaperWallet(
		ctx,
		"BUSD",
		exchange.WithPaperFee(0.001, 0.001),
		exchange.WithPaperAsset("BUSD", 500),
		exchange.WithDataFeed(binance),
	)

	// initialize local KV store for strategies
	kv, err := localkv.NewLocalKV(*databasePath)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize strategy and bot
	switch config.Strategy {
	case "DiamondHands":
		strat, err := strategies.NewDiamondHands(config, kv)
		if err != nil {
			log.Fatal(err)
		}

		bot, err := ninjabot.NewBot(
			ctx,
			settings,
			paperWallet,
			strat,
			ninjabot.WithStorage(storage),
			ninjabot.WithPaperWallet(paperWallet),
		)
		if err != nil {
			log.Fatalln(err)
		}

		err = bot.Run(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	default:
		log.Fatal("Invalid strategy")
	}
}
