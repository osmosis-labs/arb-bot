package src

import (
	"context"
	"encoding/hex"
	"log"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/osmosis-labs/osmosis/v25/app"
	"github.com/osmosis-labs/osmosis/v25/app/params"
	"google.golang.org/grpc"
)

type SeedConfig struct {
	ChainID        string
	GRPCConnection *grpc.ClientConn
	EncodingConfig params.EncodingConfig
	Key            *secp256k1.PrivKey
	DenomMap       map[string]string
}

const (
	GRPC_ADDRESS = "http://grpc.osmosis.zone:9090/"
	CHAIN_ID     = "osmosis-1"
)

var (
	apiKeyHex  string
	seedConfig SeedConfig
)

func OsmosisInit() (SeedConfig, error) {
	conn := CreateGRPCConnection(GRPC_ADDRESS)
	encCfg := app.MakeEncodingConfig()

	apiKeyHex = os.Getenv("OSMOSIS_ACCOUNT_KEY")
	bz, err := hex.DecodeString(apiKeyHex)
	if err != nil {
		return SeedConfig{}, err
	}
	privKey := &secp256k1.PrivKey{Key: bz}

	seedConfig = SeedConfig{
		ChainID:        CHAIN_ID,
		GRPCConnection: conn,
		EncodingConfig: encCfg,
		Key:            privKey,
	}

	return seedConfig, nil
}

// CreateGRPCConnection createa a grpc connection to a given url
func CreateGRPCConnection(addr string) *grpc.ClientConn {
	const GrpcConnectionTimeoutSeconds = 1000

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(GrpcConnectionTimeoutSeconds)*time.Millisecond)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)

	if err != nil {
		// TODO: fix this wart, handle the error gracefully somewhere
		// We run GRPCConnection for the edge node and the local node we ignore errors here unil this refactor
		log.Println(
			"Local client not connected, this error is here as you've attempted to run local seeds with no local node running")

		return nil
	}

	return conn
}
