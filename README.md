# Arb Bot

Osmosis arbitrage bot agains Binance.

Uses top-of-block auction.

## Setup

### Setup Keyring

```bash
osmosisd keys add test --keyring-backend file --recover

# Enter your mnemonic

# Confirm creation
ls $HOME/.osmosisd/keyring-file

# Use result for OSMOSIS_KEYRING_PATH 
```

### Setup .env

Create a `.env` file with the following content:
```bash
GRPC_ADDRESS=localhost:9090
OSMOSIS_KEYRING_PATH="/root/.osmosisd/keyring-file"
OSMOSIS_KEYRING_KEY_NAME=your_name
SQS_OSMOSIS_API_KEY=your_key
BINANCE_API_KEY=your_key
BINANCE_SECRET_KEY=yor_secret
```

```bash
go run main.go --password <your_keyring_password>
```
