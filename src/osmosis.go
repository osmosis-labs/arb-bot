package src

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
)

func GetOsmosisBTCToUSDCPrice(tokenInAmount float64) (float64, error) {
	amountWithExponentApplied := int64(tokenInAmount * math.Pow(10, osmosisWBTCExponent))

	btcPrice, err := getOsmosisPrice(BTCDenom, USDCDenom, amountWithExponentApplied)
	if err != nil {
		return 0, err
	}

	return btcPrice / math.Pow(10, osmosisUSDCExponent), nil
}

func GetOsmosisUSDCToBTCPrice(tokenInAmount float64) (float64, error) {
	amountWithExponentApplied := int64(tokenInAmount * math.Pow(10, osmosisWBTCExponent))

	usdcPrice, err := getOsmosisPrice(USDCDenom, BTCDenom, amountWithExponentApplied)
	if err != nil {
		return 0, err
	}
	return usdcPrice / math.Pow(10, osmosisUSDCExponent), nil
}

type QuoteResponse struct {
	AmountOut string `json:"amount_out"`
}

func getOsmosisPrice(tokenInDenom, tokenOutDenom string, tokenInAmount int64) (float64, error) {
	url := fmt.Sprintf("%s?tokenIn=%d%s&tokenOutDenom=%s&humanDenoms=false", osmosisQuoteAPI, tokenInAmount, tokenInDenom, tokenOutDenom)
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error fetching price from Osmosis: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error fetching price from Osmosis: status code %d", resp.StatusCode)
	}

	var osmosisResp QuoteResponse
	err = json.NewDecoder(resp.Body).Decode(&osmosisResp)
	if err != nil {
		return 0, fmt.Errorf("error decoding response: %v", err)
	}

	amountOut, err := strconv.ParseFloat(osmosisResp.AmountOut, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing amount_out: %v", err)
	}

	return amountOut, nil
}
