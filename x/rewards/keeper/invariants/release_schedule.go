package invariants

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ReleaseScheduleInvariant(k KeeperInterface) func(ctx sdk.Context) (string, bool) {
	return func(ctx sdk.Context) (string, bool) {
		schedule, err := k.ReleaseScheduleGet(ctx)
		if err != nil {
			return fmt.Sprintf("failed to get release schedule: %v", err), false
		}

		if schedule.ReleasedAmount.Amount.GT(schedule.TotalAmount.Amount) {
			return fmt.Sprintf(
				"released amount (%s) > total amount (%s) for denom %s",
				schedule.ReleasedAmount, schedule.TotalAmount, schedule.TotalAmount.Denom,
			), true
		}

		if schedule.Active && !schedule.EndTime.After(schedule.LastReleaseTime) {
			return fmt.Sprintf(
				"active schedule has end time (%s) ≤ last release time (%s)",
				schedule.EndTime, schedule.LastReleaseTime,
			), true
		}

		if schedule.ReleasedAmount.Denom != schedule.TotalAmount.Denom {
			return fmt.Sprintf(
				"denom mismatch: released=%s, total=%s",
				schedule.ReleasedAmount.Denom, schedule.TotalAmount.Denom,
			), true
		}

		return "", false
	}
}
