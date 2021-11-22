package binance

import (
	"context"
	"github.com/rodrigo-brito/ninjabot"
	"github.com/rodrigo-brito/ninjabot/exchange"
	"github.com/rodrigo-brito/ninjabot/storage"
	"github.com/seguidor777/portfel/internal/models"
	"github.com/seguidor777/portfel/internal/strategies"
	"log"
	"os"
	"strconv"
)

func Run(config *models.Config, databasePath *string) {
	var (
		ctx              = context.Background()
		binanceAPIKey    = os.Getenv("BINANCE_API_KEY")
		binanceSecretKey = os.Getenv("BINANCE_SECRET_KEY")
		telegramToken    = os.Getenv("TELEGRAM_TOKEN")
		telegramUser, _  = strconv.Atoi(os.Getenv("TELEGRAM_USER"))
		pairs            = make([]string, 0, len(config.AssetWeights))
	)

	for pair, _ := range config.AssetWeights {
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

	// creating a storage to save trades
	storage, err := storage.FromFile(*databasePath)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize exchange
	binanceCredential := exchange.WithBinanceCredentials(binanceAPIKey, binanceSecretKey)
	binance, err := exchange.NewBinance(ctx, binanceCredential)
	if err != nil {
		log.Fatalln(err)
	}

	// Initialize strategy and bot
	strategy := strategies.NewDiamondHands(config.MinimumBalance, config.AssetWeights)
	bot, err := ninjabot.NewBot(ctx, settings, binance, strategy, ninjabot.WithStorage(storage))
	if err != nil {
		log.Fatalln(err)
	}

	// Run ninjabot
	err = bot.Run(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}
