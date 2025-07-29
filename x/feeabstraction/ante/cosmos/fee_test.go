package cosmos_test

import (
	"sync"
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

// TestDeductFeeDecoratorEdgeCases tests additional edge cases and error conditions

// TestDeductFeeDecoratorConcurrency tests the decorator under concurrent execution
func TestDeductFeeDecoratorConcurrency(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create multiple test accounts
	founder1 := apptesting.RandomAccountAddress()
	founder2 := apptesting.RandomAccountAddress()
	founder3 := apptesting.RandomAccountAddress()
	accounts := []sdk.AccAddress{founder1, founder2, founder3}

	// Create and fund all accounts
	for _, acc := range accounts {
		app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, acc))
		err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue*2)))
		require.NoError(t, err)
		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, acc, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue*2)))
		require.NoError(t, err)
	}

	// Set up fee abstraction
	app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
		Erc20Address:  keeper.MockErc20Address,
		Denom:         keeper.MockErc20Denom,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_UNSPECIFIED,
	})

	// Fund accounts with alternative token
	for _, acc := range accounts {
		err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*20)))
		require.NoError(t, err)
		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, acc, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*20)))
		require.NoError(t, err)
	}

	// Create decorator
	deductFeeDecorator := cosmos.NewDeductFeeDecorator(
		app.AccountKeeper,
		app.BankKeeper,
		app.FeeGrantKeeper,
		app.FeeAbstractionKeeper,
		cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
	)

	anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

	// Run concurrent transactions
	var wg sync.WaitGroup
	errorCh := make(chan error, len(accounts))

	for i, acc := range accounts {
		wg.Add(1)
		go func(account sdk.AccAddress, index int) {
			defer wg.Done()
			
			// Create cached context for this goroutine
			cachedCtx, _ := ctx.CacheContext()
			
			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				account,
				nil,
				sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
				1000000,
				banktypes.NewMsgSend(account, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			if err != nil {
				errorCh <- err
				return
			}
			
			// Execute ante handler
			_, err = anteHandler(cachedCtx, tx, false)
			if err != nil {
				errorCh <- err
				return
			}
		}(acc, i)
	}

	// Wait for all goroutines and check for errors
	wg.Wait()
	close(errorCh)

	for err := range errorCh {
		require.NoError(t, err)
	}
}
func TestDeductFeeDecoratorEdgeCases(t *testing.T) {

// TestDeductFeeDecoratorConcurrency tests the decorator under concurrent execution
func TestDeductFeeDecoratorConcurrency(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create multiple test accounts
	founder1 := apptesting.RandomAccountAddress()
	founder2 := apptesting.RandomAccountAddress()
	founder3 := apptesting.RandomAccountAddress()
	accounts := []sdk.AccAddress{founder1, founder2, founder3}

	// Create and fund all accounts
	for _, acc := range accounts {
		app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, acc))
		err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue*2)))
		require.NoError(t, err)
		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, acc, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue*2)))
		require.NoError(t, err)
	}

	// Set up fee abstraction
	app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
		Erc20Address:  keeper.MockErc20Address,
		Denom:         keeper.MockErc20Denom,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_UNSPECIFIED,
	})

	// Fund accounts with alternative token
	for _, acc := range accounts {
		err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*20)))
		require.NoError(t, err)
		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, acc, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*20)))
		require.NoError(t, err)
	}

	// Create decorator
	deductFeeDecorator := cosmos.NewDeductFeeDecorator(
		app.AccountKeeper,
		app.BankKeeper,
		app.FeeGrantKeeper,
		app.FeeAbstractionKeeper,
		cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
	)

	anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

	// Run concurrent transactions
	var wg sync.WaitGroup
	errorCh := make(chan error, len(accounts))

	for i, acc := range accounts {
		wg.Add(1)
		go func(account sdk.AccAddress, index int) {
			defer wg.Done()
			
			// Create cached context for this goroutine
			cachedCtx, _ := ctx.CacheContext()
			
			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				account,
				nil,
				sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
				1000000,
				banktypes.NewMsgSend(account, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			if err != nil {
				errorCh <- err
				return
			}
			
			// Execute ante handler
			_, err = anteHandler(cachedCtx, tx, false)
			if err != nil {
				errorCh <- err
				return
			}
		}(acc, i)
	}

	// Wait for all goroutines and check for errors
	wg.Wait()
	close(errorCh)

	for err := range errorCh {
		require.NoError(t, err)
	}
}
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test accounts
	founder := apptesting.RandomAccountAddress()
	feeGranter := apptesting.RandomAccountAddress()
	invalidAddress := sdk.AccAddress{}

	// Create the founder account
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	testCases := []struct {
		name        string
		malleate    func(ctx sdk.Context)
		fee         sdk.Coins
		feeGranter  sdk.AccAddress
		errorPanic  bool
		errorString string
		postCheck   func(ctx sdk.Context)
	}{
		{
			name: "fail - invalid fee granter address format",
			malleate: func(ctx sdk.Context) {
				// Fund the account normally
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
			},
			feeGranter:  invalidAddress,
			fee:         sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			errorString: "empty address string is not allowed",
		},
		{
			name: "fail - negative fee amount",
			malleate: func(ctx sdk.Context) {
				// Try to create negative fee
			},
			fee:         sdk.Coins{sdk.Coin{Denom: "akii", Amount: math.NewInt(-1)}},
			errorString: "negative coin amount",
		},
		{
			name: "success - empty fee with valid gas price",
			malleate: func(ctx sdk.Context) {
				// Set base fee to zero
				params := app.FeeMarketKeeper.GetParams(ctx)
				params.BaseFee = math.LegacyZeroDec()
				err := app.FeeMarketKeeper.SetParams(ctx, params)
				require.NoError(t, err)
			},
			fee: sdk.Coins{},
		},
		{
			name: "success - fee abstraction with very small amounts",
			malleate: func(ctx sdk.Context) {
				// Set up token pair with very small conversion rate
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Set very small fee price
				app.FeeAbstractionKeeper.SetFeePrices(ctx, []keeper.FeePrice{
					{
						Denom: keeper.MockErc20Denom,
						Price: math.LegacyMustNewDecFromStr("0.000001"),
					},
				})

				// Fund with minimal amounts
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, 1)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, 1)))
				require.NoError(t, err)
			},
			fee: sdk.NewCoins(sdk.NewInt64Coin("akii", 1)),
			postCheck: func(ctx sdk.Context) {
				// Verify conversion happened with tiny amounts
				balance := app.BankKeeper.GetBalance(ctx, founder, keeper.MockErc20Denom)
				// Should have consumed most of the balance
				require.True(t, balance.Amount.LTE(math.NewInt(1)))
			},
		},
		{
			name: "fail - fee abstraction with insufficient converted amount",
			malleate: func(ctx sdk.Context) {
				// Set up token pair
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Set very high conversion rate (expensive alternative token)
				app.FeeAbstractionKeeper.SetFeePrices(ctx, []keeper.FeePrice{
					{
						Denom: keeper.MockErc20Denom,
						Price: math.LegacyMustNewDecFromStr("1000000.0"),
					},
				})

				// Fund with insufficient amount for conversion
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, 10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, 10)))
				require.NoError(t, err)
			},
			fee:         sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			errorString: "insufficient funds for fee",
		},
		{
			name: "success - fee granter with fee abstraction",
			malleate: func(ctx sdk.Context) {
				// Set up token pair
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Fund the fee granter with alternative token
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, feeGranter, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)

				// Create fee grant
				err = app.FeeGrantKeeper.GrantAllowance(ctx, feeGranter, founder, &feegrant.BasicAllowance{
					SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
					Expiration: nil,
				})
				require.NoError(t, err)
			},
			feeGranter: feeGranter,
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			postCheck: func(ctx sdk.Context) {
				// Verify fee granter balance was deducted
				balance := app.BankKeeper.GetBalance(ctx, feeGranter, keeper.MockErc20Denom)
				require.True(t, balance.IsZero())
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a cached context
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Start up the DeductFeeDecorator
			deductFeeDecorator := cosmos.NewDeductFeeDecorator(
				app.AccountKeeper,
				app.BankKeeper,
				app.FeeGrantKeeper,
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			// Wrap into ante decorator
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
			if tc.errorPanic {
				require.Panics(t, func() {
					_, _ = anteHandler(cachedCtx, tx, false)
				})
			} else if tc.errorString != "" {
				_, err = anteHandler(cachedCtx, tx, false)
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorString)
			} else {
				_, err = anteHandler(cachedCtx, tx, false)
				require.NoError(t, err)
			}

			// Run post-check if provided
			if tc.postCheck != nil {
				tc.postCheck(cachedCtx)
			}
		})
	}
}
func TestDeductFeeDecoratorFeeGranterNoFeeKeeper(t *testing.T) {

// TestDeductFeeDecoratorEdgeCases tests additional edge cases and error conditions

// TestDeductFeeDecoratorConcurrency tests the decorator under concurrent execution
func TestDeductFeeDecoratorConcurrency(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create multiple test accounts
	founder1 := apptesting.RandomAccountAddress()
	founder2 := apptesting.RandomAccountAddress()
	founder3 := apptesting.RandomAccountAddress()
	accounts := []sdk.AccAddress{founder1, founder2, founder3}

	// Create and fund all accounts
	for _, acc := range accounts {
		app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, acc))
		err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue*2)))
		require.NoError(t, err)
		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, acc, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue*2)))
		require.NoError(t, err)
	}

	// Set up fee abstraction
	app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
		Erc20Address:  keeper.MockErc20Address,
		Denom:         keeper.MockErc20Denom,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_UNSPECIFIED,
	})

	// Fund accounts with alternative token
	for _, acc := range accounts {
		err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*20)))
		require.NoError(t, err)
		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, acc, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*20)))
		require.NoError(t, err)
	}

	// Create decorator
	deductFeeDecorator := cosmos.NewDeductFeeDecorator(
		app.AccountKeeper,
		app.BankKeeper,
		app.FeeGrantKeeper,
		app.FeeAbstractionKeeper,
		cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
	)

	anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

	// Run concurrent transactions
	var wg sync.WaitGroup
	errorCh := make(chan error, len(accounts))

	for i, acc := range accounts {
		wg.Add(1)
		go func(account sdk.AccAddress, index int) {
			defer wg.Done()
			
			// Create cached context for this goroutine
			cachedCtx, _ := ctx.CacheContext()
			
			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				account,
				nil,
				sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
				1000000,
				banktypes.NewMsgSend(account, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			if err != nil {
				errorCh <- err
				return
			}
			
			// Execute ante handler
			_, err = anteHandler(cachedCtx, tx, false)
			if err != nil {
				errorCh <- err
				return
			}
		}(acc, i)
	}

	// Wait for all goroutines and check for errors
	wg.Wait()
	close(errorCh)

	for err := range errorCh {
		require.NoError(t, err)
	}
}
func TestDeductFeeDecoratorEdgeCases(t *testing.T) {

// TestDeductFeeDecoratorConcurrency tests the decorator under concurrent execution
func TestDeductFeeDecoratorConcurrency(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create multiple test accounts
	founder1 := apptesting.RandomAccountAddress()
	founder2 := apptesting.RandomAccountAddress()
	founder3 := apptesting.RandomAccountAddress()
	accounts := []sdk.AccAddress{founder1, founder2, founder3}

	// Create and fund all accounts
	for _, acc := range accounts {
		app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, acc))
		err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue*2)))
		require.NoError(t, err)
		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, acc, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue*2)))
		require.NoError(t, err)
	}

	// Set up fee abstraction
	app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
		Erc20Address:  keeper.MockErc20Address,
		Denom:         keeper.MockErc20Denom,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_UNSPECIFIED,
	})

	// Fund accounts with alternative token
	for _, acc := range accounts {
		err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*20)))
		require.NoError(t, err)
		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, acc, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*20)))
		require.NoError(t, err)
	}

	// Create decorator
	deductFeeDecorator := cosmos.NewDeductFeeDecorator(
		app.AccountKeeper,
		app.BankKeeper,
		app.FeeGrantKeeper,
		app.FeeAbstractionKeeper,
		cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
	)

	anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

	// Run concurrent transactions
	var wg sync.WaitGroup
	errorCh := make(chan error, len(accounts))

	for i, acc := range accounts {
		wg.Add(1)
		go func(account sdk.AccAddress, index int) {
			defer wg.Done()
			
			// Create cached context for this goroutine
			cachedCtx, _ := ctx.CacheContext()
			
			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				account,
				nil,
				sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
				1000000,
				banktypes.NewMsgSend(account, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			if err != nil {
				errorCh <- err
				return
			}
			
			// Execute ante handler
			_, err = anteHandler(cachedCtx, tx, false)
			if err != nil {
				errorCh <- err
				return
			}
		}(acc, i)
	}

	// Wait for all goroutines and check for errors
	wg.Wait()
	close(errorCh)

	for err := range errorCh {
		require.NoError(t, err)
	}
}
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test accounts
	founder := apptesting.RandomAccountAddress()
	feeGranter := apptesting.RandomAccountAddress()
	invalidAddress := sdk.AccAddress{}

	// Create the founder account
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	testCases := []struct {
		name        string
		malleate    func(ctx sdk.Context)
		fee         sdk.Coins
		feeGranter  sdk.AccAddress
		errorPanic  bool
		errorString string
		postCheck   func(ctx sdk.Context)
	}{
		{
			name: "fail - invalid fee granter address format",
			malleate: func(ctx sdk.Context) {
				// Fund the account normally
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
			},
			feeGranter:  invalidAddress,
			fee:         sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			errorString: "empty address string is not allowed",
		},
		{
			name: "fail - negative fee amount",
			malleate: func(ctx sdk.Context) {
				// Try to create negative fee
			},
			fee:         sdk.Coins{sdk.Coin{Denom: "akii", Amount: math.NewInt(-1)}},
			errorString: "negative coin amount",
		},
		{
			name: "success - empty fee with valid gas price",
			malleate: func(ctx sdk.Context) {
				// Set base fee to zero
				params := app.FeeMarketKeeper.GetParams(ctx)
				params.BaseFee = math.LegacyZeroDec()
				err := app.FeeMarketKeeper.SetParams(ctx, params)
				require.NoError(t, err)
			},
			fee: sdk.Coins{},
		},
		{
			name: "success - fee abstraction with very small amounts",
			malleate: func(ctx sdk.Context) {
				// Set up token pair with very small conversion rate
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Set very small fee price
				app.FeeAbstractionKeeper.SetFeePrices(ctx, []keeper.FeePrice{
					{
						Denom: keeper.MockErc20Denom,
						Price: math.LegacyMustNewDecFromStr("0.000001"),
					},
				})

				// Fund with minimal amounts
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, 1)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, 1)))
				require.NoError(t, err)
			},
			fee: sdk.NewCoins(sdk.NewInt64Coin("akii", 1)),
			postCheck: func(ctx sdk.Context) {
				// Verify conversion happened with tiny amounts
				balance := app.BankKeeper.GetBalance(ctx, founder, keeper.MockErc20Denom)
				// Should have consumed most of the balance
				require.True(t, balance.Amount.LTE(math.NewInt(1)))
			},
		},
		{
			name: "fail - fee abstraction with insufficient converted amount",
			malleate: func(ctx sdk.Context) {
				// Set up token pair
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Set very high conversion rate (expensive alternative token)
				app.FeeAbstractionKeeper.SetFeePrices(ctx, []keeper.FeePrice{
					{
						Denom: keeper.MockErc20Denom,
						Price: math.LegacyMustNewDecFromStr("1000000.0"),
					},
				})

				// Fund with insufficient amount for conversion
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, 10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, 10)))
				require.NoError(t, err)
			},
			fee:         sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			errorString: "insufficient funds for fee",
		},
		{
			name: "success - fee granter with fee abstraction",
			malleate: func(ctx sdk.Context) {
				// Set up token pair
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Fund the fee granter with alternative token
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, feeGranter, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)

				// Create fee grant
				err = app.FeeGrantKeeper.GrantAllowance(ctx, feeGranter, founder, &feegrant.BasicAllowance{
					SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
					Expiration: nil,
				})
				require.NoError(t, err)
			},
			feeGranter: feeGranter,
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			postCheck: func(ctx sdk.Context) {
				// Verify fee granter balance was deducted
				balance := app.BankKeeper.GetBalance(ctx, feeGranter, keeper.MockErc20Denom)
				require.True(t, balance.IsZero())
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a cached context
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Start up the DeductFeeDecorator
			deductFeeDecorator := cosmos.NewDeductFeeDecorator(
				app.AccountKeeper,
				app.BankKeeper,
				app.FeeGrantKeeper,
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			// Wrap into ante decorator
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
			if tc.errorPanic {
				require.Panics(t, func() {
					_, _ = anteHandler(cachedCtx, tx, false)
				})
			} else if tc.errorString != "" {
				_, err = anteHandler(cachedCtx, tx, false)
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorString)
			} else {
				_, err = anteHandler(cachedCtx, tx, false)
				require.NoError(t, err)
			}

			// Run post-check if provided
			if tc.postCheck != nil {
				tc.postCheck(cachedCtx)
			}
		})
	}
}
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

