package v90

import (
	"github.com/kiichain/kiichain/v7/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v9.0.0"
)

// Upgrade runs pending module migrations, including x/rewards consensus v1 to v2.
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
