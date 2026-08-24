package v740

import "github.com/kiichain/kiichain/v7/app/upgrades"

const (
	// UpgradeName is the on-chain identifier for this emergency upgrade Plan.
	UpgradeName = "v7.4.0"

	// UpgradeHeight is the height at which the Plan is scheduled and applied
	// within the same PreBlocker pass — the first height produced after the
	// manual halt (H+1). Mainnet's last committed height before the halt was
	// 9355723, per `latest_block_height` on the halted node's RPC.
	UpgradeHeight = int64(9355724)

	// MainnetChainID is the only chain-id this emergency upgrade is allowed to
	// move funds on, confirmed via the halted mainnet node's own RPC status
	// ("network": "kiichain_1783-1").
	MainnetChainID = "kiichain_1783-1"
)

// Upgrade registers the emergency fund-recovery handler with x/upgrade. The
// Plan itself is scheduled programmatically from app.PreBlocker (see app.go)
// instead of via governance, so no on-chain vote is required.
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