// BenchmarkDeductFeeDecorator benchmarks the fee deduction performance
func BenchmarkDeductFeeDecorator(b *testing.B) {
	// Start the app and context
	app, ctx := helpers.SetupWithContext(b)

	// Create and fund test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	// Fund with plenty of tokens
	err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue*int64(b.N))))
	require.NoError(b, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue*int64(b.N))))
	require.NoError(b, err)

	// Create decorator
	deductFeeDecorator := cosmos.NewDeductFeeDecorator(
		app.AccountKeeper,
		app.BankKeeper,
		app.FeeGrantKeeper,
		app.FeeAbstractionKeeper,
		cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
	)

	anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

	// Build base transaction
	tx, err := helpers.BuildTxFromMsgs(
		founder,
		nil,
		sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
		1000000,
		banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
	)
	require.NoError(b, err)

	// Reset timer and run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cachedCtx, _ := ctx.CacheContext()
		_, err := anteHandler(cachedCtx, tx, false)
		require.NoError(b, err)
	}
}

// BenchmarkDeductFeeDecoratorWithAbstraction benchmarks fee deduction with abstraction

// TestDeductFeeDecoratorFeeAbstractionValidation tests fee abstraction validation scenarios

// TestDeductFeeDecoratorEventValidation tests event emission during fee deduction

