# Arb Bot

Osmosis arbitrage bot agains Binance.

Uses top-of-block auction.

## Setup

Create a `.env` file with the following content:
```bash
GRPC_ADDRESS=http://localhost:26657
OSMOSIS_ACCOUNT_KEY=your_key_here
SQS_OSMOSIS_API_KEY=your_key_here
```

```
go run main.go
```
