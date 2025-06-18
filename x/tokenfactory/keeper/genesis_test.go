package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/kiichain/kiichain/v2/x/tokenfactory/types"
)

func (suite *KeeperTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		FactoryDenoms: []types.GenesisDenom{
			{
				Denom: "factory/kii1t7egva48prqmzl59x5ngv4zx0dtrwewc5thc4c/bitcoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "kii1t7egva48prqmzl59x5ngv4zx0dtrwewc5thc4c",
				},
			},
			{
				Denom: "factory/kii1t7egva48prqmzl59x5ngv4zx0dtrwewc5thc4c/diff-admin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "kii15czt5nhlnvayqq37xun9s9yus0d6y26dl40fz7",
				},
			},
			{
				Denom: "factory/kii1t7egva48prqmzl59x5ngv4zx0dtrwewc5thc4c/litecoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "kii1t7egva48prqmzl59x5ngv4zx0dtrwewc5thc4c",
				},
			},
		},
	}

	suite.SetupTestForInitGenesis()
	app := suite.App

	// Test both with bank denom metadata set, and not set.
	for i, denom := range genesisState.FactoryDenoms {
		// hacky, sets bank metadata to exist if i != 0, to cover both cases.
		if i != 0 {
			app.BankKeeper.SetDenomMetaData(suite.Ctx, banktypes.Metadata{Base: denom.GetDenom()})
		}
	}

	if err := app.TokenFactoryKeeper.SetParams(suite.Ctx, types.Params{DenomCreationFee: sdk.Coins{sdk.NewInt64Coin("stake", 100)}}); err != nil {
		panic(err)
	}
	app.TokenFactoryKeeper.InitGenesis(suite.Ctx, genesisState)

	exportedGenesis := app.TokenFactoryKeeper.ExportGenesis(suite.Ctx)
	suite.Require().NotNil(exportedGenesis)
	suite.Require().Equal(genesisState, *exportedGenesis)
}
