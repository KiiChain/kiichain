package invariants

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func CommunityPoolNonNegativeInvariant(k KeeperInterface) func(ctx sdk.Context) (string, bool) {
	return func(ctx sdk.Context) (string, bool) {
		pool, err := k.RewardPoolGet(ctx)
		if err != nil {
			return fmt.Sprintf("failed to get reward pool: %v", err), false
		}

		if pool.CommunityPool.IsAnyNegative() {
			return fmt.Sprintf(
				"community pool has negative amounts: %v",
				pool.CommunityPool,
			), true
		}

		return "", false
	}
}
