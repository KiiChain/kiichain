package v200

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/kiichain/kiichain/v1/app/upgrades"
	rewardtypes "github.com/kiichain/kiichain/v1/x/rewards/types"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v2.0.0"
)

// Upgrade is the upgrade object
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{rewardtypes.StoreKey},
	},
}
