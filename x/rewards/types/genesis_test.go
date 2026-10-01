package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

type GenesisTestSuite struct {
	suite.Suite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestNewGenesisState() {
	params := types.DefaultParams()
	pool := types.InitialRewardPool()

	genesis := types.NewGenesisState(params, pool)

	suite.Require().Equal(params, genesis.Params)
	suite.Require().Equal(pool, genesis.RewardPool)
}

func (suite *GenesisTestSuite) TestDefaultGenesisState() {
	defaultGenesis := types.DefaultGenesisState()

	suite.Require().Equal(types.DefaultParams(), defaultGenesis.Params)
	suite.Require().Equal(types.InitialRewardPool(), defaultGenesis.RewardPool)
}

func (suite *GenesisTestSuite) TestValidateGenesis() {
	testCases := []struct {
		name         string
		modifyFn     func(*types.GenesisState)
		expectedPass bool
	}{
		{
			name:         "default genesis - valid",
			modifyFn:     func(gs *types.GenesisState) {},
			expectedPass: true,
		},
		{
			name: "invalid params",
			modifyFn: func(gs *types.GenesisState) {
				gs.Params.TokenDenom = ""
			},
			expectedPass: false,
		},
		{
			name: "community pool with foreign denom",
			modifyFn: func(gs *types.GenesisState) {
				gs.RewardPool = types.RewardPool{
					CommunityPool: sdk.DecCoins{
						{Denom: "notkii", Amount: math.LegacyNewDec(100)},
					},
				}
			},
			expectedPass: false,
		},
		{
			name: "total released with foreign denom",
			modifyFn: func(gs *types.GenesisState) {
				gs.RewardPool.TotalReleased = sdk.NewCoin("notkii", math.NewInt(100))
			},
			expectedPass: false,
		},
		{
			name: "valid funded pool",
			modifyFn: func(gs *types.GenesisState) {
				gs.RewardPool = types.RewardPool{
					CommunityPool: sdk.DecCoins{
						{Denom: "akii", Amount: math.LegacyNewDec(100)},
					},
					TotalReleased: sdk.NewCoin("akii", math.NewInt(50)),
				}
			},
			expectedPass: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			genesis := types.DefaultGenesisState()
			tc.modifyFn(genesis)

			err := genesis.Validate()
			if tc.expectedPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
