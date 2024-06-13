package src

import (
	"fmt"
)

func CheckArbitrage() error {
	binanceBTCPrice, err := GetBinanceBTCToUSDTPrice()
	if err != nil {
		return fmt.Errorf("error fetching Binance BTC price: %v", err)
	}
	fmt.Println("Binance BTC Price:", binanceBTCPrice)

	osmosisBTCPrice, err := GetOsmosisBTCToUSDCPrice(defaultArbAmt)
	if err != nil {
		return fmt.Errorf("error fetching Osmosis BTC price: %v", err)
	}
	fmt.Println("Osmosis BTC Price:", osmosisBTCPrice)

	b, u, err := GetBinanceBTCUSDTBalance()
	if err != nil {
		return fmt.Errorf("error fetching Osmosis BTC price: %v", err)
	}

	fmt.Printf("btc, usdc balance: %f, %f\n", b, u)

	boughtAmt, _, err := BuyBinanceBTC(defaultArbAmt)
	if err != nil {
		return err
	}
	fmt.Printf("bought %f\n", boughtAmt)

	b, u, err = GetBinanceBTCUSDTBalance()
	if err != nil {
		return fmt.Errorf("error fetching Osmosis BTC price: %v", err)
	}

	fmt.Printf("btc, usdc balance: %f, %f\n", b, u)

	sellAmt, _, err := SellBinanceBTC(defaultArbAmt)
	if err != nil {
		return err
	}
	fmt.Printf("sold %f\n", sellAmt)
	b, u, err = GetBinanceBTCUSDTBalance()
	if err != nil {
		return fmt.Errorf("error fetching Osmosis BTC price: %v", err)
	}

	fmt.Printf("btc, usdc balance: %f, %f\n", b, u)

	if binanceBTCPrice < osmosisBTCPrice*riskFactor {
		fmt.Println("Arbitrage Opportunity: Buy BTC on Binance, Sell BTC on Osmosis")

	} else if binanceBTCPrice*riskFactor > osmosisBTCPrice {
		fmt.Println("Arbitrage Opportunity: Sell BTC on Binance, Buy BTC on Osmosis")
	} else {
		fmt.Println("no arb opportunity")
	}

	return nil
}
