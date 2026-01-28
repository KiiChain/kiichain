package keeper_test

import (
sdkmath "cosmossdk.io/math"
sdk "github.com/cosmos/cosmos-sdk/types"
"github.com/kiichain/kiichain/v7/x/tokenfactory/types"
)

// TestMintToBlockedAddress ensures minting to a blocked address is rejected without minting coins
func (suite *KeeperTestSuite) TestMintToBlockedAddress() {
suite.CreateDefaultDenom()

// Get module account address
moduleAddr := suite.App.AccountKeeper.GetModuleAddress(types.ModuleName)

// Get initial module balance
initialBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, moduleAddr, suite.defaultDenom)

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

// Check module balance after error
finalBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, moduleAddr, suite.defaultDenom)

// Assert balance unchanged - no coins should be stuck
suite.Require().True(
finalBalance.Amount.Equal(initialBalance.Amount),
"module balance should remain unchanged when minting to a blocked address",
)
}
