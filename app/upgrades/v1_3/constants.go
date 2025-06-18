package v130

import (
	"github.com/kiichain/kiichain/v2/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v1.3.0"
)

// Upgrade is the upgrade object
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
