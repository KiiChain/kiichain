package helpers

import (
	"encoding/json"
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v3/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	erc20types "github.com/cosmos/evm/x/erc20/types"

	app "github.com/kiichain/kiichain/v6/app"
	mock "github.com/kiichain/kiichain/v6/tests/e2e/mock"
	"github.com/kiichain/kiichain/v6/wasmbinding"
	evmbindingtypes "github.com/kiichain/kiichain/v6/wasmbinding/evm/types"
)

// Get the reentrancy ABI
var ReentrancyABI, _ = mock.ReentranceMetaData.GetAbi()

// DeployReentrancy deploys the reentrancy contract and returns its address
func DeployReentrancy(t *testing.T, ctx sdk.Context, app *app.KiichainApp) common.Address {
	t.Helper()

	// Set the from as the ERC20 module address, since it has the permission to deploy contracts
	from := common.BytesToAddress(authtypes.NewModuleAddress(erc20types.ModuleName).Bytes())

	// Set the abi data
	ctorArgs, err := ReentrancyABI.Pack("")
	require.NoError(t, err)
	deployData := append(common.FromHex(mock.ReentranceBin), ctorArgs...)

	// Deploy the contract
	res, err := app.EVMKeeper.CallEVMWithData(ctx, from, nil, deployData, true, nil)
	require.NoError(t, err)
	require.NotNil(t, res.Ret)

	// Derive the deployed contract address
	nonce := app.EVMKeeper.GetNonce(ctx, from)
	contractAddr := crypto.CreateAddress(from, nonce-1)
	return contractAddr
}

// BuildReflectChainQueryForEthCall builds a reflect query for EthCall
func BuildReflectChainQueryForEthCall(t *testing.T, addr common.Address, data string) []byte {
	t.Helper()

	// Inner: KiichainQuery (EVM.EthCall)
	evmQuery := evmbindingtypes.Query{
		EthCall: &evmbindingtypes.EthCallRequest{
			Contract: addr.Hex(),
			Data:     data,
		},
	}
	kiiQuery := wasmbinding.KiichainQuery{EVM: &evmQuery}
	kiiBz, err := json.Marshal(kiiQuery)
	require.NoError(t, err)

	// Wrap into Reflect QueryMsg::Chain{ request: QueryRequest::Custom(kiiBz) }
	type chainReq struct {
		Request wasmvmtypes.QueryRequest `json:"request"`
	}
	type reflectQuery struct {
		Chain *chainReq `json:"chain,omitempty"`
	}
	rq := reflectQuery{
		Chain: &chainReq{
			Request: wasmvmtypes.QueryRequest{Custom: kiiBz},
		},
	}
	final, err := json.Marshal(rq)
	require.NoError(t, err)
	return final
}
