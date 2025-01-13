package keeper_test

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kiichain/kiichain3/x/tokenfactory/types"
)

func (suite *KeeperTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		FactoryDenoms: []types.GenesisDenom{
			{
				Denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/bitcoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
				},
			},
			{
				Denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/diff-admin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "kii1hjfwcza3e3uzeznf3qthhakdr9juetl7uajv0t",
				},
			},
			{
				Denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/litecoin",
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
				},
			},
		},
	}
	app := suite.App
	suite.Ctx = app.BaseApp.NewContext(false, tmproto.Header{})
	// Test both with bank denom metadata set, and not set.
	for i, denom := range genesisState.FactoryDenoms {
		// hacky, sets bank metadata to exist if i != 0, to cover both cases.
		if i != 0 {
			app.BankKeeper.SetDenomMetaData(suite.Ctx, banktypes.Metadata{Base: denom.GetDenom()})
		}
	}

	app.TokenFactoryKeeper.InitGenesis(suite.Ctx, genesisState)
	exportedGenesis := app.TokenFactoryKeeper.ExportGenesis(suite.Ctx)
	suite.Require().NotNil(exportedGenesis)
	suite.Require().Equal(genesisState, *exportedGenesis)
}
