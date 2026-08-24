package v740

import "github.com/kiichain/kiichain/v7/app/upgrades"

const (
	// UpgradeName is the on-chain identifier for this emergency upgrade Plan.
	UpgradeName = "emergency-fix-v1"

	// UpgradeHeight is the height at which the Plan is scheduled and applied
	// within the same PreBlocker pass — the first height produced after the
	// manual halt (H+1). The chain was stopped at height 1613 in the local
	// reproduction.
	UpgradeHeight = int64(325)
)

// Upgrade registers the emergency fund-recovery handler with x/upgrade. The
// Plan itself is scheduled programmatically from app.PreBlocker (see app.go)
// instead of via governance, so no on-chain vote is required.
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
