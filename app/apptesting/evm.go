package apptesting

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/evm/contracts"
	erc20types "github.com/cosmos/evm/x/erc20/types"

	app "github.com/kiichain/kiichain/v3/app"
)

// DeployERC20 deploys an ERC20 contract
func DeployERC20(ctx sdk.Context, app *app.KiichainApp) (common.Address, error) {
	// Select the from as the erc20 module address
	from := common.BytesToAddress(authtypes.NewModuleAddress(erc20types.ModuleName).Bytes())

	// Set the data
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI
	ctorArgs, err := erc20ABI.Pack("", "Test", "TEST", uint8(18))
	if err != nil {
		return common.Address{}, err
	}
	deployData := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...) //nolint:gocritic

	// Deploy the contract
	res, err := app.EVMKeeper.CallEVMWithData(ctx, from, nil, deployData, true)
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

// MintERC20 mints an ERC20 token
func MintERC20(ctx sdk.Context, app *app.KiichainApp, contractAddr common.Address, to common.Address, amount *big.Int) error {
	// Sender must be an account with ETH balance and nonce tracking
	from := common.BytesToAddress(authtypes.NewModuleAddress(erc20types.ModuleName).Bytes())

	// Load the ABI and pack the mint() call
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI
	inputData, err := erc20ABI.Pack("mint", to, amount)
	if err != nil {
		return err
	}

	// Send transaction to call mint
	_, err = app.EVMKeeper.CallEVMWithData(ctx, from, &contractAddr, inputData, true)
	if err != nil {
		return err
	}

	return nil
}

// CreateERC20Allowance creates an ERC20 allowance
func CreateERC20Allowance(ctx sdk.Context, app *app.KiichainApp, contractAddr common.Address, owner common.Address, spender common.Address, amount *big.Int) error {
	// Load the ABI and pack the mint() call
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI
	inputData, err := erc20ABI.Pack("approve", spender, amount)
	if err != nil {
		return err
	}

	// Send transaction to call mint
	_, err = app.EVMKeeper.CallEVMWithData(ctx, owner, &contractAddr, inputData, true)
	if err != nil {
		return err
	}
	return nil
}
