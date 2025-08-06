package v310

import (
	storetypes "cosmossdk.io/store/types"

	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/kiichain/kiichain/v3/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v3.1.0"
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
