package utils

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	BlocksPerMinute = uint64(17)
	BlocksPerHour   = BlocksPerMinute * 60
	BlocksPerDay    = BlocksPerHour * 24
	BlocksPerWeek   = BlocksPerDay * 7
	BlocksPerMonth  = BlocksPerDay * 30
	BlocksPerYear   = BlocksPerDay * 365
)

// IsPeriodLastBlock checks if the block time on the context means the
// last block to finish the blocksPerPeriod
func IsPeriodLastBlock(ctx sdk.Context, blocksPerPeriod uint64) bool {
	nextBlockHeight := uint64(ctx.BlockHeight() + 1) // Get the next block height
	return nextBlockHeight%blocksPerPeriod == 0      // Check if the next block height is equal to the blocks per period
}
