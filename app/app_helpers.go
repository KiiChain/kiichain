package kiichain

import (
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"
	icstest "github.com/cosmos/interchain-security/v7/testutil/integration"
)

// GetStakingKeeper implements the TestingApp interface. Needed for ICS.
func (app *KiichainApp) GetStakingKeeper() icstest.TestStakingKeeper { //nolint:nolintlint
	return app.StakingKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *KiichainApp) GetIBCKeeper() *ibckeeper.Keeper { //nolint:nolintlint
	return app.IBCKeeper
}
