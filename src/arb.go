package src

import (
	"fmt"
	"time"
)

func CheckArbitrage(seedConfig SeedConfig) error {
	time := getTime()
	fmt.Println("=======Starting ARB in ", time, "=======")

	btcBalance, usdtBalance, err := GetTotalBalance(seedConfig)
	if err != nil {
		return err
	}

	fmt.Println("Balance before arb is, btc: ", btcBalance, " usdt: ", usdtBalance)

	binanceBTCPrice, err := GetBinanceBTCToUSDTPrice()
	if err != nil {
		return fmt.Errorf("error fetching Binance BTC price: %v", err)
	}

	arbAmount, err := calculateArbAmount(btcBalance, usdtBalance, binanceBTCPrice)
	if err != nil {
		return err
	}
	fmt.Println("Binance BTC Price:", binanceBTCPrice)

	osmosisBTCPrice, route, err := GetOsmosisBTCToUSDCPriceAndRoute(arbAmount)
	if err != nil {
		return fmt.Errorf("error fetching Osmosis BTC price: %v", err)
	}

	fmt.Println("Osmosis BTC Price:", osmosisBTCPrice)

	if binanceBTCPrice < osmosisBTCPrice*riskFactor {
		fmt.Println("Arbitrage Opportunity: Buy BTC on Binance, Sell BTC on Osmosis")

		_, route, err := GetOsmosisUSDCToBTCPriceAndRoute(arbAmount)

		if err != nil {
			return err
		}

		err = SellOsmosisBTC(seedConfig, route, binanceBTCPrice)
		if err != nil {
			return err
		}

		_, _, err = BuyBinanceBTC(arbAmount)
		if err != nil {
			return err
		}

		btcBalance, usdtBalance, err := GetTotalBalance(seedConfig)
		if err != nil {
			return err
		}

		fmt.Println("Balance after arb is, btc: ", btcBalance, " usdt: ", usdtBalance)

	} else if binanceBTCPrice*riskFactor > osmosisBTCPrice {
		fmt.Println("Arbitrage Opportunity: Sell BTC on Binance, Buy BTC on Osmosis")

		err = BuyOsmosisBTC(seedConfig, route, binanceBTCPrice)
		if err != nil {
			return err
		}

		_, _, err = SellBinanceBTC(arbAmount)
		if err != nil {
			return err
		}

		btcBalance, usdtBalance, err := GetTotalBalance(seedConfig)
		if err != nil {
			return err
		}

		fmt.Println("Balance after arb is, btc: ", btcBalance, " usdt: ", usdtBalance)

	} else {
		fmt.Println("No arb opportunity")
	}

	return nil
}

// for arb amount, we use 10% of the smaller asset we have between btc and usdt
// amount being returned is in units of btc
func calculateArbAmount(btcBalance, usdtBalance, btcPrice float64) (float64, error) {
	const arbPercentage = 0.1

	if btcBalance == 0 || usdtBalance == 0 {
		return 0, fmt.Errorf("insufficient balance for arbitrage")
	}

	// Calculate the BTC equivalent of the USDT balance
	btcEquivalent := usdtBalance / btcPrice

	// Calculate the arbitrage amount based on the smaller balance in BTC units
	arbAmount := arbPercentage * min(btcBalance, btcEquivalent)
	return arbAmount, nil
}

func GetTotalBalance(seedConfig SeedConfig) (float64, float64, error) {
	binanceBTCBalance, binanceUSDTBalance, err := GetBinanceBTCUSDTBalance()
	if err != nil {
		return 0, 0, fmt.Errorf("error fetching Binance balance: %v", err)
	}

	osmosisBTCBalance, osmosisUSDTBalance, err := GetOsmosisBTCUSDTBalance(seedConfig)
	if err != nil {
		return 0, 0, fmt.Errorf("error fetching Osmosis balance: %v", err)
	}

	return binanceBTCBalance + osmosisBTCBalance, binanceUSDTBalance + osmosisUSDTBalance, nil
}

func getTime() string {
	// Get the current time
	currentTime := time.Now()

	// Format the time as "YYYY-MM-DD HH:MM:SS"
	formattedTime := currentTime.Format("2006-01-02 15:04:05")
	return formattedTime
}
