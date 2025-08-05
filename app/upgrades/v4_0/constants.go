package v400

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/kiichain/kiichain/v3/app/upgrades"
	feeabstractiontypes "github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v4.0.0"
)

// Upgrade defines the upgrade
// This adds the fee abstraction module store key
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			feeabstractiontypes.StoreKey,
		},
	},
}
