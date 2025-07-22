package cosmos_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	cosmosevmante "github.com/cosmos/evm/ante/evm"
	"github.com/cosmos/evm/contracts"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/kiichain/kiichain/v3/app/apptesting"
	"github.com/kiichain/kiichain/v3/app/helpers"
	"github.com/kiichain/kiichain/v3/x/feeabstraction/ante/cosmos"
	"github.com/kiichain/kiichain/v3/x/feeabstraction/keeper"
)

const (
	DefaultFirstERC20      = "0x80b5a32E4F032B2a058b4F29EC95EEfEEB87aDcd"
	DefaultFirstERC20Denom = "erc20/" + DefaultFirstERC20
	DefaultMinFeeValue     = 875000000000000
)

// TestDeductFeeDecorator tests the DeductFeeDecorator
// This function tests the fee conversion and deduction logic
func TestDeductFeeDecorator(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create a fee payer
	founder := apptesting.RandomAccountAddress()

	// Create the funder account
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	// Set the different test cases
	testCases := []struct {
		name        string
		malleate    func(ctx sdk.Context)
		fee         sdk.Coins
		expected    sdk.Coins
		errContains string
		postCheck   func(ctx sdk.Context)
	}{
		{
			name: "success - valid fee deduction",
			malleate: func(ctx sdk.Context) {
				// Fun the account with enough funds to pay the fee
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
			},
			fee:      sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			expected: sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
		},
		{
			name: "success - zero fee",
			malleate: func(ctx sdk.Context) {
				// Get the current params on the feemarket module
				params := app.FeeMarketKeeper.GetParams(ctx)
				// Set the min gas price to zero
				params.BaseFee = math.LegacyZeroDec()
				// Set the params back
				err := app.FeeMarketKeeper.SetParams(ctx, params)
				require.NoError(t, err)
			},
			fee:      sdk.NewCoins(),
			expected: sdk.NewCoins(),
		},
		{
			name:        "fail - not enough funds",
			fee:         sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			expected:    sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			errContains: "insufficient funds for fee",
		},
		{
			name: "nonexistent fee payer",
			malleate: func(ctx sdk.Context) {
				// Get the funder account
				founderAcc := app.AccountKeeper.GetAccount(ctx, founder)
				// Remove the account
				app.AccountKeeper.RemoveAccount(ctx, founderAcc)
			},
			fee:         sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			expected:    sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			errContains: "does not exist",
		},
		// Now we start with the fee abstraction tests
		{
			name: "fee abstraction - fee conversion, native token",
			malleate: func(ctx sdk.Context) {
				// Set the token pair on the erc20 keeper
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Now we mint tokens for the fee payer
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
			},
			fee:      sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			expected: sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)),
			postCheck: func(ctx sdk.Context) {
				// Check the user balance, should be zero since all was user for fees
				balance := app.BankKeeper.GetBalance(ctx, founder, keeper.MockErc20Denom)
				require.True(t, balance.IsZero())
			},
		},
		{
			name: "fee abstraction - fee conversion, erc20 token",
			malleate: func(ctx sdk.Context) {
				// Deploy the erc20 token
				erc20Address, err := apptesting.DeployERC20(ctx, app)
				require.NoError(t, err)

				// Mint to the founder account
				err = apptesting.MintERC20(ctx, app, erc20Address, common.BytesToAddress(founder.Bytes()), big.NewInt(DefaultMinFeeValue))
				require.NoError(t, err)

				// Set the token pair on the erc20 keeper
				_, err = app.Erc20Keeper.RegisterERC20(ctx, &erc20types.MsgRegisterERC20{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Erc20Addresses: []string{
						erc20Address.Hex(),
					},
				})
				require.NoError(t, err)

				// Set the pair on the fee abstraction keeper
				erc20NativeAddress := "erc20/" + erc20Address.Hex()
				app.FeeAbstractionKeeper.SetFeePrices(ctx, []keeper.FeePrice{
					{
						Denom: erc20NativeAddress,
						Price: math.LegacyMustNewDecFromStr("0.5"),
					},
				})
			},
			fee:      sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			expected: sdk.NewCoins(sdk.NewInt64Coin(DefaultFirstERC20Denom, DefaultMinFeeValue/2)),
			postCheck: func(ctx sdk.Context) {
				// Check the user balance, should be zero on the native token
				balance := app.BankKeeper.GetBalance(ctx, founder, DefaultFirstERC20Denom)
				require.True(t, balance.IsZero())

				// Get the erc20 balance
				erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
				erc20Balance := app.Erc20Keeper.BalanceOf(
					ctx,
					erc20,
					common.HexToAddress(DefaultFirstERC20),
					common.BytesToAddress(founder.Bytes()),
				)

				// Check the erc20 balance, should be equal to the expected value
				require.Equal(t, big.NewInt(DefaultMinFeeValue/2), erc20Balance)
			},
		},
	}

	// Iterate and run the tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a cached context
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Start the mock bank keeper
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockBankKeeper := authtestutil.NewMockBankKeeper(ctrl)

			// If we have a expected value, we set on the mock bank keeper
			if !tc.expected.IsZero() {
				mockBankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr string, amt sdk.Coins) error {
						// Check if the amount is equal to the expected value
						require.Equal(t, tc.expected, amt)
						return nil
					},
				).AnyTimes()
			}

			// Start up the DeductFeeDecorator
			deductFeeDecorator := cosmos.NewDeductFeeDecorator(
				app.AccountKeeper,
				mockBankKeeper,
				nil, // Skip all the feegrant shenanigans
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			// Wrap into a ante decorator
			anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

			// Build a TX
			tx, err := helpers.BuildTxFromMsgs(
				founder,
				tc.fee,
				banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			require.NoError(t, err)

			// Call the ante handler
			_, err = anteHandler(cachedCtx, tx, false)
			if tc.errContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)

				if tc.postCheck != nil {
					tc.postCheck(cachedCtx)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