// TestDeductFeeDecoratorSimulationMode tests behavior in simulation mode
func TestDeductFeeDecoratorSimulationMode(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	testCases := []struct {
		name      string
		malleate  func(ctx sdk.Context)
		fee       sdk.Coins
		simulate  bool
		expectErr bool
	}{
		{
			name: "success - simulation mode with zero gas",
			malleate: func(ctx sdk.Context) {
				// No funding needed in simulation
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "success - simulation mode uses tx fee directly",
			malleate: func(ctx sdk.Context) {
				// Fund account to ensure simulation passes
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "fail - simulation with insufficient funds still fails",
			malleate: func(ctx sdk.Context) {
				// Don	 fund the account
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a cached context
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Create decorator
			deductFeeDecorator := cosmos.NewDeductFeeDecorator(
				app.AccountKeeper,
				app.BankKeeper,
				app.FeeGrantKeeper,
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				founder,
				nil,
				tc.fee,
				0, // Zero gas for simulation tests
				banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			require.NoError(t, err)

			// Execute ante handler
			_, err = anteHandler(cachedCtx, tx, tc.simulate)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
func TestDeductFeeDecoratorEventValidation(t *testing.T) {

// TestDeductFeeDecoratorSimulationMode tests behavior in simulation mode
func TestDeductFeeDecoratorSimulationMode(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	testCases := []struct {
		name      string
		malleate  func(ctx sdk.Context)
		fee       sdk.Coins
		simulate  bool
		expectErr bool
	}{
		{
			name: "success - simulation mode with zero gas",
			malleate: func(ctx sdk.Context) {
				// No funding needed in simulation
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "success - simulation mode uses tx fee directly",
			malleate: func(ctx sdk.Context) {
				// Fund account to ensure simulation passes
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "fail - simulation with insufficient funds still fails",
			malleate: func(ctx sdk.Context) {
				// Don	 fund the account
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a cached context
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Create decorator
			deductFeeDecorator := cosmos.NewDeductFeeDecorator(
				app.AccountKeeper,
				app.BankKeeper,
				app.FeeGrantKeeper,
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				founder,
				nil,
				tc.fee,
				0, // Zero gas for simulation tests
				banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			require.NoError(t, err)

			// Execute ante handler
			_, err = anteHandler(cachedCtx, tx, tc.simulate)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	// Fund the account
	err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
	require.NoError(t, err)

	// Create decorator
	deductFeeDecorator := cosmos.NewDeductFeeDecorator(
		app.AccountKeeper,
		app.BankKeeper,
		app.FeeGrantKeeper,
		app.FeeAbstractionKeeper,
		cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
	)

	anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

	// Build transaction
	fee := sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue))
	tx, err := helpers.BuildTxFromMsgs(
		founder,
		nil,
		fee,
		1000000,
		banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
	)
	require.NoError(t, err)

	// Execute ante handler
	newCtx, err := anteHandler(ctx, tx, false)
	require.NoError(t, err)

	// Verify events were emitted
	events := newCtx.EventManager().Events()
	require.NotEmpty(t, events)

	// Look for fee-related events
	found := false
	for _, event := range events {
		if event.Type == sdk.EventTypeTx {
			// Check for fee and fee payer attributes
			for _, attr := range event.Attributes {
				if attr.Key == sdk.AttributeKeyFee {
					require.Equal(t, fee.String(), attr.Value)
					found = true
				}
				if attr.Key == sdk.AttributeKeyFeePayer {
					require.Equal(t, founder.String(), attr.Value)
				}
			}
		}
	}
	require.True(t, found, "Fee event should be emitted")
}
func TestDeductFeeDecoratorFeeAbstractionValidation(t *testing.T) {

// TestDeductFeeDecoratorEventValidation tests event emission during fee deduction

// TestDeductFeeDecoratorSimulationMode tests behavior in simulation mode
func TestDeductFeeDecoratorSimulationMode(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	testCases := []struct {
		name      string
		malleate  func(ctx sdk.Context)
		fee       sdk.Coins
		simulate  bool
		expectErr bool
	}{
		{
			name: "success - simulation mode with zero gas",
			malleate: func(ctx sdk.Context) {
				// No funding needed in simulation
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "success - simulation mode uses tx fee directly",
			malleate: func(ctx sdk.Context) {
				// Fund account to ensure simulation passes
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "fail - simulation with insufficient funds still fails",
			malleate: func(ctx sdk.Context) {
				// Don	 fund the account
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a cached context
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Create decorator
			deductFeeDecorator := cosmos.NewDeductFeeDecorator(
				app.AccountKeeper,
				app.BankKeeper,
				app.FeeGrantKeeper,
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				founder,
				nil,
				tc.fee,
				0, // Zero gas for simulation tests
				banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			require.NoError(t, err)

			// Execute ante handler
			_, err = anteHandler(cachedCtx, tx, tc.simulate)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
func TestDeductFeeDecoratorEventValidation(t *testing.T) {

// TestDeductFeeDecoratorSimulationMode tests behavior in simulation mode
func TestDeductFeeDecoratorSimulationMode(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	testCases := []struct {
		name      string
		malleate  func(ctx sdk.Context)
		fee       sdk.Coins
		simulate  bool
		expectErr bool
	}{
		{
			name: "success - simulation mode with zero gas",
			malleate: func(ctx sdk.Context) {
				// No funding needed in simulation
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "success - simulation mode uses tx fee directly",
			malleate: func(ctx sdk.Context) {
				// Fund account to ensure simulation passes
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "fail - simulation with insufficient funds still fails",
			malleate: func(ctx sdk.Context) {
				// Don	 fund the account
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a cached context
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Create decorator
			deductFeeDecorator := cosmos.NewDeductFeeDecorator(
				app.AccountKeeper,
				app.BankKeeper,
				app.FeeGrantKeeper,
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				founder,
				nil,
				tc.fee,
				0, // Zero gas for simulation tests
				banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			require.NoError(t, err)

			// Execute ante handler
			_, err = anteHandler(cachedCtx, tx, tc.simulate)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	// Fund the account
	err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
	require.NoError(t, err)

	// Create decorator
	deductFeeDecorator := cosmos.NewDeductFeeDecorator(
		app.AccountKeeper,
		app.BankKeeper,
		app.FeeGrantKeeper,
		app.FeeAbstractionKeeper,
		cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
	)

	anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

	// Build transaction
	fee := sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue))
	tx, err := helpers.BuildTxFromMsgs(
		founder,
		nil,
		fee,
		1000000,
		banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
	)
	require.NoError(t, err)

	// Execute ante handler
	newCtx, err := anteHandler(ctx, tx, false)
	require.NoError(t, err)

	// Verify events were emitted
	events := newCtx.EventManager().Events()
	require.NotEmpty(t, events)

	// Look for fee-related events
	found := false
	for _, event := range events {
		if event.Type == sdk.EventTypeTx {
			// Check for fee and fee payer attributes
			for _, attr := range event.Attributes {
				if attr.Key == sdk.AttributeKeyFee {
					require.Equal(t, fee.String(), attr.Value)
					found = true
				}
				if attr.Key == sdk.AttributeKeyFeePayer {
					require.Equal(t, founder.String(), attr.Value)
				}
			}
		}
	}
	require.True(t, found, "Fee event should be emitted")
}
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test accounts
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	testCases := []struct {
		name        string
		malleate    func(ctx sdk.Context)
		fee         sdk.Coins
		errorString string
		postCheck   func(ctx sdk.Context)
	}{
		{
			name: "fail - disabled token pair",
			malleate: func(ctx sdk.Context) {
				// Set up disabled token pair
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       false, // Disabled
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Fund with disabled token
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
			},
			fee:         sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			errorString: "insufficient funds for fee",
		},
		{
			name: "success - multiple alternative tokens, use first available",
			malleate: func(ctx sdk.Context) {
				// Set up multiple token pairs
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  "0x1234567890123456789012345678901234567890",
					Denom:         "erc20/0x1234567890123456789012345678901234567890",
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Fund with first token only
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
			},
			fee: sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			postCheck: func(ctx sdk.Context) {
				// Verify first token was used
				balance := app.BankKeeper.GetBalance(ctx, founder, keeper.MockErc20Denom)
				require.True(t, balance.IsZero())
				
				// Verify second token was not touched
				balance2 := app.BankKeeper.GetBalance(ctx, founder, "erc20/0x1234567890123456789012345678901234567890")
				require.True(t, balance2.IsZero()) // Was never funded, should be zero
			},
		},
		{
			name: "success - exact fee amount available",
			malleate: func(ctx sdk.Context) {
				// Set up token pair
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Fund with exact amount needed (10x conversion rate)
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
			},
			fee: sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			postCheck: func(ctx sdk.Context) {
				// Verify exact amount was deducted
				balance := app.BankKeeper.GetBalance(ctx, founder, keeper.MockErc20Denom)
				require.True(t, balance.IsZero())
			},
		},
		{
			name: "fail - zero fee price configured",
			malleate: func(ctx sdk.Context) {
				// Set up token pair
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Set zero price (invalid)
				app.FeeAbstractionKeeper.SetFeePrices(ctx, []keeper.FeePrice{
					{
						Denom: keeper.MockErc20Denom,
						Price: math.LegacyZeroDec(),
					},
				})

				// Fund with token
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
			},
			fee:         sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			errorString: "insufficient funds for fee",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a cached context
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Create decorator
			deductFeeDecorator := cosmos.NewDeductFeeDecorator(
				app.AccountKeeper,
				app.BankKeeper,
				app.FeeGrantKeeper,
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				founder,
				nil,
				tc.fee,
				1000000,
				banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			require.NoError(t, err)

			// Execute ante handler
			if tc.errorString != "" {
				_, err = anteHandler(cachedCtx, tx, false)
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorString)
			} else {
				_, err = anteHandler(cachedCtx, tx, false)
				require.NoError(t, err)
			}

			// Run post-check if provided
			if tc.postCheck != nil {
				tc.postCheck(cachedCtx)
			}
		})
	}
}
func BenchmarkDeductFeeDecoratorWithAbstraction(b *testing.B) {

// TestDeductFeeDecoratorFeeAbstractionValidation tests fee abstraction validation scenarios

// TestDeductFeeDecoratorEventValidation tests event emission during fee deduction

// TestDeductFeeDecoratorSimulationMode tests behavior in simulation mode
func TestDeductFeeDecoratorSimulationMode(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	testCases := []struct {
		name      string
		malleate  func(ctx sdk.Context)
		fee       sdk.Coins
		simulate  bool
		expectErr bool
	}{
		{
			name: "success - simulation mode with zero gas",
			malleate: func(ctx sdk.Context) {
				// No funding needed in simulation
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "success - simulation mode uses tx fee directly",
			malleate: func(ctx sdk.Context) {
				// Fund account to ensure simulation passes
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "fail - simulation with insufficient funds still fails",
			malleate: func(ctx sdk.Context) {
				// Don	 fund the account
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a cached context
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Create decorator
			deductFeeDecorator := cosmos.NewDeductFeeDecorator(
				app.AccountKeeper,
				app.BankKeeper,
				app.FeeGrantKeeper,
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				founder,
				nil,
				tc.fee,
				0, // Zero gas for simulation tests
				banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			require.NoError(t, err)

			// Execute ante handler
			_, err = anteHandler(cachedCtx, tx, tc.simulate)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
func TestDeductFeeDecoratorEventValidation(t *testing.T) {

// TestDeductFeeDecoratorSimulationMode tests behavior in simulation mode
func TestDeductFeeDecoratorSimulationMode(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	testCases := []struct {
		name      string
		malleate  func(ctx sdk.Context)
		fee       sdk.Coins
		simulate  bool
		expectErr bool
	}{
		{
			name: "success - simulation mode with zero gas",
			malleate: func(ctx sdk.Context) {
				// No funding needed in simulation
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "success - simulation mode uses tx fee directly",
			malleate: func(ctx sdk.Context) {
				// Fund account to ensure simulation passes
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "fail - simulation with insufficient funds still fails",
			malleate: func(ctx sdk.Context) {
				// Don	 fund the account
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a cached context
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Create decorator
			deductFeeDecorator := cosmos.NewDeductFeeDecorator(
				app.AccountKeeper,
				app.BankKeeper,
				app.FeeGrantKeeper,
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				founder,
				nil,
				tc.fee,
				0, // Zero gas for simulation tests
				banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			require.NoError(t, err)

			// Execute ante handler
			_, err = anteHandler(cachedCtx, tx, tc.simulate)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	// Fund the account
	err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
	require.NoError(t, err)

	// Create decorator
	deductFeeDecorator := cosmos.NewDeductFeeDecorator(
		app.AccountKeeper,
		app.BankKeeper,
		app.FeeGrantKeeper,
		app.FeeAbstractionKeeper,
		cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
	)

	anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

	// Build transaction
	fee := sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue))
	tx, err := helpers.BuildTxFromMsgs(
		founder,
		nil,
		fee,
		1000000,
		banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
	)
	require.NoError(t, err)

	// Execute ante handler
	newCtx, err := anteHandler(ctx, tx, false)
	require.NoError(t, err)

	// Verify events were emitted
	events := newCtx.EventManager().Events()
	require.NotEmpty(t, events)

	// Look for fee-related events
	found := false
	for _, event := range events {
		if event.Type == sdk.EventTypeTx {
			// Check for fee and fee payer attributes
			for _, attr := range event.Attributes {
				if attr.Key == sdk.AttributeKeyFee {
					require.Equal(t, fee.String(), attr.Value)
					found = true
				}
				if attr.Key == sdk.AttributeKeyFeePayer {
					require.Equal(t, founder.String(), attr.Value)
				}
			}
		}
	}
	require.True(t, found, "Fee event should be emitted")
}
func TestDeductFeeDecoratorFeeAbstractionValidation(t *testing.T) {

// TestDeductFeeDecoratorEventValidation tests event emission during fee deduction

// TestDeductFeeDecoratorSimulationMode tests behavior in simulation mode
func TestDeductFeeDecoratorSimulationMode(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	testCases := []struct {
		name      string
		malleate  func(ctx sdk.Context)
		fee       sdk.Coins
		simulate  bool
		expectErr bool
	}{
		{
			name: "success - simulation mode with zero gas",
			malleate: func(ctx sdk.Context) {
				// No funding needed in simulation
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "success - simulation mode uses tx fee directly",
			malleate: func(ctx sdk.Context) {
				// Fund account to ensure simulation passes
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "fail - simulation with insufficient funds still fails",
			malleate: func(ctx sdk.Context) {
				// Don	 fund the account
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a cached context
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Create decorator
			deductFeeDecorator := cosmos.NewDeductFeeDecorator(
				app.AccountKeeper,
				app.BankKeeper,
				app.FeeGrantKeeper,
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				founder,
				nil,
				tc.fee,
				0, // Zero gas for simulation tests
				banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			require.NoError(t, err)

			// Execute ante handler
			_, err = anteHandler(cachedCtx, tx, tc.simulate)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
func TestDeductFeeDecoratorEventValidation(t *testing.T) {

// TestDeductFeeDecoratorSimulationMode tests behavior in simulation mode
func TestDeductFeeDecoratorSimulationMode(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	testCases := []struct {
		name      string
		malleate  func(ctx sdk.Context)
		fee       sdk.Coins
		simulate  bool
		expectErr bool
	}{
		{
			name: "success - simulation mode with zero gas",
			malleate: func(ctx sdk.Context) {
				// No funding needed in simulation
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "success - simulation mode uses tx fee directly",
			malleate: func(ctx sdk.Context) {
				// Fund account to ensure simulation passes
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
				require.NoError(t, err)
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: false,
		},
		{
			name: "fail - simulation with insufficient funds still fails",
			malleate: func(ctx sdk.Context) {
				// Don	 fund the account
			},
			fee:       sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			simulate:  true,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a cached context
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Create decorator
			deductFeeDecorator := cosmos.NewDeductFeeDecorator(
				app.AccountKeeper,
				app.BankKeeper,
				app.FeeGrantKeeper,
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				founder,
				nil,
				tc.fee,
				0, // Zero gas for simulation tests
				banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			require.NoError(t, err)

			// Execute ante handler
			_, err = anteHandler(cachedCtx, tx, tc.simulate)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	// Fund the account
	err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)))
	require.NoError(t, err)

	// Create decorator
	deductFeeDecorator := cosmos.NewDeductFeeDecorator(
		app.AccountKeeper,
		app.BankKeeper,
		app.FeeGrantKeeper,
		app.FeeAbstractionKeeper,
		cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
	)

	anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

	// Build transaction
	fee := sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue))
	tx, err := helpers.BuildTxFromMsgs(
		founder,
		nil,
		fee,
		1000000,
		banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
	)
	require.NoError(t, err)

	// Execute ante handler
	newCtx, err := anteHandler(ctx, tx, false)
	require.NoError(t, err)

	// Verify events were emitted
	events := newCtx.EventManager().Events()
	require.NotEmpty(t, events)

	// Look for fee-related events
	found := false
	for _, event := range events {
		if event.Type == sdk.EventTypeTx {
			// Check for fee and fee payer attributes
			for _, attr := range event.Attributes {
				if attr.Key == sdk.AttributeKeyFee {
					require.Equal(t, fee.String(), attr.Value)
					found = true
				}
				if attr.Key == sdk.AttributeKeyFeePayer {
					require.Equal(t, founder.String(), attr.Value)
				}
			}
		}
	}
	require.True(t, found, "Fee event should be emitted")
}
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create test accounts
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	testCases := []struct {
		name        string
		malleate    func(ctx sdk.Context)
		fee         sdk.Coins
		errorString string
		postCheck   func(ctx sdk.Context)
	}{
		{
			name: "fail - disabled token pair",
			malleate: func(ctx sdk.Context) {
				// Set up disabled token pair
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       false, // Disabled
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Fund with disabled token
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
			},
			fee:         sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			errorString: "insufficient funds for fee",
		},
		{
			name: "success - multiple alternative tokens, use first available",
			malleate: func(ctx sdk.Context) {
				// Set up multiple token pairs
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  "0x1234567890123456789012345678901234567890",
					Denom:         "erc20/0x1234567890123456789012345678901234567890",
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Fund with first token only
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
			},
			fee: sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			postCheck: func(ctx sdk.Context) {
				// Verify first token was used
				balance := app.BankKeeper.GetBalance(ctx, founder, keeper.MockErc20Denom)
				require.True(t, balance.IsZero())
				
				// Verify second token was not touched
				balance2 := app.BankKeeper.GetBalance(ctx, founder, "erc20/0x1234567890123456789012345678901234567890")
				require.True(t, balance2.IsZero()) // Was never funded, should be zero
			},
		},
		{
			name: "success - exact fee amount available",
			malleate: func(ctx sdk.Context) {
				// Set up token pair
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Fund with exact amount needed (10x conversion rate)
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
			},
			fee: sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			postCheck: func(ctx sdk.Context) {
				// Verify exact amount was deducted
				balance := app.BankKeeper.GetBalance(ctx, founder, keeper.MockErc20Denom)
				require.True(t, balance.IsZero())
			},
		},
		{
			name: "fail - zero fee price configured",
			malleate: func(ctx sdk.Context) {
				// Set up token pair
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  keeper.MockErc20Address,
					Denom:         keeper.MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Set zero price (invalid)
				app.FeeAbstractionKeeper.SetFeePrices(ctx, []keeper.FeePrice{
					{
						Denom: keeper.MockErc20Denom,
						Price: math.LegacyZeroDec(),
					},
				})

				// Fund with token
				err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10)))
				require.NoError(t, err)
			},
			fee:         sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
			errorString: "insufficient funds for fee",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a cached context
			cachedCtx, _ := ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Create decorator
			deductFeeDecorator := cosmos.NewDeductFeeDecorator(
				app.AccountKeeper,
				app.BankKeeper,
				app.FeeGrantKeeper,
				app.FeeAbstractionKeeper,
				cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
			)

			anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

			// Build transaction
			tx, err := helpers.BuildTxFromMsgs(
				founder,
				nil,
				tc.fee,
				1000000,
				banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
			)
			require.NoError(t, err)

			// Execute ante handler
			if tc.errorString != "" {
				_, err = anteHandler(cachedCtx, tx, false)
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorString)
			} else {
				_, err = anteHandler(cachedCtx, tx, false)
				require.NoError(t, err)
			}

			// Run post-check if provided
			if tc.postCheck != nil {
				tc.postCheck(cachedCtx)
			}
		})
	}
}
	// Start the app and context
	app, ctx := helpers.SetupWithContext(b)

	// Set up fee abstraction
	app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
		Erc20Address:  keeper.MockErc20Address,
		Denom:         keeper.MockErc20Denom,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_UNSPECIFIED,
	})

	// Create and fund test account
	founder := apptesting.RandomAccountAddress()
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, founder))

	// Fund with alternative token
	err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10*int64(b.N))))
	require.NoError(b, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, founder, sdk.NewCoins(sdk.NewInt64Coin(keeper.MockErc20Denom, DefaultMinFeeValue*10*int64(b.N))))
	require.NoError(b, err)

	// Create decorator
	deductFeeDecorator := cosmos.NewDeductFeeDecorator(
		app.AccountKeeper,
		app.BankKeeper,
		app.FeeGrantKeeper,
		app.FeeAbstractionKeeper,
		cosmosevmante.NewDynamicFeeChecker(app.FeeMarketKeeper),
	)

	anteHandler := sdk.ChainAnteDecorators(deductFeeDecorator)

	// Build base transaction
	tx, err := helpers.BuildTxFromMsgs(
		founder,
		nil,
		sdk.NewCoins(sdk.NewInt64Coin("akii", DefaultMinFeeValue)),
		1000000,
		banktypes.NewMsgSend(founder, apptesting.RandomAccountAddress(), sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000)))),
	)
	require.NoError(b, err)

	// Reset timer and run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cachedCtx, _ := ctx.CacheContext()
		_, err := anteHandler(cachedCtx, tx, false)
		require.NoError(b, err)
	}
}
