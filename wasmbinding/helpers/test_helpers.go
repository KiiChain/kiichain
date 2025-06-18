package helpers

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tenderminttypes "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"

	app "github.com/kiichain/kiichain/v2/app"
	"github.com/kiichain/kiichain/v2/app/helpers"
	tokenfactorytypes "github.com/kiichain/kiichain/v2/x/tokenfactory/types"
)

// CreateTestInput initializes a new test input for the app and returns the app instance and context
func CreateTestInput(t *testing.T) (*app.KiichainApp, sdk.Context) {
	t.Helper()
	chain := helpers.Setup(t)
	ctx := chain.BaseApp.NewUncachedContext(true, tenderminttypes.Header{Height: 1, ChainID: "testing_1010-1", Time: time.Now().UTC()})
	allVal, err := chain.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)

	// Validator consensus address
	valConsAddr, err := allVal[0].GetConsAddr()
	require.NoError(t, err)

	// Set a final context with the proposer address for the EVM module
	ctx = chain.BaseApp.NewUncachedContext(true, tenderminttypes.Header{Height: 1, ChainID: "testing_1010-1", Time: time.Now().UTC(), ProposerAddress: valConsAddr})
	return chain, sdk.UnwrapSDKContext(ctx)
}

// we need to make this deterministic (same every test run), as content might affect gas costs
func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	key := ed25519.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

// RandomAccountAddress generates a random account address
func RandomAccountAddress() sdk.AccAddress {
	_, _, addr := keyPubAddr()
	return addr
}

// RandomBech32AccountAddress generates a random bech32 account address
func RandomBech32AccountAddress() string {
	return RandomAccountAddress().String()
}

// storeReflectCode stores the reflect contract code
func storeReflectCode(t *testing.T, ctx sdk.Context, app *app.KiichainApp, addr sdk.AccAddress) uint64 {
	t.Helper()
	wasmCode, err := os.ReadFile("../testdata/token_reflect.wasm")
	require.NoError(t, err)

	contractKeeper := keeper.NewDefaultPermissionKeeper(app.WasmKeeper)
	codeID, _, err := contractKeeper.Create(ctx, addr, wasmCode, nil)
	require.NoError(t, err)

	return codeID
}

// InstantiateReflectContract instantiates the reflect contract
func InstantiateReflectContract(t *testing.T, ctx sdk.Context, app *app.KiichainApp, funder sdk.AccAddress) sdk.AccAddress {
	t.Helper()
	initMsgBz := []byte("{}")
	contractKeeper := keeper.NewDefaultPermissionKeeper(app.WasmKeeper)
	codeID := uint64(1)
	addr, _, err := contractKeeper.Instantiate(ctx, codeID, funder, funder, initMsgBz, "demo contract", nil)
	require.NoError(t, err)

	return addr
}

// FundAccount funds an account with specified coins
func FundAccount(t *testing.T, ctx sdk.Context, app *app.KiichainApp, addr sdk.AccAddress, coins sdk.Coins) {
	t.Helper()
	err := app.BankKeeper.MintCoins(ctx, tokenfactorytypes.ModuleName, coins)
	require.NoError(t, err)

	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, tokenfactorytypes.ModuleName, addr, coins)
	require.NoError(t, err)
}

// SetupCustomApp sets up a custom app for testing
func SetupCustomApp(t *testing.T, addr sdk.AccAddress) (*app.KiichainApp, sdk.Context) {
	t.Helper()

	// Start a new app
	app, ctx := CreateTestInput(t)
	wasmKeeper := app.WasmKeeper

	// Store the reflect code
	storeReflectCode(t, ctx, app, addr)

	// Check if all is correct with the code
	cInfo := wasmKeeper.GetCodeInfo(ctx, 1)
	require.NotNil(t, cInfo)

	return app, ctx
}
