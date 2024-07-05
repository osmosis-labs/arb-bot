package domain

// OsmoBinanceArbPairMetadata is a struct that holds the metadata for a pair of tokens that we want to arbitrage between
type OsmoBinanceArbPairMetadata struct {
	BaseChainDenom  string
	QuoteChainDenom string
	ExponentBase    int
	ExponentQuote   int

	BinancePairTicker string

	RiskFactor float64
}
