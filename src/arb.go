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

	binanceUSDCPrice, err := GetBinanceUSDCToBTCPrice()
	if err != nil {
		return fmt.Errorf("error fetching Binance USDC price: %v", err)
	}
	fmt.Println("Binance USDC Price:", binanceUSDCPrice)

	// balance, err := getBTCBalance()
	// if err != nil {
	// 	return fmt.Errorf("error fetching BTC balance: %v", err)
	// }
	// fmt.Println("BTC Balance:", balance)

	// amount := int64(defaultArbAmt) * int64(math.Pow(10, osmosisWBTCExponent))
	// osmosisBTCPrice, err := GetOsmosisBTCToUSDCPrice(amount)
	// if err != nil {
	// 	return fmt.Errorf("error fetching Osmosis BTC price: %v", err)
	// }
	// fmt.Println("Osmosis BTC Price:", osmosisBTCPrice)

	// if binanceBTCPrice < osmosisBTCPrice*riskFactor {
	// 	fmt.Println("Arbitrage Opportunity: Buy BTC on Binance, Sell BTC on Osmosis")
	// } else if binanceBTCPrice*riskFactor > osmosisBTCPrice {
	// 	fmt.Println("Arbitrage Opportunity: Sell BTC on Binance, Buy BTC on Osmosis")
	// }

	return nil
}
