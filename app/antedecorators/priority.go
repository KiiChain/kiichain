package antedecorators

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	EVMAssociatePriority = math.MaxInt64 - 101
	// This is the max priority a non oracle or associate tx can take
	MaxPriority = math.MaxInt64 - 1000
)

type PriorityDecorator struct{}

func NewPriorityDecorator() PriorityDecorator {
	return PriorityDecorator{}
}

func intMin(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// Assigns higher priority to certain types of transactions including oracle
func (pd PriorityDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// Cap priority
	// Use higher priorities for tiers including oracle tx's
	priority := intMin(ctx.Priority(), MaxPriority)

	newCtx := ctx.WithPriority(priority)

	return next(newCtx, tx, simulate)
}
