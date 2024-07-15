package src

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	auctiontypes "github.com/skip-mev/block-sdk/x/auction/types"

	"github.com/osmosis-labs/arb-bot/src/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

// Note that the amount here should be in human readable exponent
// E.g) getting usdc price of 1 bitcoin would be GetOsmosisBTCToUSDCPrice(1)
func GetOsmosisBTCToUSDCPriceAndRoute(tokenInAmount float64, baseDenom, quoteDenom string, baseExponent, quoteExponent int) (float64, []poolmanagertypes.SwapAmountInSplitRoute, error) {
	fmt.Printf("tokenInAmount: %f, baseExponent: %d\n", tokenInAmount, baseExponent)
	amountWithExponentApplied := int64(tokenInAmount * math.Pow(10, float64(baseExponent)))

	btcExecutionPrice, route, err := getOsmosisPriceAndRoute(baseDenom, quoteDenom, amountWithExponentApplied)
	if err != nil {
		return 0, []poolmanagertypes.SwapAmountInSplitRoute{}, err
	}

	return btcExecutionPrice * math.Pow(10, float64(baseExponent)-float64(quoteExponent)), route, nil
}

func GetOsmosisUSDCToBTCPriceAndRoute(tokenInAmount float64, baseDenom, quoteDenom string, baseExponent, quoteExponent int) (float64, []poolmanagertypes.SwapAmountInSplitRoute, error) {
	amountWithExponentApplied := int64(tokenInAmount * math.Pow(10, float64(quoteExponent)))

	usdcPrice, route, err := getOsmosisPriceAndRoute(quoteDenom, baseDenom, amountWithExponentApplied)
	if err != nil {
		return 0, []poolmanagertypes.SwapAmountInSplitRoute{}, err
	}
	return usdcPrice / math.Pow(10, float64(baseExponent)-float64(quoteExponent)), route, nil
}

type QuoteResponse struct {
	AmountOut string  `json:"amount_out"`
	Route     []Route `json:"route"`
}

type Route struct {
	Pools    []Pool `json:"pools"`
	InAmount string `json:"in_amount"`
}

type Pool struct {
	PoolId        uint64 `json:"id"`
	TokenOutdenom string `json:"token_out_denom"`
}

func getOsmosisPriceAndRoute(tokenInDenom, tokenOutDenom string, tokenInAmount int64) (float64, []poolmanagertypes.SwapAmountInSplitRoute, error) {
	url := fmt.Sprintf("%s?tokenIn=%d%s&tokenOutDenom=%s&humanDenoms=false", osmosisQuoteAPI, tokenInAmount, tokenInDenom, tokenOutDenom)

	// Define the headers
	headers := map[string]string{
		"x-api-key": os.Getenv("SQS_OSMOSIS_API_KEY"),
	}

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0, []poolmanagertypes.SwapAmountInSplitRoute{}, fmt.Errorf("error fetching price from Osmosis: %v", err)
	}

	// Add headers to the request
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error url: %s\n", url)
		return 0, []poolmanagertypes.SwapAmountInSplitRoute{}, fmt.Errorf("error fetching price from Osmosis: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("error url: %s\n", url)
		return 0, []poolmanagertypes.SwapAmountInSplitRoute{}, fmt.Errorf("error fetching price from Osmosis: status code %d", resp.StatusCode)
	}

	var quoteResponse QuoteResponse
	err = json.NewDecoder(resp.Body).Decode(&quoteResponse)
	if err != nil {
		return 0, []poolmanagertypes.SwapAmountInSplitRoute{}, fmt.Errorf("error decoding response: %v", err)
	}

	// manually unmarshal struct returned from sqs to poolmanager type struct
	// we can't directly unmarshal into pool manager's SwapAmountInSplitRoute
	// due to json tag differences
	route := make([]poolmanagertypes.SwapAmountInSplitRoute, len(quoteResponse.Route))

	for i, quotedRoute := range quoteResponse.Route {
		inAmount, ok := sdk.NewIntFromString(quotedRoute.InAmount)
		if !ok {
			return 0, []poolmanagertypes.SwapAmountInSplitRoute{}, fmt.Errorf("error parsing in amount from endpoint")
		}
		route[i].TokenInAmount = inAmount

		// Initialize the Pools slice for the current route with the length of quotedRoute.Pools
		route[i].Pools = make([]poolmanagertypes.SwapAmountInRoute, len(quotedRoute.Pools))

		for j, pool := range quotedRoute.Pools {
			route[i].Pools[j].PoolId = pool.PoolId
			route[i].Pools[j].TokenOutDenom = pool.TokenOutdenom
		}
	}

	amountOut, err := strconv.ParseFloat(quoteResponse.AmountOut, 64)
	if err != nil {
		return 0, []poolmanagertypes.SwapAmountInSplitRoute{}, fmt.Errorf("error parsing amount_out: %v", err)
	}

	return amountOut / float64(tokenInAmount), route, nil
}

