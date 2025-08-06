package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	kiichain "github.com/kiichain/kiichain/v3/app"
	"github.com/kiichain/kiichain/v3/app/helpers"
	"github.com/kiichain/kiichain/v3/x/feeabstraction/keeper"
	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// KeeperTestSuite is a test suite for the keeper package
type KeeperTestSuite struct {
	suite.Suite

	// Suite basics
	app *kiichain.KiichainApp
	ctx sdk.Context

	// The fee abstraction keeper
	keeper keeper.Keeper

	// The message service
	msgServer types.MsgServer
	querier   keeper.Querier
}

// TestKeeperTestSuite runs the keeper test suite
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// SetupTest initializes the test suite
func (suite *KeeperTestSuite) SetupTest() {
	// Setup both the app and the context
	app, ctx := helpers.SetupWithContext(suite.T())

	// Set the app and context
	suite.app = app
	suite.ctx = ctx

	// Set the keeper
	suite.keeper = app.FeeAbstractionKeeper

	// Create the message service and the querier
	suite.msgServer = keeper.NewMsgServer(suite.keeper)
	suite.querier = keeper.NewQuerier(suite.keeper)
}
