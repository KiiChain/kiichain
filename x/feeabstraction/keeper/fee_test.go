package keeper_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/cosmos/evm/contracts"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/kiichain/kiichain/v4/app/apptesting"
	"github.com/kiichain/kiichain/v4/x/feeabstraction/types"
)

// TestConvertNativeFee tests the ConvertNativeFee function
func (s *KeeperTestSuite) TestConvertNativeFee() {
	// Fee payer
	feePayer := apptesting.RandomAccountAddress()
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, feePayer))

	// Default erc20 address to use in tests
	DefaultFirstERC20 := "0x80b5a32E4F032B2a058b4F29EC95EEfEEB87aDcd"

	// Build the test cases
	testCases := []struct {
		name        string
		malleate    func(sdk.Context) sdk.Context
		fees        sdk.Coins
		expected    sdk.Coins
		postCheck   func(sdk.Context, sdk.Coins)
		errContains string
	}{
		{
			name: "success - nothing happens, module disabled",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Disable the module
				params, err := s.keeper.Params.Get(ctx)
				s.Require().NoError(err)
				params.Enabled = false
				s.Require().NoError(s.keeper.Params.Set(ctx, params))
				return ctx
			},
			fees:     sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000))),
			expected: sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000))),
		},
		{
			name:     "success - no fee tokens, no conversion",
			fees:     sdk.NewCoins(),
			expected: sdk.NewCoins(),
		},
		{
			name: "success - multiple fee tokens, no conversion",
			fees: sdk.NewCoins(
				sdk.NewCoin("uatom", math.NewInt(1000)),
				sdk.NewCoin("akii", math.NewInt(500)),
			),
			expected: sdk.NewCoins(
				sdk.NewCoin("uatom", math.NewInt(1000)),
				sdk.NewCoin("akii", math.NewInt(500)),
			),
		},
		{
			name:     "success - fee is not based on the native denom, no conversion",
			fees:     sdk.NewCoins(sdk.NewCoin("uatom", math.NewInt(1000))),
			expected: sdk.NewCoins(sdk.NewCoin("uatom", math.NewInt(1000))),
		},
		{
			name: "success - user has sufficient native balance, no conversion",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Fund the user with sufficient native balance
				s.fundAccount(ctx, feePayer, sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000))))
				return ctx
			},
			fees:     sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000))),
			expected: sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000))),
		},
		{
			name: "fail - user has insufficient native balance, all tokens are disabled",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Register a fee token but do not fund the user with it
				err := s.keeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					types.FeeTokenMetadata{
						Denom:       "coin",
						OracleDenom: "oraclecoin",
						Decimals:    18,
						Price:       math.LegacyOneDec(),
						Enabled:     false,
					},
				))
				s.Require().NoError(err)

				return ctx
			},
			fees:        sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1000))),
			errContains: "insufficient funds for fee",
		},
		{
			name: "success - user has sufficient balance, no erc20 unwrap needed",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Register a fee token and fund the user with it
				err := s.keeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					// 0.1 atom per kii
					types.NewFeeTokenMetadata("uatom", "atomoracle", 6, math.LegacyOneDec()),
				))
				s.Require().NoError(err)

				// Fund the user with sufficient native balance
				s.fundAccount(ctx, feePayer, sdk.NewCoins(sdk.NewCoin("uatom", convertToMinimalDenomination(1, 18))))
				return ctx
			},
			fees: sdk.NewCoins(sdk.NewCoin("akii", convertToMinimalDenomination(1, 18))), // 1 Kii
			// Since 1 Kii is 1 Atom, and we are asking for 1 kii fee, we expect 1 Atom in return
			expected: sdk.NewCoins(sdk.NewCoin("uatom", convertToMinimalDenomination(1, 6))), // 1 Atom
		},
		{
			name: "success - user has insufficient balance, multiple fee tokens, pay with the middle token",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Register multiple fee tokens
				err := s.keeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					types.NewFeeTokenMetadata("uatom", "atomoracle", 6, math.LegacyMustNewDecFromStr("0.123")),
					types.NewFeeTokenMetadata("usol", "usoloracle", 9, math.LegacyMustNewDecFromStr("0.125")),
					types.NewFeeTokenMetadata("mbtc", "btcoracle", 8, math.LegacyMustNewDecFromStr("2")),
				))
				s.Require().NoError(err)

				// Fund the user with insufficient native balance
				s.fundAccount(ctx, feePayer, sdk.NewCoins(sdk.NewCoin("usol", convertToMinimalDenomination(1, 18))))
				return ctx
			},
			fees:     sdk.NewCoins(sdk.NewCoin("akii", convertToMinimalDenomination(1, 17))),  // 0.1 Kii
			expected: sdk.NewCoins(sdk.NewCoin("usol", convertToMinimalDenomination(125, 5))), // 0.125 USOL
		},
		{
			name: "success - user has insufficient balance, multiple fee tokens, pay with the first available token",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Register multiple fee tokens
				err := s.keeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					types.NewFeeTokenMetadata("uatom", "atomoracle", 6, math.LegacyMustNewDecFromStr("0.123")),
					types.NewFeeTokenMetadata("usol", "usoloracle", 9, math.LegacyMustNewDecFromStr("0.125")),
					types.NewFeeTokenMetadata("mbtc", "btcoracle", 8, math.LegacyMustNewDecFromStr("2")),
				))
				s.Require().NoError(err)

				// Fund the user with insufficient native balance
				s.fundAccount(ctx, feePayer, sdk.NewCoins(sdk.NewCoin("usol", convertToMinimalDenomination(1, 18))))
				s.fundAccount(ctx, feePayer, sdk.NewCoins(sdk.NewCoin("mbtc", convertToMinimalDenomination(1, 18))))
				return ctx
			},
			fees:     sdk.NewCoins(sdk.NewCoin("akii", convertToMinimalDenomination(1, 17))),  // 0.1 Kii
			expected: sdk.NewCoins(sdk.NewCoin("usol", convertToMinimalDenomination(125, 5))), // 0.00125 usol
		},
		{
			name: "fail - token with price as zero",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Register a fee token with zero price
				err := s.keeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					types.NewFeeTokenMetadata("uatom", "atomoracle", 6, math.LegacyZeroDec()),
				))
				s.Require().NoError(err)

				// Fund the user with more than enough tokens to pay the fee
				s.fundAccount(ctx, feePayer, sdk.NewCoins(sdk.NewCoin("uatom", convertToMinimalDenomination(1, 18))))
				return ctx
			},
			fees:        sdk.NewCoins(sdk.NewCoin("akii", convertToMinimalDenomination(1, 18))), // 1 Kii
			errContains: "insufficient funds for fee",
		},
		{
			name: "success - pay using erc20 token (exact amount)",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Deploy the erc20 token
				erc20Address, err := apptesting.DeployERC20(ctx, s.app)
				s.Require().NoError(err)

				// Mint to the fee payer
				feeAmount := big.NewInt(20000)
				err = apptesting.MintERC20(ctx, s.app, erc20Address, common.BytesToAddress(feePayer.Bytes()), feeAmount)
				s.Require().NoError(err)

				// Set the token pair on the erc20 keeper
				_, err = s.app.Erc20Keeper.RegisterERC20(ctx, &erc20types.MsgRegisterERC20{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Erc20Addresses: []string{
						erc20Address.Hex(),
					},
				})
				s.Require().NoError(err)

				// Register the fee token
				erc20NativeAddress := "erc20/" + erc20Address.Hex()
				err = s.keeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					types.NewFeeTokenMetadata(
						erc20NativeAddress,
						"oracleerc20",
						6,
						math.LegacyMustNewDecFromStr("1"),
					),
				))
				s.Require().NoError(err)

				return ctx
			},
			fees:     sdk.NewCoins(sdk.NewCoin("akii", convertToMinimalDenomination(2, 16))),    // 0.02 Kii
			expected: sdk.NewCoins(sdk.NewCoin("erc20/"+DefaultFirstERC20, math.NewInt(20000))), // 20000 of the erc20 token
			postCheck: func(ctx sdk.Context, convertedFees sdk.Coins) {
				// The user now should have the balance as native token available for fees
				balance := s.app.BankKeeper.GetBalance(ctx, feePayer, "erc20/"+DefaultFirstERC20)
				s.Require().Equal(math.NewInt(20000), balance.Amount)

				// The contract should have zero balance
				// Get the erc20 balance
				erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
				erc20Balance := s.app.Erc20Keeper.BalanceOf(
					ctx,
					erc20,
					common.HexToAddress(DefaultFirstERC20),
					common.BytesToAddress(feePayer.Bytes()),
				)

				// The balance should be zero
				s.Require().EqualValues(big.NewInt(0).Int64(), erc20Balance.Int64())
			},
		},
		{
			name: "fail - pay using erc20 token (insufficient funds)",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Deploy the erc20 token
				erc20Address, err := apptesting.DeployERC20(ctx, s.app)
				s.Require().NoError(err)

				// Mint to the fee payer
				feeAmount := big.NewInt(10000) // Only 10000
				err = apptesting.MintERC20(ctx, s.app, erc20Address, common.BytesToAddress(feePayer.Bytes()), feeAmount)
				s.Require().NoError(err)

				// Set the token pair on the erc20 keeper
				_, err = s.app.Erc20Keeper.RegisterERC20(ctx, &erc20types.MsgRegisterERC20{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Erc20Addresses: []string{
						erc20Address.Hex(),
					},
				})
				s.Require().NoError(err)

				// Register the fee token
				erc20NativeAddress := "erc20/" + erc20Address.Hex()
				err = s.keeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					types.NewFeeTokenMetadata(
						erc20NativeAddress,
						"oracleerc20",
						6,
						math.LegacyMustNewDecFromStr("1"),
					),
				))
				s.Require().NoError(err)

				return ctx
			},
			fees:        sdk.NewCoins(sdk.NewCoin("akii", convertToMinimalDenomination(2, 16))), // 0.02 Kii
			errContains: "insufficient funds for fee",
		},
		{
			name: "success - more than needed erc20 balance",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Deploy the erc20 token
				erc20Address, err := apptesting.DeployERC20(ctx, s.app)
				s.Require().NoError(err)

				// Mint to the fee payer
				feeAmount := big.NewInt(50000) // 50000
				err = apptesting.MintERC20(ctx, s.app, erc20Address, common.BytesToAddress(feePayer.Bytes()), feeAmount)
				s.Require().NoError(err)

				// Set the token pair on the erc20 keeper
				_, err = s.app.Erc20Keeper.RegisterERC20(ctx, &erc20types.MsgRegisterERC20{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Erc20Addresses: []string{
						erc20Address.Hex(),
					},
				})
				s.Require().NoError(err)

				// Register the fee token
				erc20NativeAddress := "erc20/" + erc20Address.Hex()
				err = s.keeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					types.NewFeeTokenMetadata(
						erc20NativeAddress,
						"oracleerc20",
						6,
						math.LegacyMustNewDecFromStr("1"),
					),
				))
				s.Require().NoError(err)

				return ctx
			},
			fees:     sdk.NewCoins(sdk.NewCoin("akii", convertToMinimalDenomination(2, 16))),    // 0.02 Kii
			expected: sdk.NewCoins(sdk.NewCoin("erc20/"+DefaultFirstERC20, math.NewInt(20000))), // 20000 of the erc20 token
			postCheck: func(ctx sdk.Context, convertedFees sdk.Coins) {
				// The user now should have the balance as native token available for fees
				balance := s.app.BankKeeper.GetBalance(ctx, feePayer, "erc20/"+DefaultFirstERC20)
				s.Require().Equal(math.NewInt(20000), balance.Amount)

				// The contract should have some balance left
				erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
				erc20Balance := s.app.Erc20Keeper.BalanceOf(
					ctx,
					erc20,
					common.HexToAddress(DefaultFirstERC20),
					common.BytesToAddress(feePayer.Bytes()),
				)

				// The balance should be 30000 (50000 minted - 20000 converted)
				s.Require().EqualValues(30000, erc20Balance.Int64())
			},
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Create a cached context
			cachedCtx, _ := s.ctx.CacheContext()

			// Malleate the system
			if tc.malleate != nil {
				cachedCtx = tc.malleate(cachedCtx)
			}

			// Call the ConvertNativeFee function
			convertedFees, err := s.keeper.ConvertNativeFee(cachedCtx, feePayer, tc.fees)

			// Check for expected error
			if tc.errContains != "" {
				// Error should be contain
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)

				// Check if the fee match
				s.Require().Equal(tc.expected, convertedFees)
			}

			// Run any post-checks if provided
			if tc.postCheck != nil {
				tc.postCheck(cachedCtx, convertedFees)
			}
		})
	}
}

// convertToMinimalDenomination converts a int to a base denom given a decimals
func convertToMinimalDenomination(amount int, decimals int) math.Int {
	// Convert it to LegacyDec
	dec := math.LegacyNewDec(int64(amount)).Mul(math.LegacyNewDec(10).Power(uint64(decimals)))

	// Return truncated
	return dec.RoundInt()
}

// fundAccount is a helper function to fund an account with a specific amount of coins
func (s *KeeperTestSuite) fundAccount(ctx sdk.Context, account sdk.AccAddress, amount sdk.Coins) {
	// Fund the account with some coins
	err := s.app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, amount)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, account, amount)
	s.Require().NoError(err)
}
