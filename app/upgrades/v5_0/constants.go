package v500

import (
	storetypes "cosmossdk.io/store/types"

	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	"github.com/kiichain/kiichain/v5/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v5.0.0"
)

// Upgrade defines the upgrade
// It will delete the crisis module from the store
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Deleted: []string{
			crisistypes.ModuleName,
		},
	},
}
