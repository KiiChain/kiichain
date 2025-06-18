package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v2/x/rewards/types"
)

type GenesisTestSuite struct {
	suite.Suite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestNewGenesisState() {
	// Create test data
	params := types.DefaultParams()
	pool := types.InitialRewardPool()
	schedule := types.InitialReleaseSchedule()

	// Test creation
	genesis := types.NewGenesisState(params, pool, schedule)

	suite.Require().Equal(params, genesis.Params)
	suite.Require().Equal(pool, genesis.RewardPool)
	suite.Require().Equal(schedule, genesis.ReleaseSchedule)
}

func (suite *GenesisTestSuite) TestDefaultGenesisState() {
	defaultGenesis := types.DefaultGenesisState()

	suite.Require().Equal(types.DefaultParams(), defaultGenesis.Params)
	suite.Require().Equal(types.InitialRewardPool(), defaultGenesis.RewardPool)
	suite.Require().Equal(types.InitialReleaseSchedule(), defaultGenesis.ReleaseSchedule)
}

func (suite *GenesisTestSuite) TestValidateGenesis() {
	validParams := types.DefaultParams()
	validPool := types.InitialRewardPool()
	validSchedule := types.ReleaseSchedule{
		TotalAmount:     sdk.NewCoin("akii", math.NewInt(1000)),
		ReleasedAmount:  sdk.NewCoin("akii", math.NewInt(0)),
		EndTime:         time.Now().Add(time.Hour * 24),
		LastReleaseTime: time.Time{},
		Active:          true,
	}

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
			name: "custom valid genesis",
			modifyFn: func(gs *types.GenesisState) {
				gs.Params = validParams
				gs.RewardPool = validPool
				gs.ReleaseSchedule = validSchedule
			},
			expectedPass: true,
		},
		{
			name: "invalid params",
			modifyFn: func(gs *types.GenesisState) {
				gs.Params.TokenDenom = "" // invalid empty denom
			},
			expectedPass: false,
		},
		{
			name: "invalid release schedule - past end time",
			modifyFn: func(gs *types.GenesisState) {
				gs.ReleaseSchedule.EndTime = time.Now().Add(-time.Hour)
			},
			expectedPass: false,
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
