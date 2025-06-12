package bech32_test

import (
	"encoding/json"
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	app "github.com/kiichain/kiichain/v1/app"
	"github.com/kiichain/kiichain/v1/app/apptesting"
	"github.com/kiichain/kiichain/v1/wasmbinding"
	bech32bindingtypes "github.com/kiichain/kiichain/v1/wasmbinding/bech32/types"
	"github.com/kiichain/kiichain/v1/wasmbinding/helpers"
)

// TestHexToBech32Reflect tests the hex to bech32 reflect query
func TestHexToBech32Reflect(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	reflect := helpers.InstantiateReflectContract(t, ctx, app, actor)
	require.NotEmpty(t, reflect)

	// query full denom
	query := bech32bindingtypes.Query{
		HexToBech32: &bech32bindingtypes.HexToBech32{
			Address: "0x7cB61D4117AE31a12E393a1Cfa3BaC666481D02E",
			Prefix:  "kii",
		},
	}

	resp := bech32bindingtypes.HexToBech32Response{}
	queryCustom(t, ctx, app, reflect, query, &resp)
	require.Equal(t, "kii10jmp6sgh4cc6zt3e8gw05wavvejgr5pwfe2u6n", resp.Address)
}

// TestBech32ToHexReflect tests the bech32 to hex reflect query
func TestBech32ToHexReflect(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	reflect := helpers.InstantiateReflectContract(t, ctx, app, actor)
	require.NotEmpty(t, reflect)

	// query full denom
	query := bech32bindingtypes.Query{
		Bech32ToHex: &bech32bindingtypes.Bech32ToHex{
			Address: "kii10jmp6sgh4cc6zt3e8gw05wavvejgr5pwfe2u6n",
		},
	}

	resp := bech32bindingtypes.Bech32ToHexResponse{}
	queryCustom(t, ctx, app, reflect, query, &resp)
	require.Equal(t, "0x7cB61D4117AE31a12E393a1Cfa3BaC666481D02E", resp.Address)
}

// ReflectQuery is the wrapper for the reflect query
type ReflectQuery struct {
	Chain *ChainRequest `json:"chain,omitempty"`
}

// ChainRequest is the request to the chain
type ChainRequest struct {
	Request wasmvmtypes.QueryRequest `json:"request"`
}

// ChainResponse is the response from the chain
type ChainResponse struct {
	Data []byte `json:"data"`
}

// queryCustom is a helper function to query the custom contract
func queryCustom(t *testing.T, ctx sdk.Context, app *app.KiichainApp, contract sdk.AccAddress, request bech32bindingtypes.Query, response interface{}) {
	t.Helper()

	// Make the request a kiichain query
	kiichainQuery := wasmbinding.KiichainQuery{
		Bech32: &request,
	}

	// Marshal the request to JSON
	msgBz, err := json.Marshal(kiichainQuery)
	require.NoError(t, err)
	t.Log("queryCustom1", string(msgBz))

	query := ReflectQuery{
		Chain: &ChainRequest{
			Request: wasmvmtypes.QueryRequest{Custom: msgBz},
		},
	}
	queryBz, err := json.Marshal(query)
	require.NoError(t, err)
	t.Log("queryCustom3", string(queryBz))

	resBz, err := app.WasmKeeper.QuerySmart(ctx, contract, queryBz)
	require.NoError(t, err)
	var resp ChainResponse
	err = json.Unmarshal(resBz, &resp)
	require.NoError(t, err)
	err = json.Unmarshal(resp.Data, response)
	require.NoError(t, err)
}
