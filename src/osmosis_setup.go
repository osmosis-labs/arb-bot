package src

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/99designs/keyring"
	sdkkeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v25/app"
	"github.com/osmosis-labs/osmosis/v25/app/params"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SeedConfig struct {
	ChainID        string
	GRPCConnection *grpc.ClientConn
	EncodingConfig params.EncodingConfig
	Key            *secp256k1.PrivKey
	DenomMap       map[string]string
}

const (
	CHAIN_ID  = "osmosis-1"
	FEE_DENOM = "uosmo"
)

var (
	apiKeyHex  string
	seedConfig SeedConfig
)

func OsmosisInit(keyringPassword string) (SeedConfig, error) {
	grpcAddress := os.Getenv("GRPC_ADDRESS")
	conn, err := CreateGRPCConnection(grpcAddress)
	if err != nil {
		return SeedConfig{}, err
	}
	encCfg := app.MakeEncodingConfig()

	keyringConfig := keyring.Config{
		ServiceName:              "osmosis",
		FileDir:                  os.Getenv("OSMOSIS_KEYRING_PATH"),
		KeychainTrustApplication: true,
		FilePasswordFunc: func(prompt string) (string, error) {
			return keyringPassword, nil
		},
	}

	// Open the keyring
	openKeyring, err := keyring.Open(keyringConfig)
	if err != nil {
		return SeedConfig{}, err
	}

	// Get the keyring record
	openRecord, err := openKeyring.Get(os.Getenv("OSMOSIS_KEYRING_KEY_NAME"))
	if err != nil {
		return SeedConfig{}, err
	}

	// Unmarshal the keyring record
	keyringRecord := new(sdkkeyring.Record)
	if err := keyringRecord.Unmarshal(openRecord.Data); err != nil {
		return SeedConfig{}, err
	}

	// Get the right type
	localRecord := keyringRecord.GetLocal()

	// Unmarshal the private key
	privKey := &secp256k1.PrivKey{}
	if err := privKey.Unmarshal(localRecord.PrivKey.Value); err != nil {
		return SeedConfig{}, err

	}

	// Get the address
	osmosisAddress := sdk.AccAddress(privKey.PubKey().Address())

	fmt.Println("your Osmosis address: ", osmosisAddress.String())

	seedConfig = SeedConfig{
		ChainID:        CHAIN_ID,
		GRPCConnection: conn,
		EncodingConfig: encCfg,
		Key:            privKey,
	}

	return seedConfig, nil
}

// CreateGRPCConnection createa a grpc connection to a given url
func CreateGRPCConnection(addr string) (*grpc.ClientConn, error) {
	const GrpcConnectionTimeoutSeconds = 5

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(GrpcConnectionTimeoutSeconds)*time.Second)
	defer cancel()
	grpcClient, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(50*1024*1024*1024)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(50*1024*1024*1024)),
	)

	if err != nil {
		return nil, err
	}

	return grpcClient, nil
}
