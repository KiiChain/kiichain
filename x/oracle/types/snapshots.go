package types

// Custom type used in params.proto. PriceSnapshotItems represents an array of PriceSnapshotItem
type PriceSnapshotItems []PriceSnapshotItem

// PriceSnapshots represents an array of PriceSnapshot on query.go
type PriceSnapshots []PriceSnapshot

// OracleTwaps represents an array of OracleTwap on query.go
type OracleTwaps []OracleTwap

// Constructor functions
// NewPriceSnapshot creates a new instance of PriceSnapshot
func NewPriceSnapshot(snapshotTimestamp int64, priceSnapshotItems PriceSnapshotItems) PriceSnapshot {
	return PriceSnapshot{
		SnapshotTimestamp:  snapshotTimestamp,
		PriceSnapshotItems: priceSnapshotItems,
	}
}

// NewPriceSnapshotItem creates a new instance of PriceSnapshotItem
func NewPriceSnapshotItem(denom string, exchangeRate OracleExchangeRate) PriceSnapshotItem {
	return PriceSnapshotItem{
		Denom:              denom,
		OracleExchangeRate: exchangeRate,
	}
}
