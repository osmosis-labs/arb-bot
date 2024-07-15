package src

import (
	"fmt"
	"sync"
	"time"

	"github.com/osmosis-labs/arb-bot/src/domain"
	"go.uber.org/atomic"
)

var (
	arbitrageOpCount *atomic.Int64 = atomic.NewInt64(0)
	arbLock          sync.Mutex
)

func CheckArbitrage(seedConfig SeedConfig, arbPairMetaData domain.OsmoBinanceArbPairMetadata) error {
	// Note: this is for prints to look nice in console, we can optimize this in the future.
	arbLock.Lock()
	defer arbLock.Unlock()

	time := getTime()
	fmt.Println("=======Starting ARB in ", time, "=======")

	baseBalance, quoteBalance, err := GetTotalBalance(seedConfig, arbPairMetaData)
	if err != nil {
		return err
	}

	fmt.Println("Balance before arb is, base: ", baseBalance, " quote: ", quoteBalance)

	binanceBasePrice, err := GetBinancePrice(arbPairMetaData.BinancePairTicker)
	if err != nil {
		return fmt.Errorf("error fetching Binance price: %v", err)
	}

	arbAmount, err := calculateArbAmount(baseBalance, quoteBalance, binanceBasePrice)
	if err != nil {
		return err
	}
	fmt.Println("Binance base Price:", binanceBasePrice)

	osmosisBTCPrice, route, err := GetOsmosisBTCToUSDCPriceAndRoute(arbAmount, arbPairMetaData.BaseChainDenom, arbPairMetaData.QuoteChainDenom, arbPairMetaData.ExponentBase, arbPairMetaData.ExponentQuote)
	if err != nil {
		return fmt.Errorf("error fetching Osmosis base price: %v", err)
	}

	fmt.Println("Osmosis BTC Price:", osmosisBTCPrice)

	if binanceBasePrice < osmosisBTCPrice*arbPairMetaData.RiskFactor {

		fmt.Println("Arbitrage Opportunity: Buy base on Binance, Sell base on Osmosis")

		_, route, err := GetOsmosisUSDCToBTCPriceAndRoute(arbAmount, arbPairMetaData.BaseChainDenom, arbPairMetaData.QuoteChainDenom, arbPairMetaData.ExponentBase, arbPairMetaData.ExponentQuote)

		if err != nil {
			return err
		}

		err = SellOsmosisBase(seedConfig, arbPairMetaData.BaseChainDenom, route, binanceBasePrice)
		if err != nil {
			return err
		}

		_, _, err = BuyBinanceBTC(arbAmount)
		if err != nil {
			return err
		}

		btcBalance, usdtBalance, err := GetTotalBalance(seedConfig, arbPairMetaData)
		if err != nil {
			return err
		}

		fmt.Println("Balance after arb is, base: ", btcBalance, " quote: ", usdtBalance)

		arbitrageOpCount.Add(1)

	} else if binanceBasePrice*arbPairMetaData.RiskFactor > osmosisBTCPrice {
		fmt.Println("Arbitrage Opportunity: Sell base on Binance, Buy base on Osmosis")

		err = BuyOsmosisBase(seedConfig, arbPairMetaData.BaseChainDenom, route, binanceBasePrice)
		if err != nil {
			return err
		}

		_, _, err = SellBinanceBTC(arbAmount)
		if err != nil {
			return err
		}

		baseBalance, quoteBalance, err := GetTotalBalance(seedConfig, arbPairMetaData)
		if err != nil {
			return err
		}

		fmt.Println("Balance after arb is, base: ", baseBalance, " quote: ", quoteBalance)

		arbitrageOpCount.Add(1)

	} else {
		fmt.Println("No arb opportunity")
	}

	fmt.Println("arb count: ", arbitrageOpCount)

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

func GetTotalBalance(seedConfig SeedConfig, arbMetadata domain.OsmoBinanceArbPairMetadata) (float64, float64, error) {
	binanceBTCBalance, binanceUSDTBalance, err := GetBinanceBTCUSDTBalance()
	if err != nil {
		return 0, 0, fmt.Errorf("error fetching Binance balance: %v", err)
	}

	osmosisBTCBalance, osmosisUSDTBalance, err := GetOsmosisBTCUSDTBalance(seedConfig, arbMetadata)
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
