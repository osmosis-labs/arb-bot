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

func getBalance(SeedConfig) error {
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

	//	sequenceOffset := uint64(1)
	bidMsg := &auctiontypes.MsgAuctionBid{
		Bidder:       senderAddress.String(),
		Bid:          sdk.NewCoin(USDCDenom, sdk.NewInt(1000000)),
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
