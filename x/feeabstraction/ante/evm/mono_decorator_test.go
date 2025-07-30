package evm_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/cosmos/evm/contracts"
	"github.com/cosmos/evm/testutil/integration/os/keyring"
	"github.com/cosmos/evm/testutil/tx"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	"github.com/cosmos/evm/x/vm/statedb"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	kiichain "github.com/kiichain/kiichain/v3/app"
	"github.com/kiichain/kiichain/v3/app/apptesting"
	"github.com/kiichain/kiichain/v3/app/helpers"
	"github.com/kiichain/kiichain/v3/app/params"
	kiievmante "github.com/kiichain/kiichain/v3/x/feeabstraction/ante/evm"
	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

var (
	MockErc20Address = "0x816644F8bc4633D268842628EB10ffC0AdcB6099"
	// The mock ERC20 denom
	MockErc20Denom = "erc20/" + MockErc20Address
	// The mock ERC20 price
	MockErc20Price = math.LegacyNewDecFromInt(math.NewInt(10)) // 10 uatom = 1 kii
)

// TestMonoDecoratorTx tests the MonoDecorator with specific transaction cases
// This focus on bad cases, while the good cases are covered by the TestMonoDecorator
func TestMonoDecoratorTx(t *testing.T) {
	// Create the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Define the test cases
	testCases := []struct {
		name        string
		tx          signing.Tx
		errContains string
	}{
		{
			name: "tx validate fail",
			tx: func() signing.Tx {
				// Start the tx builder
				encodingConfig := params.MakeEncodingConfig()
				txBuilder := encodingConfig.TxConfig.NewTxBuilder()

				return txBuilder.GetTx()
			}(),
			errContains: "eth tx length of ExtensionOptions should be 1",
		},
		{
			name: "tx fail - invalid signature",
			tx: func() signing.Tx {
				// Create a transaction with an invalid amount
				ethChainID := big.NewInt(1010)
				msgEthereumTx := evmtypes.NewTx(getDefaultEVMTxArgs(ethChainID, 20000000, big.NewInt(1000000), -10))

				// Start the tx builder
				encodingConfig := params.MakeEncodingConfig()
				txBuilder := encodingConfig.TxConfig.NewTxBuilder()

				// Create the TX
				tx, err := msgEthereumTx.BuildTx(txBuilder, "akii")
				require.NoError(t, err)

				return tx
			}(),
			errContains: "tx intended signer does not match the given signer",
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a cached context
			cacheCtx, _ := ctx.CacheContext()

			// Start up the wrapped ante decorator
			monoDecorator := kiievmante.NewEVMMonoDecorator(
				app.AccountKeeper,
				app.FeeMarketKeeper,
				app.EVMKeeper,
				app.FeeAbstractionKeeper,
				20000000,
			)
			anteHandler := sdk.ChainAnteDecorators(monoDecorator)

			// Execute the ante handler
			_, err := anteHandler(cacheCtx, tc.tx, false)
			if tc.errContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestMonoDecorator tests the MonoDecorator and its changes related to the fee abstraction
// This test focus on the fee abstraction changes and slight cover the mono decorator implementation
// Full coverage is ensured by the evm module
func TestMonoDecorator(t *testing.T) {
	// Start the app and the context
	app, ctx := helpers.SetupWithContext(t)

	// Create a keyring and separate a single key
	keys := keyring.New(1)

	// Set the fee market fees to a good value for calculations
	feeMarketParams := app.FeeMarketKeeper.GetParams(ctx)
	feeMarketParams.MinGasPrice = math.LegacyMustNewDecFromStr("1000000")
	feeMarketParams.BaseFee = math.LegacyMustNewDecFromStr("1000000")
	err := app.FeeMarketKeeper.SetParams(ctx, feeMarketParams)
	require.NoError(t, err)

	// Define the test cases
	testCases := []struct {
		name        string
		malleate    func(ctx sdk.Context) sdk.Context
		gasLimit    uint64
		gasPrice    *big.Int
		txAmount    int64
		errContains string
		postCheck   func(ctx sdk.Context)
	}{
		{
			name: "success - no fee charged - account is created",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Get the fee market params
				feeMarketParams := app.FeeMarketKeeper.GetParams(ctx)
				// Set the fee market params
				feeMarketParams.MinGasPrice = math.LegacyZeroDec()
				feeMarketParams.BaseFee = math.LegacyZeroDec()
				err = app.FeeMarketKeeper.SetParams(ctx, feeMarketParams)
				require.NoError(t, err)

				// Make sure that the account doesnt exist yet
				acc := app.AccountKeeper.GetAccount(ctx, keys.GetKey(0).AccAddr)
				require.Nil(t, acc)

				return ctx
			},
			gasLimit: 20000000,
			gasPrice: big.NewInt(0),
			postCheck: func(ctx sdk.Context) {
				// Check the account balance after the transaction, all should be consumed
				balance := app.BankKeeper.GetBalance(ctx, keys.GetKey(0).AccAddr, "akii")
				require.Equal(t, math.NewInt(0), balance.Amount)

				// Check that the account was created
				acc := app.AccountKeeper.GetAccount(ctx, keys.GetKey(0).AccAddr)
				require.NotNil(t, acc)
			},
		},
		{
			name: "success - fee charged",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Fund the account with token to pay for the fees
				amount := sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(20000000*1000000)))
				err := mintCoins(app, ctx, keys.GetKey(0).AccAddr, amount)
				require.NoError(t, err)
				return ctx
			},
			gasLimit: 20000000,
			gasPrice: big.NewInt(1000000),
			postCheck: func(ctx sdk.Context) {
				// Check the account balance after the transaction, all should be consumed
				balance := app.BankKeeper.GetBalance(ctx, keys.GetKey(0).AccAddr, "akii")
				require.Equal(t, math.NewInt(0), balance.Amount)
			},
		},
		{
			name: "success - fee charged with fee abstraction native token",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Set up the token pair on the erc20 module
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  MockErc20Address,
					Denom:         MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Mint the tokens for the fee payer
				amount := sdk.NewCoins(sdk.NewInt64Coin(MockErc20Denom, 20000000*1000000*10))
				err := mintCoins(app, ctx, keys.GetKey(0).AccAddr, amount)
				require.NoError(t, err)
				return ctx
			},
			gasLimit: 20000000,
			gasPrice: big.NewInt(1000000),
			postCheck: func(ctx sdk.Context) {
				// Check the user balance, should be zero since all was used for fees
				balance := app.BankKeeper.GetBalance(ctx, keys.GetKey(0).AccAddr, MockErc20Denom)
				require.True(t, balance.IsZero())
			},
		},
		{
			name: "success - fee with fee abstraction and transaction value native token",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Set up the token pair on the erc20 module
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  MockErc20Address,
					Denom:         MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Mint the tokens for the fee payer
				amount := sdk.NewCoins(sdk.NewInt64Coin(MockErc20Denom, 20000000*1000000*10))
				err := mintCoins(app, ctx, keys.GetKey(0).AccAddr, amount)
				require.NoError(t, err)

				// Mint the native token for the transaction value
				err = mintCoins(app, ctx, keys.GetKey(0).AccAddr, sdk.NewCoins(sdk.NewInt64Coin("akii", 1000000)))
				require.NoError(t, err)
				return ctx
			},
			gasLimit: 20000000,
			txAmount: 1000000,
			gasPrice: big.NewInt(1000000),
			postCheck: func(ctx sdk.Context) {
				// Check the user balance, should be zero since all was used for fees
				balance := app.BankKeeper.GetBalance(ctx, keys.GetKey(0).AccAddr, MockErc20Denom)
				require.True(t, balance.IsZero())

				// The user should have the transaction value in the native token
				// Since its only transferred during the state transition
				balance = app.BankKeeper.GetBalance(ctx, keys.GetKey(0).AccAddr, "akii")
				require.Equal(t, math.NewInt(1000000), balance.Amount)
			},
		},
		{
			name: "success - fee with fee abstraction and erc20 balance",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Deploy the erc20 token
				erc20Address, err := apptesting.DeployERC20(ctx, app)
				require.NoError(t, err)

				// Mint for our address
				err = apptesting.MintERC20(ctx, app, erc20Address, keys.GetAddr(0), big.NewInt(20000000*1000000*2))
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
						math.LegacyMustNewDecFromStr("2"),
						math.LegacyMustNewDecFromStr("2"),
					),
				))
				require.NoError(t, err)

				// Write up the token address on the context for reuse
				return ctx.WithValue("erc20_token", erc20Address)
			},
			gasLimit: 20000000,
			gasPrice: big.NewInt(1000000),
			postCheck: func(ctx sdk.Context) {
				// Read the erc20 token address from the context
				erc20Address, ok := ctx.Value("erc20_token").(common.Address)
				require.True(t, ok)

				// Check the user erc20 balance, should be zero
				erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
				erc20Balance := app.Erc20Keeper.BalanceOf(
					ctx,
					erc20,
					erc20Address,
					keys.GetAddr(0),
				)
				require.EqualValues(t, 0, erc20Balance.Int64())

				// Check the value on the FeeCollector
				feeCollectorBalance := app.BankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(authtypes.FeeCollectorName), "erc20/"+erc20Address.Hex())
				require.EqualValues(t, 20000000*1000000*2, feeCollectorBalance.Amount.Int64())
			},
		},
		{
			name:        "fail - not enough funds for the fee",
			gasLimit:    20000000,
			gasPrice:    big.NewInt(1000000),
			errContains: "insufficient funds for fee",
		},
		{
			name: "fail - not enough gas",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Fund the account with token to pay for the fees
				amount := sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(20000000*1000000)))
				err := mintCoins(app, ctx, keys.GetKey(0).AccAddr, amount)
				require.NoError(t, err)
				return ctx
			},
			gasLimit:    1, // This is less than the gas limit set in the decorator
			gasPrice:    big.NewInt(1000000),
			errContains: "gas limit too low",
		},
		{
			name: "fail - fee charged with fee abstraction native token but no funds for tx payment",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Set up the token pair on the erc20 module
				app.Erc20Keeper.SetToken(ctx, erc20types.TokenPair{
					Erc20Address:  MockErc20Address,
					Denom:         MockErc20Denom,
					Enabled:       true,
					ContractOwner: erc20types.OWNER_UNSPECIFIED,
				})

				// Mint the tokens for the fee payer
				amount := sdk.NewCoins(sdk.NewInt64Coin(MockErc20Denom, 20000000*1000000*10))
				err := mintCoins(app, ctx, keys.GetKey(0).AccAddr, amount)
				require.NoError(t, err)
				return ctx
			},
			gasLimit:    20000000,
			gasPrice:    big.NewInt(1000000),
			txAmount:    1000000, // This is the amount to be sent in the tx, this makes the tx fail since the user has no funds to pay for the tx value
			errContains: "insufficient funds",
		},
		{
			name: "fail - account is a contract",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Set the EVM account beforehand
				fromAddr := keys.GetKey(0).Addr

				// Create a contract account
				// To define as a contract we need to set the code hash
				err = app.EVMKeeper.SetAccount(ctx, fromAddr, statedb.Account{
					Nonce:    0,
					Balance:  big.NewInt(1000000),
					CodeHash: []byte("contract code"),
				})
				require.NoError(t, err)

				return ctx
			},
			gasLimit:    20000000,
			gasPrice:    big.NewInt(1000000),
			errContains: "the sender is not EOA",
		},
		{
			name: "fail - invalid amount",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Fund the account with token to pay for the fees
				amount := sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(20000000*1000000)))
				err := mintCoins(app, ctx, keys.GetKey(0).AccAddr, amount)
				require.NoError(t, err)
				return ctx
			},
			gasLimit:    20000000,
			gasPrice:    big.NewInt(1000000),
			txAmount:    -10,
			errContains: "(-10) is negative and invalid",
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a cached context
			cacheCtx, _ := ctx.CacheContext()
			cacheCtx = cacheCtx.WithBlockGasMeter(storetypes.NewGasMeter(20000000))

			// Malleate the context
			if tc.malleate != nil {
				cacheCtx = tc.malleate(cacheCtx)
			}

			// Start up the wrapped ante decorator
			monoDecorator := kiievmante.NewEVMMonoDecorator(
				app.AccountKeeper,
				app.FeeMarketKeeper,
				app.EVMKeeper,
				app.FeeAbstractionKeeper,
				20000000,
			)
			anteHandler := sdk.ChainAnteDecorators(monoDecorator)

			// Build the tx
			tx, err := createAndSignTx(keys.GetKey(0), tc.gasLimit, tc.gasPrice, tc.txAmount)
			require.NoError(t, err)

			// Execute the ante handler
			_, err = anteHandler(cacheCtx, tx, false)
			if tc.errContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			// Post check if needed
			if tc.postCheck != nil {
				tc.postCheck(cacheCtx)
			}
		})
	}
}

