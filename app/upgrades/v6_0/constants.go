package v600

import (
	"github.com/kiichain/kiichainv7/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v6.0.0"
)

// Upgrade defines the upgrade, nothing is done here
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
