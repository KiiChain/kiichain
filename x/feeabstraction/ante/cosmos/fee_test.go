package cosmos_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	"cosmossdk.io/x/feegrant"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	cosmosevmante "github.com/cosmos/evm/ante/evm"
	"github.com/cosmos/evm/contracts"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/kiichain/kiichain/v4/app/apptesting"
	"github.com/kiichain/kiichain/v4/app/helpers"
	"github.com/kiichain/kiichain/v4/x/feeabstraction/ante/cosmos"
	"github.com/kiichain/kiichain/v4/x/feeabstraction/types"
)

var (
	DefaultFirstERC20      = "0x80b5a32E4F032B2a058b4F29EC95EEfEEB87aDcd"
	DefaultFirstERC20Denom = "erc20/" + DefaultFirstERC20
	DefaultMinFeeValue     = int64(875000000000000)

	MockErc20Address = "0x816644F8bc4633D268842628EB10ffC0AdcB6099"
	// The mock ERC20 denom
	MockErc20Denom = "erc20/" + MockErc20Address
	// The mock ERC20 price
	MockErc20Price = math.LegacyNewDecFromInt(math.NewInt(10)) // 10 uatom = 1 kii
)

// TestDeductFeeDecorator tests the DeductFeeDecorator
// This function tests the fee conversion and deduction logic
func TestDeductFeeDecorator(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create a fee payer
	founder := apptesting.RandomAccountAddress()
	feeGranter := apptesting.RandomAccountAddress()

	// Create the funder account
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	// Set the different test cases
	testCases := []struct {
		name        string
		malleate    func(ctx sdk.Context)
		fee         sdk.Coins
		expected    sdk.Coins
		feeGranter  sdk.AccAddress
		errContains string
		postCheck   func(ctx sdk.Context)
	}{
		{
			name: "success - valid fee deduction",
			malleate: func(ctx sdk.Context) {
				// Fund the account with enough funds to pay the fee
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
			},
			fee:      sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			expected: sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
		},
		{
			name: "success - valid fee deduction with fee granter",
			malleate: func(ctx sdk.Context) {
				// Fund the fee granter account with enough funds to pay the fee
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, feeGranter, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)

				// Create the fee grant
				err = app.FeeGrantKeeper.GrantAllowance(ctx, feeGranter, founder, &feegrant.BasicAllowance{
					SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
					Expiration: nil,
				})
				require.NoError(t, err)
			},
			feeGranter: feeGranter,
			fee:        sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			expected:   sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
		},
		{
			name: "success - valid fee deduction with multiple coins",
			malleate: func(ctx sdk.Context) {
				// Fund the account with enough funds to pay the fee
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
			},
			fee: sdk.NewCoins(
				sdk.NewInt64Coin("akii", DefaultMinFeeValue),
				sdk.NewInt64Coin("other", DefaultMinFeeValue),
			),
			// Even with multiple coins, only akii is used for fees
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
			name:        "fail - failure on tx fee checker",
			fee:         sdk.NewCoins(sdk.NewInt64Coin("akii", 1)),
			expected:    sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			errContains: " Please retry using a higher gas price or a higher fee",
		},
		{
			name: "fail - nonexistent fee payer",
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
					Erc20Address:  MockErc20Address,
					Denom:         MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Set the pair on the fee abstraction keeper
				err := app.FeeAbstractionKeeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					types.NewFeeTokenMetadata(
						MockErc20Denom,
						MockErc20Denom,
						18,
						MockErc20Price,
					),
				))
				require.NoError(t, err)

				// Now we mint tokens for the fee payer
				err = app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
			},
			fee:      sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			expected: sdk.NewCoins(sdk.NewInt64Coin(MockErc20Denom, DefaultMinFeeValue*10)),
			postCheck: func(ctx sdk.Context) {
				// Check the user balance, should be zero since all was user for fees
				balance := app.BankKeeper.GetBalance(ctx, founder, MockErc20Denom)
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
				err = app.FeeAbstractionKeeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					types.NewFeeTokenMetadata(
						erc20NativeAddress,
						erc20NativeAddress,
						18,
						math.LegacyMustNewDecFromStr("0.5"),
					),
				))
				require.NoError(t, err)
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
		{
			name:        "fail - unauthorized fee grant",
			feeGranter:  feeGranter,
			fee:         sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			expected:    sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			errContains: "fee-grant not found",
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
						// Check the fromAddr is the fee payer
						if tc.feeGranter != nil {
							require.Equal(t, tc.feeGranter, fromAddr)
						} else {
							require.Equal(t, founder, fromAddr)
						}

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
				app.FeeGrantKeeper,
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			// Wrap into a ante decorator
			anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

			// Build a TX
			tx, err := helpers.BuildTxFromMsgs(
				founder,
				tc.feeGranter,
				tc.fee,
				1000000,
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

// TestDeductFeeDecoratorCheckerNil tests the DeductFeeDecorator with a nil checker
func TestDeductFeeDecoratorCheckerNil(t *testing.T) {
	// Start the app and the context
	app, _ := helpers.SetupWithContext(t)

	// Start up the DeductFeeDecorator with a nil checker
	require.PanicsWithValue(t, "txFeeChecker cannot be nil", func() {
		cosmos.NewDeductFeeDecorator(
			app.AccountKeeper,
			app.BankKeeper,
			nil, // Skip all the feegrant shenanigans
			app.FeeAbstractionKeeper,
			nil, // Set checker to nil
		)
	})
}

// TestDeductFeeDecoratorGasZero tests the DeductFeeDecorator with a zero gas limit
func TestDeductFeeDecoratorGasZero(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Start up the DeductFeeDecorator with a nil checker
	deductFeeDecorator := cosmos.NewDeductFeeDecorator(
		app.AccountKeeper,
		app.BankKeeper,
		nil, // Skip all the feegrant shenanigans
		app.FeeAbstractionKeeper,
		cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
	)

	// Create a fee payer
	founder := apptesting.RandomAccountAddress()

	// Wrap into a ante decorator
	anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

	// Build a TX
	tx, err := helpers.BuildTxFromMsgs(
		founder,
		nil,
		sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
		0, // Set gas limit to zero
		banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
	)
	require.NoError(t, err)

	// Run the ante handler
	_, err = anteHandler(ctx, tx, false)
	require.ErrorContains(t, err, "must provide positive gas")
}

// TestDeductFeeDecoratorFeeGranterNoFeeKeeper tests the DeductFeeDecorator with a nil fee keeper
func TestDeductFeeDecoratorFeeGranterNoFeeKeeper(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Start up the DeductFeeDecorator with a nil checker
	deductFeeDecorator := cosmos.NewDeductFeeDecorator(
		app.AccountKeeper,
		app.BankKeeper,
		nil,
		app.FeeAbstractionKeeper,
		cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
	)

	// Create a fee payer
	founder := apptesting.RandomAccountAddress()

	// Wrap into a ante decorator
	anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

	// Build a TX
	tx, err := helpers.BuildTxFromMsgs(
		founder,
		apptesting.RandomAccountAddress(), // Set fee granter to a random address
		sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
		1000000,
		banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
	)
	require.NoError(t, err)

	// Run the ante handler
	_, err = anteHandler(ctx, tx, false)
	require.ErrorContains(t, err, "fee grants are not enabled")
}
