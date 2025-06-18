package v300

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/kiichain/kiichain/v1/app/upgrades"
	oracletypes "github.com/kiichain/kiichain/v1/x/oracle/types"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v3.0.0"
)

// Upgrade defines the upgrade
// This adds the new precompile into the precompiles list for the EVM module
// And starts the oracle module store key
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			oracletypes.StoreKey,
		},
	},
}
