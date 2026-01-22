package v611

import (
	"github.com/kiichain/kiichain/v6/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v6.1.1"
)

// Upgrade defines the upgrade, nothing is done here
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
