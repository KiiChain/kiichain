package invariants

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

type KeeperInterface interface {
	ReleaseScheduleGet(ctx sdk.Context) (types.ReleaseSchedule, error)
	RewardPoolGet(ctx sdk.Context) (types.RewardPool, error)
	ParamsGet(ctx sdk.Context) (types.Params, error)
}