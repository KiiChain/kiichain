package v710mainnet

import (
	"github.com/kiichain/kiichain/v7/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v7.1.0-mainnet"
)

// Upgrade defines the upgrade, nothing is done here
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
