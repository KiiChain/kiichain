package v730

import (
	"github.com/kiichain/kiichain/v7/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v7.3.0"
)

// Upgrade defines the coordinated upgrade for the EVM v0.6.0-fork.3 dependency
// bump. The fork.3 changes (precompile gas accounting, statedb locked-balance
// handling, and feemarket EndBlock) are state-machine-breaking, so this upgrade
// ensures every validator switches to the new binary at the same height. No
// store migrations are required.
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
