package keeper_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/kiichain/kiichain3/x/tokenfactory/keeper"
	"github.com/kiichain/kiichain3/x/tokenfactory/types"
)

func (suite *KeeperTestSuite) TestDenomMetadataRequest() {
	var (
		req         *types.QueryDenomMetadataRequest
		expMetadata = banktypes.Metadata{}
	)
	tokenFactoryDenom := "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/NEWCOIN"
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty denom",
			func() {
				req = &types.QueryDenomMetadataRequest{}
			},
			false,
		},
		{
			"not found denom",
			func() {
				req = &types.QueryDenomMetadataRequest{
					Denom: tokenFactoryDenom,
				}
			},
			false,
		},
		{
			"success",
			func() {

				expMetadata = banktypes.Metadata{
					Description: "Token factory custom token",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    tokenFactoryDenom,
							Exponent: 0,
							Aliases:  []string{tokenFactoryDenom},
						},
					},
					Base:    tokenFactoryDenom,
					Display: tokenFactoryDenom,
				}

				suite.App.BankKeeper.SetDenomMetaData(suite.Ctx, expMetadata)
				req = &types.QueryDenomMetadataRequest{
					Denom: expMetadata.Base,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.Ctx)

			res, err := suite.queryClient.DenomMetadata(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expMetadata, res.Metadata)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDenomAllowListRequest() {

	tokenFactoryDenom := "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/NEWCOIN"
	allowList := banktypes.AllowList{
		Addresses: []string{"kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs", "kii1k98zjg7scsmk6d4ye8hhrv3an6ppykvtwayzpy"},
	}
	type args struct {
		req *types.QueryDenomAllowListRequest
	}
	testCases := []struct {
		name          string
		args          args
		malleate      func()
		expAllowList  banktypes.AllowList
		expectedError string
		wantErr       bool
	}{
		{
			name:     "fails on empty denom",
			malleate: func() {},
			args: args{
				req: &types.QueryDenomAllowListRequest{},
			},
			expectedError: "rpc error: code = InvalidArgument desc = invalid denom",
			wantErr:       true,
		},
		{
			name:     "returns empty list for denom that does not have allow list",
			malleate: func() {},
			args: args{
				req: &types.QueryDenomAllowListRequest{
					Denom: tokenFactoryDenom,
				},
			},
			expAllowList: banktypes.AllowList{},
			wantErr:      false,
		},
		{
			name: "returns allow list for denom that has allow list",
			malleate: func() {
				suite.App.BankKeeper.SetDenomAllowList(suite.Ctx, tokenFactoryDenom, allowList)
			},
			args: args{
				req: &types.QueryDenomAllowListRequest{
					Denom: tokenFactoryDenom,
				},
			},
			expAllowList: allowList,
			wantErr:      false,
		},
	}

	for _, tc := range testCases {

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.Ctx)

			res, err := suite.queryClient.DenomAllowList(ctx, tc.args.req)

			if tc.wantErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expectedError)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(tc.expAllowList, res.AllowList)
			}
		})
	}
}

// TestQueryServerDenomsFromCreator tests the query server for the query DenomsFromCreator
func (suite *KeeperTestSuite) TestQueryServerDenomsFromCreator() {
	testCases := []struct {
		name        string
		args        *types.QueryDenomsFromCreatorRequest
		denomsRes   []string
		malleate    func()
		errContains string
	}{
		{
			name: "Pass - No data set",
			args: &types.QueryDenomsFromCreatorRequest{
				Creator: "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
			},
		},
		{
			name: "Pass - Data set",
			args: &types.QueryDenomsFromCreatorRequest{
				Creator: "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
			},
			malleate: func() {
				_, err := suite.App.TokenFactoryKeeper.CreateDenom(
					suite.Ctx,
					"kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
					"test",
				)
				suite.Require().NoError(err)
			},
			denomsRes: []string{"factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/test"},
		},
		{
			name: "Pass - Big dataset paginated",
			args: &types.QueryDenomsFromCreatorRequest{
				Creator: "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
				Pagination: &query.PageRequest{
					Limit:  3,
					Offset: 2,
				},
			},
			malleate: func() {
				for i := range 100 {
					_, err := suite.App.TokenFactoryKeeper.CreateDenom(
						suite.Ctx,
						"kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
						fmt.Sprintf("test%d", i),
					)
					suite.Require().NoError(err)
				}
			},
			// Offset will skip index 0 and 1 and start in alphabetical order, 10, 11, 12
			denomsRes: []string{
				"factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/test10",
				"factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/test11",
				"factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/test12",
			},
		},
		{
			name: "Error - No creator",
			args: &types.QueryDenomsFromCreatorRequest{
				Creator: "",
			},
			errContains: "creator address cannot be empty",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Restart the environment
			suite.SetupTest()
			// Wrap the context for the message server
			ctx := sdk.WrapSDKContext(suite.Ctx)

			// Change the system
			if tc.malleate != nil {
				tc.malleate()
			}

			// Run the query
			res, err := suite.queryClient.DenomsFromCreator(ctx, tc.args)

			// Check for error
			if tc.errContains == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tc.errContains)
				return
			}

			// Check the response
			suite.Require().Equal(tc.denomsRes, res.Denoms)
		})
	}
}

func TestKeeper_DenomAllowList(t *testing.T) {
	type args struct {
		req *types.QueryDenomAllowListRequest
		c   context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *types.QueryDenomAllowListResponse
		wantErr bool
		errMsg  string
	}{
		{
			name: "nil request",
			args: args{
				req: nil,
				c:   context.Background(),
			},
			wantErr: true,
			errMsg:  "rpc error: code = InvalidArgument desc = empty request",
		},
		{
			name: "empty denom",
			args: args{
				req: &types.QueryDenomAllowListRequest{},
				c:   context.Background(),
			},
			wantErr: true,
			errMsg:  "rpc error: code = InvalidArgument desc = invalid denom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := keeper.Keeper{}
			got, err := k.DenomAllowList(tt.args.c, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DenomAllowList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("DenomAllowList() error = %v, wantErr %v", err, tt.errMsg)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DenomAllowList() got = %v, want %v", got, tt.want)
			}
		})
	}
}
