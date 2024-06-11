package main

import (
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// src.GetBinanceBalance()

	// note that having a small quota would give us 502 error from SQS
	// tokenInAmount := "10000"

	// btcToUsdcPrice, err := src.GetOsmosisBTCToUSDCPrice(tokenInAmount)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Printf("BTC to USDC price for %s BTC: %f\n", tokenInAmount, btcToUsdcPrice)
}
