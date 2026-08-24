package v732

import (
	"github.com/kiichain/kiichain/v7/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v7.3.2"
)

// Upgrade defines the coordinated upgrade that ships the August 2026 Cosmos EVM
// hotfix and enables the bank send restriction for the incident addresses.
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
