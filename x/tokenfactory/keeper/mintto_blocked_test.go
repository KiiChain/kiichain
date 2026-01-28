package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/v7/x/tokenfactory/types"
)

// TestMintToBlockedAddress proves that coins are minted before blocked address check
func (suite *KeeperTestSuite) TestMintToBlockedAddress() {
	suite.CreateDefaultDenom()

	// Get module account address
	moduleAddr := suite.App.AccountKeeper.GetModuleAddress(types.ModuleName)
	
	// Get initial module balance
	initialBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, moduleAddr, suite.defaultDenom)
	suite.T().Logf("Initial module balance: %s", initialBalance)

	// Get a blocked address (fee collector is typically blocked)
	blockedAddr := suite.App.AccountKeeper.GetModuleAddress("fee_collector")
	suite.Require().True(suite.App.BankKeeper.BlockedAddr(blockedAddr), "fee_collector should be blocked")

	// Try to mint to blocked address using MsgMintTo
	mintAmount := sdk.NewCoin(suite.defaultDenom, sdkmath.NewInt(1000))
	
	_, err := suite.msgServer.Mint(suite.Ctx, types.NewMsgMintTo(
		suite.TestAccs[0].String(),
		mintAmount,
		blockedAddr.String(),
	))

	// Should return error
	suite.Require().Error(err)
	suite.T().Logf("Error returned: %v", err)

	// Check module balance after error
	finalBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, moduleAddr, suite.defaultDenom)
	suite.T().Logf("Final module balance: %s", finalBalance)

	// BUG: If finalBalance > initialBalance, coins were minted but stuck
	if finalBalance.Amount.GT(initialBalance.Amount) {
		suite.T().Logf("BUG CONFIRMED: Module balance increased by %s", 
			finalBalance.Amount.Sub(initialBalance.Amount))
		suite.T().Log("Coins are stuck in module account!")
	}
}
