package src

import (
	"os"

	"github.com/99designs/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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
	openKeyring, err := keyring.Open(keyringConfig)
	if err != nil {
		return SeedConfig{}, err
	}

	keyring, err := openKeyring.Get(os.Getenv("OSMOSIS_KEYRING_KEY_NAME"))
	if err != nil {
		return SeedConfig{}, err
	}

	privKey := &secp256k1.PrivKey{}

	if err := privKey.Unmarshal(keyring.Data); err != nil {
		return SeedConfig{}, err
	}

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
	const GrpcConnectionTimeoutMs = 5000

	grpcClient, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(50*1024*1024*1024)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(50*1024*1024*1024)),
	)

	if err != nil {
		return nil, err
	}

	return grpcClient, nil
}
