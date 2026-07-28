package apptesting

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/evm/contracts"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	"github.com/cosmos/evm/x/vm/statedb"

	app "github.com/kiichain/kiichain/v7/app"
)

// ERC20DeployerSeed is used to derive a deterministic non-module deployer for
// test ERC20 deployments. Module accounts cannot receive balance updates via
// SetBalanceWithLocked after the Cosmos EVM locked-balance hotfix.
const ERC20DeployerSeed = "kiichain/apptesting/erc20-deployer"

// ERC20DeployerAddress returns the deterministic EOA used to deploy/mint test
// ERC20 contracts.
func ERC20DeployerAddress() common.Address {
	seed := crypto.Keccak256([]byte(ERC20DeployerSeed))
	key, err := crypto.ToECDSA(seed)
	if err != nil {
		panic(err)
	}
	return crypto.PubkeyToAddress(key.PublicKey)
}

// DefaultFirstERC20 is the contract address produced by the first DeployERC20
// call (deployer nonce 0).
var DefaultFirstERC20 = crypto.CreateAddress(ERC20DeployerAddress(), 0).Hex()

// ensureAccountExists creates the account if missing so nonce tracking works.
func ensureAccountExists(ctx sdk.Context, app *app.KiichainApp, addr sdk.AccAddress) {
	if app.AccountKeeper.GetAccount(ctx, addr) == nil {
		app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, addr))
	}
}

// DeployERC20 deploys an ERC20 contract from a deterministic non-module EOA.
func DeployERC20(ctx sdk.Context, app *app.KiichainApp) (common.Address, error) {
	from := ERC20DeployerAddress()
	ensureAccountExists(ctx, app, sdk.AccAddress(from.Bytes()))

	// Set the data
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI
	ctorArgs, err := erc20ABI.Pack("", "Test", "TEST", uint8(18))
	if err != nil {
		return common.Address{}, err
	}
	deployData := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...) //nolint:gocritic

	// Deploy the contract
	stateDB := statedb.New(ctx, app.EVMKeeper, statedb.NewEmptyTxConfig())
	res, err := app.EVMKeeper.CallEVMWithData(ctx, stateDB, from, nil, deployData, true, false, nil)
	if err != nil {
		return common.Address{}, err
	}
	if res == nil || res.Ret == nil {
		return common.Address{}, errorsmod.Wrap(erc20types.ErrEVMCall, "failed to deploy ERC20 contract: empty response")
	}

	// Derive the deployed contract address
	nonce := app.EVMKeeper.GetNonce(ctx, from)
	contractAddr := crypto.CreateAddress(from, nonce-1)
	return contractAddr, nil
}

// MintERC20 mints an ERC20 token using the same deployer EOA that owns the
// contract from DeployERC20.
func MintERC20(ctx sdk.Context, app *app.KiichainApp, contractAddr common.Address, to common.Address, amount *big.Int) error {
	from := ERC20DeployerAddress()
	ensureAccountExists(ctx, app, sdk.AccAddress(from.Bytes()))

	// Load the ABI and pack the mint() call
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI
	inputData, err := erc20ABI.Pack("mint", to, amount)
	if err != nil {
		return err
	}

	// Send transaction to call mint
	stateDB := statedb.New(ctx, app.EVMKeeper, statedb.NewEmptyTxConfig())
	_, err = app.EVMKeeper.CallEVMWithData(ctx, stateDB, from, &contractAddr, inputData, true, false, nil)
	if err != nil {
		return err
	}

	return nil
}

// CreateERC20Allowance creates an ERC20 allowance
func CreateERC20Allowance(ctx sdk.Context, app *app.KiichainApp, contractAddr common.Address, owner common.Address, spender common.Address, amount *big.Int) error {
	// Load the ABI and pack the approve() call
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI
	inputData, err := erc20ABI.Pack("approve", spender, amount)
	if err != nil {
		return err
	}

	stateDB := statedb.New(ctx, app.EVMKeeper, statedb.NewEmptyTxConfig())
	_, err = app.EVMKeeper.CallEVMWithData(ctx, stateDB, owner, &contractAddr, inputData, true, false, nil)
	if err != nil {
		return err
	}
	return nil
}
