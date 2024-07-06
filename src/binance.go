package src

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/adshao/go-binance/v2"
	"github.com/osmosis-labs/arb-bot/src/domain"
)

type BinanceResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func GetBinancePrice(ticker string) (float64, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", ticker)
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error fetching price from Binance: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error fetching price from Binance: status code %d", resp.StatusCode)
	}

	var binanceResp BinanceResponse
	err = json.NewDecoder(resp.Body).Decode(&binanceResp)
	if err != nil {
		return 0, fmt.Errorf("error decoding response: %v", err)
	}

	price, err := strconv.ParseFloat(binanceResp.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing price: %v", err)
	}

	return price, nil
}

func GetBinanceInvertedPrice(ticker string) (float64, error) {
	basePrice, err := GetBinancePrice(ticker)
	if err != nil {
		return 0, err
	}

	if basePrice != 0 {
		return 1 / basePrice, nil
	} else {
		return 0, fmt.Errorf("BTC price from Binance is zero, cannot compute USDC to BTC price")
	}
}

func GetBinanceBaseQuoteBalance(arbMetadata domain.OsmoBinanceArbPairMetadata) (baseBalance float64, quoteBalance float64, err error) {
	apiKey := os.Getenv("BINANCE_API_KEY")
	secretKey := os.Getenv("BINANCE_SECRET_KEY")

	client := binance.NewClient(apiKey, secretKey)
	accountService := client.NewGetAccountService()
	res, err := accountService.Do(context.Background())
	if err != nil {
		return 0, 0, err
	}

	filteredBalances := filterBalances(res.Balances, []string{arbMetadata.BaseBinanceSymbol, arbMetadata.QuoteBinanceSymbol})
	for _, balance := range filteredBalances {
		fmt.Printf("Asset: %s, Free: %s\n", balance.Asset, balance.Free)
	}

	baseBalance, err = strconv.ParseFloat(filteredBalances[0].Free, 64)
	if err != nil {
		return 0, 0, err
	}

	quoteBalance, err = strconv.ParseFloat(filteredBalances[1].Free, 64)
	if err != nil {
		return 0, 0, err
	}

	return baseBalance, quoteBalance, nil
}

func BuyBinanceBase(amount float64, ticker string) (boughtAmount, boughtPrice float64, err error) {
	apiKey := os.Getenv("BINANCE_API_KEY")
	secretKey := os.Getenv("BINANCE_SECRET_KEY")
	client := binance.NewClient(apiKey, secretKey)

	amountStr := strconv.FormatFloat(amount, 'f', -1, 64)

	// TODO: consider doing limit orders here
	order, err := client.NewCreateOrderService().Symbol(ticker).Side(binance.SideTypeBuy).Type(binance.OrderTypeMarket).Quantity(amountStr).Do(context.Background())
	if err != nil {
		return 0, 0, err
	}

	boughtPrice, err = strconv.ParseFloat(order.Fills[0].Price, 64)
	if err != nil {
		return 0, 0, err
	}

	boughtAmount, err = strconv.ParseFloat(order.ExecutedQuantity, 64)
	if err != nil {
		return 0, 0, err
	}

	return boughtAmount, boughtPrice, nil
}

func SellBinanceBase(amount float64, ticker string) (soldAmount, soldPrice float64, err error) {
	apiKey := os.Getenv("BINANCE_API_KEY")
	secretKey := os.Getenv("BINANCE_SECRET_KEY")
	client := binance.NewClient(apiKey, secretKey)

	amountStr := strconv.FormatFloat(amount, 'f', -1, 64)

	// TODO: consider doing limit orders here
	order, err := client.NewCreateOrderService().Symbol(ticker).Side(binance.SideTypeSell).Type(binance.OrderTypeMarket).Quantity(amountStr).Do(context.Background())
	if err != nil {
		return 0, 0, err
	}

	soldPrice, err = strconv.ParseFloat(order.Fills[0].Price, 64)
	if err != nil {
		return 0, 0, err
	}

	soldAmount, err = strconv.ParseFloat(order.ExecutedQuantity, 64)
	if err != nil {
		return 0, 0, err
	}

	return soldAmount, soldPrice, nil
}

func filterBalances(balances []binance.Balance, assets []string) []binance.Balance {
	var filtered []binance.Balance
	// We do a N^2 loop to preserve the order of the assets in the filtered list
	for _, asset := range assets {
		for _, balance := range balances {
			if balance.Asset == asset {
				filtered = append(filtered, balance)
			}
		}
	}
	return filtered
}
