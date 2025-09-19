package v500

import (
	storetypes "cosmossdk.io/store/types"

	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
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
			crisistypes.StoreKey,
			capabilitytypes.StoreKey,
			ibcfeetypes.StoreKey,
		},
	},
}
