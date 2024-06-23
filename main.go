package main

import (
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"

	"github.com/osmosis-labs/arb-bot/src"
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

	// Set up a ticker to run the function every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		// Execute the function
		err = runArbitrageCheck(seedConfig)
		fmt.Println(err)

		// Wait for the next tick
		<-ticker.C
	}
}

func runArbitrageCheck(seedConfig src.SeedConfig) error {
	err := src.CheckArbitrage(seedConfig)
	return err
}