// mintCoins mints coins to the given account
func mintCoins(app *kiichain.KiichainApp, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) error {
	// Mint the coins to the module account
	err := app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, coins)
	if err != nil {
		return err
	}

	// Send the coins from the module account to the given address
	return app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, addr, coins)
}

// createAndSignTx creates and signs a transaction with the given key
func createAndSignTx(key keyring.Key, gasLimit uint64, gasPrice *big.Int, amount int64) (signing.Tx, error) {
	ethChainID := big.NewInt(1010)
	msgEthereumTx := evmtypes.NewTx(getDefaultEVMTxArgs(ethChainID, gasLimit, gasPrice, amount))

	signer := gethtypes.LatestSignerForChainID(ethChainID)

	msgEthereumTx.From = common.BytesToAddress(key.Priv.PubKey().Address().Bytes()).String()

	err := msgEthereumTx.Sign(signer, tx.NewSigner(key.Priv))
	if err != nil {
		return nil, err
	}

	// Start the tx builder
	encodingConfig := params.MakeEncodingConfig()
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	return msgEthereumTx.BuildTx(txBuilder, "akii")
}

// getDefaultEVMTxArgs returns the default EVM transaction arguments
func getDefaultEVMTxArgs(chainID *big.Int, gasLimit uint64, gasPrice *big.Int, amount int64) *evmtypes.EvmTxArgs {
	return &evmtypes.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     0,
		GasPrice:  gasPrice,
		GasFeeCap: big.NewInt(1000000000),
		GasLimit:  gasLimit,
		Amount:    big.NewInt(amount),
	}
}
