package v731

import (
	"github.com/kiichain/kiichain/v7/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v7.3.1"
)

// Upgrade defines the coordinated upgrade that ships the gov module bank-block
// fix so EVM gov-precompile deposits no longer fail StateDB mirroring on commit.
// No store migrations are required; the handler only runs pending module
// migrations so validators switch binaries at the same height.
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
