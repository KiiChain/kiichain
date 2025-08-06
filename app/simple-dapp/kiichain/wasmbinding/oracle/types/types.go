package types

// Query defines the structure for oracle queries
type Query struct {
	ExchangeRate  *ExchangeRateQuery  `json:"exchange_rate,omitempty"`
	ExchangeRates *ExchangeRatesQuery `json:"exchange_rates,omitempty"`
	Twaps         *TwapsQuery         `json:"twaps,omitempty"`
}

// ExchangeRateQuery defines the structure for querying a single exchange rate
type ExchangeRateQuery struct {
	Denom string `json:"denom"`
}

// ExchangeRatesQuery defines the structure for querying multiple exchange rates
type ExchangeRatesQuery struct{}

// TwapsQuery defines the structure for querying time-weighted average prices
type TwapsQuery struct {
	// LookbackSeconds is how much we should look back in seconds
	LookbackSeconds uint64 `json:"lookback_seconds"`
}
