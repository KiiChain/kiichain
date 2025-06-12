package tokenfactory_test

import (
	"encoding/json"
	"fmt"
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	app "github.com/kiichain/kiichain/v1/app"
	"github.com/kiichain/kiichain/v1/app/apptesting"
	"github.com/kiichain/kiichain/v1/wasmbinding"
	"github.com/kiichain/kiichain/v1/wasmbinding/helpers"
	bindingtypes "github.com/kiichain/kiichain/v1/wasmbinding/tokenfactory/types"
)

// TestQueryDenomAdmin tests the GetDenomAdmin query
func TestQueryFullDenom(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	reflect := helpers.InstantiateReflectContract(t, ctx, app, actor)
	require.NotEmpty(t, reflect)

	// query full denom
	query := bindingtypes.Query{
		FullDenom: &bindingtypes.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "ustart",
		},
	}
	resp := bindingtypes.FullDenomResponse{}
	queryCustom(t, ctx, app, reflect, query, &resp)

	expected := fmt.Sprintf("factory/%s/ustart", reflect.String())
	require.EqualValues(t, expected, resp.Denom)
}

// TestQueryDenomAdmin tests the GetDenomAdmin query
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
func queryCustom(t *testing.T, ctx sdk.Context, app *app.KiichainApp, contract sdk.AccAddress, request bindingtypes.Query, response interface{}) {
	t.Helper()

	// Make the request a kiichain query
	kiichainQuery := wasmbinding.KiichainQuery{
		TokenFactory: &request,
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
