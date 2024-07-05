package main

import (
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"

	"github.com/osmosis-labs/arb-bot/src"
	"github.com/osmosis-labs/arb-bot/src/domain"
)

func main() {
	// Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	seedConfig, err := src.OsmosisInit()
	if err != nil {
		fmt.Println(err)
	}

	pairs := []domain.OsmoBinanceArbPairMetadata{
		{
			BaseChainDenom:    "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			QuoteChainDenom:   "factory/osmo1em6xs47hd82806f5cxgyufguxrrc7l0aqx7nzzptjuqgswczk8csavdxek/alloyed/allUSDT",
			ExponentBase:      6,
			ExponentQuote:     6,
			BinancePairTicker: "ATOMUSDT",
			RiskFactor:        0.99,
		},
		{
			BaseChainDenom:    "factory/osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3/alloyed/allBTC",
			QuoteChainDenom:   "factory/osmo1em6xs47hd82806f5cxgyufguxrrc7l0aqx7nzzptjuqgswczk8csavdxek/alloyed/allUSDT",
			ExponentBase:      8,
			ExponentQuote:     6,
			BinancePairTicker: "BTCUSDT",
			RiskFactor:        0.99,
		},
	}

	for _, pair := range pairs {
		go runArbitrageCheck(seedConfig, pair)
	}

	// Keep the program running
	select {}
}

func runArbitrageCheck(seedConfig src.SeedConfig, arbPairMetaData domain.OsmoBinanceArbPairMetadata) {
	// Set up a ticker to run the function every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		// Execute the function
		err := src.CheckArbitrage(seedConfig, arbPairMetaData)
		fmt.Println(err)

		// Wait for the next tick
		<-ticker.C
	}
}
