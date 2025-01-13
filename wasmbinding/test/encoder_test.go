package wasmbinding

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain3/wasmbinding/bindings"
	tokenfactorywasm "github.com/kiichain/kiichain3/x/tokenfactory/client/wasm"
	tokenfactorytypes "github.com/kiichain/kiichain3/x/tokenfactory/types"
	"github.com/stretchr/testify/require"
)

const (
	TEST_TARGET_CONTRACT = "kii14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sr8zwk6"
	TEST_CREATOR         = "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs"
)

func TestEncodeCreateDenom(t *testing.T) {
	contractAddr, err := sdk.AccAddressFromBech32("kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs")
	require.NoError(t, err)
	msg := bindings.CreateDenom{
		Subdenom: "subdenom",
	}
	serializedMsg, _ := json.Marshal(msg)

	decodedMsgs, err := tokenfactorywasm.EncodeTokenFactoryCreateDenom(serializedMsg, contractAddr)
	require.NoError(t, err)
	require.Equal(t, 1, len(decodedMsgs))
	typedDecodedMsg, ok := decodedMsgs[0].(*tokenfactorytypes.MsgCreateDenom)
	require.True(t, ok)
	expectedMsg := tokenfactorytypes.MsgCreateDenom{
		Sender:   "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
		Subdenom: "subdenom",
	}
	require.Equal(t, expectedMsg, *typedDecodedMsg)
}

func TestEncodeMint(t *testing.T) {
	contractAddr, err := sdk.AccAddressFromBech32("kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs")
	require.NoError(t, err)
	msg := bindings.MintTokens{
		Amount: sdk.Coin{Amount: sdk.NewInt(100), Denom: "subdenom"},
	}
	serializedMsg, _ := json.Marshal(msg)

	decodedMsgs, err := tokenfactorywasm.EncodeTokenFactoryMint(serializedMsg, contractAddr)
	require.NoError(t, err)
	require.Equal(t, 1, len(decodedMsgs))
	typedDecodedMsg, ok := decodedMsgs[0].(*tokenfactorytypes.MsgMint)
	require.True(t, ok)
	expectedMsg := tokenfactorytypes.MsgMint{
		Sender: "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
		Amount: sdk.Coin{Amount: sdk.NewInt(100), Denom: "subdenom"},
	}
	require.Equal(t, expectedMsg, *typedDecodedMsg)
}

func TestEncodeBurn(t *testing.T) {
	contractAddr, err := sdk.AccAddressFromBech32("kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs")
	require.NoError(t, err)
	msg := bindings.BurnTokens{
		Amount: sdk.Coin{Amount: sdk.NewInt(10), Denom: "subdenom"},
	}
	serializedMsg, _ := json.Marshal(msg)

	decodedMsgs, err := tokenfactorywasm.EncodeTokenFactoryBurn(serializedMsg, contractAddr)
	require.NoError(t, err)
	require.Equal(t, 1, len(decodedMsgs))
	typedDecodedMsg, ok := decodedMsgs[0].(*tokenfactorytypes.MsgBurn)
	require.True(t, ok)
	expectedMsg := tokenfactorytypes.MsgBurn{
		Sender: "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
		Amount: sdk.Coin{Amount: sdk.NewInt(10), Denom: "subdenom"},
	}
	require.Equal(t, expectedMsg, *typedDecodedMsg)
}

func TestEncodeChangeAdmin(t *testing.T) {
	contractAddr, err := sdk.AccAddressFromBech32("kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs")
	require.NoError(t, err)
	msg := bindings.ChangeAdmin{
		Denom:           "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/subdenom",
		NewAdminAddress: "kii1hjfwcza3e3uzeznf3qthhakdr9juetl7uajv0t",
	}
	serializedMsg, _ := json.Marshal(msg)

	decodedMsgs, err := tokenfactorywasm.EncodeTokenFactoryChangeAdmin(serializedMsg, contractAddr)
	require.NoError(t, err)
	require.Equal(t, 1, len(decodedMsgs))
	typedDecodedMsg, ok := decodedMsgs[0].(*tokenfactorytypes.MsgChangeAdmin)
	require.True(t, ok)
	expectedMsg := tokenfactorytypes.MsgChangeAdmin{
		Sender:   "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
		Denom:    "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/subdenom",
		NewAdmin: "kii1hjfwcza3e3uzeznf3qthhakdr9juetl7uajv0t",
	}
	require.Equal(t, expectedMsg, *typedDecodedMsg)
}
