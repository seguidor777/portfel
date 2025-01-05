package main

import (
	"flag"
	"fmt"
	"github.com/seguidor777/portfel/internal/backtesting"
	"github.com/seguidor777/portfel/internal/exchanges/binance"
	"github.com/seguidor777/portfel/internal/exchanges/paperwallet"
	"log"
	"os"
)

var version = "0.0.1"

func main() {
	log.Println("Running version", version)
	tradeCmd := flag.NewFlagSet("trade", flag.ExitOnError)
	dryRun := tradeCmd.Bool("dry-run", false, "Trade in dry-run mode")
	tradeConfigPath := tradeCmd.String("config", "user_data/config.yml", "Configuration file path")
	databasePath := tradeCmd.String("database", "user_data/db", "Database path")

	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	testConfigPath := testCmd.String("config", "user_data/config.yml", "Configuration file path")

	if len(os.Args) < 2 {
		log.Fatalln("expected 'trade' or 'test' subcommands")
	}

	switch os.Args[1] {

	case "trade":
		tradeCmd.Parse(os.Args[2:])
		config, err := readConfig(tradeConfigPath)
		if err != nil {
			log.Fatalf("cannot read config file: %v", err)
		}

		if *dryRun {
			paperwallet.Run(config, databasePath)
			return
		}

		binance.Run(config, databasePath)
	case "test":
		testCmd.Parse(os.Args[2:])
		config, err := readConfig(testConfigPath)
		if err != nil {
			log.Fatalf("cannot read config file: %v", err)
		}

		backtesting.Run(config, databasePath)
	default:
		fmt.Println("expected 'trade' or 'test' subcommands")
		os.Exit(1)
	}
}
