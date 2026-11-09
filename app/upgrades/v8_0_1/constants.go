package v801

import (
	"github.com/kiichain/kiichain/v7/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v8.0.1"
)

// Upgrade defines the coordinated upgrade that ships the v8.0.1 binary.
// No store migrations are required; the handler only runs pending module
// migrations so validators switch binaries at the same height.
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
