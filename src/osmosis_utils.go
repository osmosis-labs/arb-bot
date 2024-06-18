package src

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authenticatortypes "github.com/osmosis-labs/osmosis/v25/x/smart-account/types"

	"github.com/osmosis-labs/osmosis/v25/app/params"
)

func SignAuthenticatorMsgMultiSignersBytes(
	senderPrivKeys []cryptotypes.PrivKey,
	signerPrivKeys []cryptotypes.PrivKey,
	cosignerPrivKeys map[int][]cryptotypes.PrivKey,
	encCfg params.EncodingConfig,
	tm tmservice.ServiceClient,
	ac authtypes.QueryClient,
	txClient txtypes.ServiceClient,
	chainID string,
	msgs []sdk.Msg,
	selectedAuthenticators []uint64,
	sequenceOffset uint64,
) ([]byte, error) {
	log.Println("Creating signed txn to include in bundle")

	var accNums []uint64
	var accSeqs []uint64

	for _, privKey := range senderPrivKeys {
		// Generate the account address from the private key
		addr := sdk.AccAddress(privKey.PubKey().Address()).String()

		// Get the account information
		res, err := ac.Account(
			context.Background(),
			&authtypes.QueryAccountRequest{Address: addr},
		)
		if err != nil {
			return nil, err
		}

		var acc authtypes.AccountI
		if err := encCfg.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
			return nil, err
		}

		log.Println("Signer account: " + acc.GetAddress().String())
		accNums = append(accNums, acc.GetAccountNumber())
		// XXX: here we return + 1 to offset the seq
		accSeqs = append(accSeqs, acc.GetSequence()+sequenceOffset)
	}

	block, err := tm.GetLatestBlock(context.Background(), &tmservice.GetLatestBlockRequest{})
	if err != nil {
		return nil, err
	}

	// Sign the message
	txBytes, _ := SignAuthenticatorMsgWithHeight(
		encCfg.TxConfig,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(FEE_DENOM, 7000)},
		1700000,
		chainID,
		accNums,
		accSeqs,
		senderPrivKeys,
		signerPrivKeys,
		cosignerPrivKeys,
		selectedAuthenticators,
		uint64(block.Block.Header.Height)+1,
	)

	return txBytes, nil
}

// GenTx generates a signed mock transaction.
func SignAuthenticatorMsgWithHeight(
	gen client.TxConfig,
	msgs []sdk.Msg,
	feeAmt sdk.Coins,
	gas uint64,
	chainID string,
	accNums, accSeqs []uint64,
	signers, signatures []cryptotypes.PrivKey,
	cosigners map[int][]cryptotypes.PrivKey,
	selectedAuthenticators []uint64,
	timeoutHeight uint64,
) ([]byte, error) {
	sigs := make([]signing.SignatureV2, len(signers))

	// create a random length memo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))
	signMode := gen.SignModeHandler().DefaultMode()

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range signers {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
			Sequence: accSeqs[i],
		}
	}

	baseTxBuilder := gen.NewTxBuilder()

	txBuilder, ok := baseTxBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, fmt.Errorf("expected authtx.ExtensionOptionsTxBuilder, got %T", baseTxBuilder)
	}
	if len(selectedAuthenticators) > 0 {
		value, err := types.NewAnyWithValue(&authenticatortypes.TxExtension{
			SelectedAuthenticators: selectedAuthenticators,
		})
		if err != nil {
			return nil, err
		}
		txBuilder.SetNonCriticalExtensionOptions(value)
	}

	err := txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = txBuilder.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(feeAmt)
	txBuilder.SetGasLimit(gas)
	txBuilder.SetTimeoutHeight(timeoutHeight)
	// TODO: set fee payer

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range signatures {
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		signBytes, err := gen.SignModeHandler().GetSignBytes(
			signMode,
			signerData,
			txBuilder.GetTx(),
		)
		if err != nil {
			return nil, err
		}

		// Assuming cosigners is already initialized and populated
		var compoundSignatures [][]byte
		if value, exists := cosigners[1]; exists {
			for _, item := range value {
				sig, err := item.Sign(signBytes)
				if err != nil {
					return nil, err
				}
				compoundSignatures = append(compoundSignatures, sig)
			}
		}

		// Marshalling the array of SignatureV2 for compound signatures
		if len(compoundSignatures) >= 1 {
			compoundSigData, err := json.Marshal(compoundSignatures)
			if err != nil {
				return nil, err
			}
			sigs[i] = signing.SignatureV2{
				PubKey: signers[i].PubKey(),
				Data: &signing.SingleSignatureData{
					SignMode:  signMode,
					Signature: compoundSigData,
				},
				Sequence: accSeqs[i],
			}
		} else {
			sig, err := p.Sign(signBytes)
			if err != nil {
				return nil, err
			}
			sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
		}

		err = txBuilder.SetSignatures(sigs...)
		if err != nil {
			return nil, err
		}
	}

	txBytes, err := gen.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

func SignAndBroadcastAuthenticatorMsgMultiSignersWithBlock(
	senderPrivKeys []cryptotypes.PrivKey,
	signerPrivKeys []cryptotypes.PrivKey,
	cosignerPrivKeys map[int][]cryptotypes.PrivKey,
	encCfg params.EncodingConfig,
	tm tmservice.ServiceClient,
	ac authtypes.QueryClient,
	txClient txtypes.ServiceClient,
	chainID string,
	msgs []sdk.Msg,
	selectedAuthenticators []uint64,
) error {
	log.Println("Signing and broadcasting message flow")

	var accNums []uint64
	var accSeqs []uint64

	for _, privKey := range senderPrivKeys {
		// Generate the account address from the private key
		addr := sdk.AccAddress(privKey.PubKey().Address()).String()

		// Get the account information
		res, err := ac.Account(
			context.Background(),
			&authtypes.QueryAccountRequest{Address: addr},
		)
		if err != nil {
			return err
		}

		var acc authtypes.AccountI
		if err := encCfg.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
			return err
		}

		log.Println("Signer account: " + acc.GetAddress().String())
		accNums = append(accNums, acc.GetAccountNumber())
		accSeqs = append(accSeqs, acc.GetSequence())
	}

	block, err := tm.GetLatestBlock(context.Background(), &tmservice.GetLatestBlockRequest{})
	if err != nil {
		return err
	}

	// Sign the message
	txBytes, _ := SignAuthenticatorMsgWithHeight(
		encCfg.TxConfig,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(FEE_DENOM, 7000)},
		1700000,
		chainID,
		accNums,
		accSeqs,
		senderPrivKeys,
		signerPrivKeys,
		cosignerPrivKeys,
		selectedAuthenticators,
		uint64(block.Block.Header.Height)+1,
	)

	resp, err := txClient.BroadcastTx(
		context.Background(),
		&txtypes.BroadcastTxRequest{
			Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		return err
	}
	log.Println("Transaction Hash:", resp.TxResponse.TxHash)
	if resp.TxResponse.RawLog != "" {
		log.Println("Transaction failed reason:", resp.TxResponse.RawLog)
	}

	time.Sleep(6 * time.Second)

	tx, err := txClient.GetTx(
		context.Background(),
		&txtypes.GetTxRequest{
			Hash: resp.TxResponse.TxHash,
		},
	)
	if err != nil {
		return err
	} else {
		if tx.TxResponse.Code == 0 {
			log.Println("Transaction Success...")
		} else {
			log.Println(tx.TxResponse)
		}
	}
	log.Println("Gas Used:", tx.TxResponse.GasUsed)

	return nil
}
