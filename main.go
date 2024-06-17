package main

import (
	"log"

	"github.com/joho/godotenv"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

	err = src.TopOfBlockAuction(seedConfig, sdk.NewCoin("uosmo", sdk.NewInt(10)), sdk.NewCoin("test", sdk.NewInt(10)))
	if err != nil {
		log.Fatal(err)
	}

	// err = src.CheckArbitrage()

	// if err != nil {
	// 	log.Fatal("Error in arb logic", err)
	// }
}
