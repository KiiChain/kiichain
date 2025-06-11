package v200

import (
	"github.com/kiichain/kiichain/v1/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v2.0.0"
)

// Upgrade is the upgrade object
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
