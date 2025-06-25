package ante

import (
	storetypes "cosmossdk.io/store/types"
)

// NewNoConsumptionGasMeter creates a new gas meter that does not consume any gas
func NewNoConsumptionGasMeter() storetypes.GasMeter {
	return &noConsumptionGasMeter{}
}

// noConsumptionGasMeter is a GasMeter implementation that does not consume any gas
type noConsumptionGasMeter struct {
	consumed storetypes.Gas
}

// Type check for the noConsumptionGasMeter
var _ storetypes.GasMeter = (*noConsumptionGasMeter)(nil)

// GasConsumed implements the interface for GasMeter
func (g *noConsumptionGasMeter) GasConsumed() storetypes.Gas {
	return 0
}

// GasConsumedToLimit implements the interface for GasMeter
func (g *noConsumptionGasMeter) GasConsumedToLimit() storetypes.Gas {
	return 0
}

// GasRemaining implements the interface for GasMeter
func (g *noConsumptionGasMeter) GasRemaining() storetypes.Gas {
	return storetypes.Gas(0) - g.consumed
}

// Limit implements the interface for GasMeter
func (g *noConsumptionGasMeter) Limit() storetypes.Gas {
	return storetypes.Gas(0)
}

// ConsumeGas implements the interface for GasMeter
func (g *noConsumptionGasMeter) ConsumeGas(amount storetypes.Gas, descriptor string) {
}

// RefundGas implements the interface for GasMeter
func (g *noConsumptionGasMeter) RefundGas(amount storetypes.Gas, descriptor string) {
	if g.consumed < amount {
		panic(storetypes.ErrorNegativeGasConsumed{Descriptor: descriptor})
	}

	g.consumed -= amount
}

// IsPastLimit implements the interface for GasMeter
func (g *noConsumptionGasMeter) IsPastLimit() bool {
	return false
}

// IsOutOfGas implements the interface for GasMeter
func (g *noConsumptionGasMeter) IsOutOfGas() bool {
	return false
}

// String implements the interface for GasMeter
func (g *noConsumptionGasMeter) String() string {
	return "NoConsumptionGasMeter"
}
