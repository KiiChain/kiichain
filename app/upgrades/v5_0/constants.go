package v500

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/kiichain/kiichain/v5/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v5.0.0"
)

// Upgrade defines the upgrade
// This adds the fee abstraction module store key
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Deleted: []string{
			"crisis",
			"capability",
			"feeibc",
		},
	},
}
