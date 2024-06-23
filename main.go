package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/osmosis-labs/arb-bot/src"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	seedConfig, err := src.OsmosisInit()
	if err != nil {
		log.Fatal(err)
	}

	btcBalance, usdcBalance, err := src.GetOsmosisBTCUSDTBalance(seedConfig)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(btcBalance)
	fmt.Println(usdcBalance)
}
