package v510

import (
	"github.com/kiichain/kiichain/v5/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v5.1.0"
)

// Upgrade defines the upgrade, nothing is done here
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: nil,
}
