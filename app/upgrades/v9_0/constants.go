package v90

import (
	"github.com/kiichain/kiichain/v7/app/upgrades"
)

const (
	// UpgradeName is the name of the upgrade
	UpgradeName = "v9.0.0"
)

// Upgrade defines the coordinated upgrade that moves the chain onto
// cosmos/evm v0.7.3, cosmos-sdk v0.54, ibc-go v11, and wasmd v0.70.
// No persistent stores are added or removed, so StoreUpgrades is empty.
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
