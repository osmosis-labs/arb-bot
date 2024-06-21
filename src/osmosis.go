package src

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	auctiontypes "github.com/skip-mev/block-sdk/x/auction/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

// Note that the amount here should be in human readable exponent
// E.g) getting usdc price of 1 bitcoin would be GetOsmosisBTCToUSDCPrice(1)
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

	var quoteResponse QuoteResponse
	err = json.NewDecoder(resp.Body).Decode(&quoteResponse)
	if err != nil {
		return 0, fmt.Errorf("error decoding response: %v", err)
	}

	// manually unmarshal struct returned from sqs to poolmanager type struct
	// we can't directly unmarshal into pool manager's SwapAmountInSplitRoute
	// due to json tag differences
	route := make([]poolmanagertypes.SwapAmountInSplitRoute, len(quoteResponse.Route))

	for i, quotedRoute := range quoteResponse.Route {
		inAmount, ok := sdk.NewIntFromString(quotedRoute.InAmount)
		if !ok {
			return 0, fmt.Errorf("error parsing in amount from endpoint")
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
		return 0, fmt.Errorf("error parsing amount_out: %v", err)
	}

	return amountOut, nil
}

func GetBalance(SeedConfig) error {
	grpcConnection := seedConfig.GRPCConnection
	senderAddress := sdk.AccAddress(seedConfig.Key.PubKey().Address())

	bankClient := banktypes.NewQueryClient(grpcConnection)
	balancePre, err := bankClient.Balance(
		context.Background(),
		&banktypes.QueryBalanceRequest{Address: senderAddress.String(), Denom: "uosmo"},
	)

	if err != nil {
		return err
	}

	fmt.Println(balancePre)
	return nil
}

func TopOfBlockAuction(seedConfig SeedConfig, from sdk.Coin, to sdk.Coin) error {
	grpcConnection := seedConfig.GRPCConnection
	senderAddress := sdk.AccAddress(seedConfig.Key.PubKey().Address())
	txClient := txtypes.NewServiceClient(grpcConnection)
	ac := auth.NewQueryClient(grpcConnection)
	tm := tmservice.NewServiceClient(grpcConnection)

	// TODO: implement this with SQS Routes
	swapTokenMsg := &poolmanagertypes.MsgSwapExactAmountOut{
		Sender: senderAddress.String(),
		Routes: []poolmanagertypes.SwapAmountOutRoute{
			{
				PoolId:       1265,
				TokenInDenom: from.Denom,
			},
		},
		TokenInMaxAmount: osmomath.NewInt(1000000000),
		TokenOut:         sdk.NewCoin(to.Denom, osmomath.NewInt(10000000)),
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
