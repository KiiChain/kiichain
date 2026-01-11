package invariants

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ParamsInvariant(k KeeperInterface) func(ctx sdk.Context) (string, bool) {
	return func(ctx sdk.Context) (string, bool) {
		params, err := k.ParamsGet(ctx)
		if err != nil {
			return fmt.Sprintf("failed to get params: %v", err), false
		}

		if params.TokenDenom == "" {
			return "params: token denom is empty", true
		}

		return "", false
	}
}
