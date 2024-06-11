package src

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type OsmosisResponse struct {
	AmountOut string `json:"amount_out"`
}

func getOsmosisPrice(tokenIn, tokenOut, tokenInAmount string) (float64, error) {
	url := fmt.Sprintf("%s?tokenIn=%s%s&tokenOutDenom=%s&humanDenoms=false", osmosisQuoteAPI, tokenInAmount, tokenIn, tokenOut)
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error fetching price from Osmosis: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error fetching price from Osmosis: status code %d", resp.StatusCode)
	}

	var osmosisResp OsmosisResponse
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

func GetOsmosisBTCToUSDCPrice(tokenInAmount string) (float64, error) {
	return getOsmosisPrice(BTCDenom, USDCDenom, tokenInAmount)
}

func GetOsmosisUSDCToBTCPrice(tokenInAmount string) (float64, error) {
	return getOsmosisPrice(USDCDenom, BTCDenom, tokenInAmount)
}