// GetOsmosisBTCUSDTalance returns the balances in human readable exponents
func GetOsmosisBaseQuoteBalance(seedConfig SeedConfig, arbMetadata domain.OsmoBinanceArbPairMetadata) (float64, float64, error) {
	grpcConnection := seedConfig.GRPCConnection
	senderAddress := sdk.AccAddress(seedConfig.Key.PubKey().Address())

	bankClient := banktypes.NewQueryClient(grpcConnection)
	usdcBalanceResponse, err := bankClient.Balance(
		context.Background(),
		&banktypes.QueryBalanceRequest{Address: senderAddress.String(), Denom: arbMetadata.QuoteChainDenom},
	)

	if err != nil {
		return 0, 0, err
	}
	baseBalanceResponse, err := bankClient.Balance(
		context.Background(),
		&banktypes.QueryBalanceRequest{Address: senderAddress.String(), Denom: arbMetadata.BaseChainDenom},
	)

	if err != nil {
		return 0, 0, err
	}
	usdcAmount := usdcBalanceResponse.Balance.Amount
	btcAmount := baseBalanceResponse.Balance.Amount

	usdcAmountWithExponent := float64(usdcAmount.Int64()) / math.Pow(10, float64(arbMetadata.ExponentQuote))
	btcAmountWithExponent := float64(btcAmount.Int64()) / math.Pow(10, float64(arbMetadata.ExponentBase))
	return usdcAmountWithExponent, btcAmountWithExponent, nil
}

func BuyOsmosisBase(seedConfig SeedConfig, baseDenom string, route []poolmanagertypes.SwapAmountInSplitRoute, binanceBTCPrice float64) error {
	return SwapWithTopOfBlockAuction(seedConfig, route, baseDenom, 1)
}

func SellOsmosisBase(seedConfig SeedConfig, baseDenom string, route []poolmanagertypes.SwapAmountInSplitRoute, binanceBTCPrice float64) error {
	return SwapWithTopOfBlockAuction(seedConfig, route, baseDenom, 1)
}

func SwapWithTopOfBlockAuction(seedConfig SeedConfig,
	route []poolmanagertypes.SwapAmountInSplitRoute,
	tokenInDenom string,
	tokenOutMinAmount uint64,
) error {
	grpcConnection := seedConfig.GRPCConnection
	senderAddress := sdk.AccAddress(seedConfig.Key.PubKey().Address())
	txClient := txtypes.NewServiceClient(grpcConnection)
	ac := auth.NewQueryClient(grpcConnection)
	tm := tmservice.NewServiceClient(grpcConnection)

	swapTokenMsg := &poolmanagertypes.MsgSplitRouteSwapExactAmountIn{
		Sender:            senderAddress.String(),
		Routes:            route,
		TokenInDenom:      tokenInDenom,
		TokenOutMinAmount: sdk.NewIntFromUint64(tokenOutMinAmount),
	}

	txBytes1, err := SignAuthenticatorMsgMultiSignersBytes(
		[]cryptotypes.PrivKey{seedConfig.Key},
		[]cryptotypes.PrivKey{seedConfig.Key},
		nil,
		seedConfig.EncodingConfig,
		tm,
		ac,
		txClient,
		seedConfig.ChainID,
		[]sdk.Msg{swapTokenMsg},
		[]uint64{},
		1,
	)

	if err != nil {
		return err
	}

	bundle := [][]byte{txBytes1}

	bidMsg := &auctiontypes.MsgAuctionBid{
		Bidder:       senderAddress.String(),
		Bid:          sdk.NewCoin(BidDenom, sdk.NewInt(100)),
		Transactions: bundle,
	}

	err = SignAndBroadcastAuthenticatorMsgMultiSignersWithBlock(
		[]cryptotypes.PrivKey{seedConfig.Key},
		[]cryptotypes.PrivKey{seedConfig.Key},
		nil,
		seedConfig.EncodingConfig,
		tm,
		ac,
		txClient,
		seedConfig.ChainID,
		[]sdk.Msg{bidMsg},
		[]uint64{},
	)
	if err != nil {
		return err
	}

	return nil
}
