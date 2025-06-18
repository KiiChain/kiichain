package bech32

import (
	"encoding/json"
	"fmt"
	"strings"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bech32bindingtypes "github.com/kiichain/kiichain/v2/wasmbinding/bech32/types"
)

// QueryPlugin is the query plugin object for the bech32 queries
type QueryPlugin struct{}

// NewQueryPlugin returns a new query plugin
func NewQueryPlugin() *QueryPlugin {
	return &QueryPlugin{}
}

// HandleBech32Query is a custom querier for the bech32 module
func (qp *QueryPlugin) HandleBech32Query(ctx sdk.Context, bech32Query bech32bindingtypes.Query) ([]byte, error) {
	// Match the query under the module
	switch {
	// The query is a hex to bech32 query
	case bech32Query.HexToBech32 != nil:
		// Apply the request
		address, err := HandleHexToBech32(*bech32Query.HexToBech32)
		if err != nil {
			return nil, err
		}

		// Marshal the response
		bz, err := json.Marshal(address)
		if err != nil {
			return nil, err
		}
		return bz, nil

	// The query is a bech32 to hex query
	case bech32Query.Bech32ToHex != nil:
		// Apply the request
		address, err := HandleBech32ToHex(*bech32Query.Bech32ToHex)
		if err != nil {
			return nil, err
		}

		// Marshal the response
		bz, err := json.Marshal(address)
		if err != nil {
			return nil, err
		}

		return bz, nil
	default:
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown bech32 query variant"}
	}
}

// HandleHexToBech32 handles the hex to bech32 conversion
func HandleHexToBech32(req bech32bindingtypes.HexToBech32) (*bech32bindingtypes.HexToBech32Response, error) {
	// Convert the hex string to a address
	address := common.HexToAddress(req.Address)

	// Convert the address to a bech32 address
	bech32Address, err := sdk.Bech32ifyAddressBytes(req.Prefix, address.Bytes())
	if err != nil {
		return nil, err
	}

	// Return the response
	return &bech32bindingtypes.HexToBech32Response{
		Address: bech32Address,
	}, nil
}

// HandleBech32ToHex handles the bech32 to hex conversion
func HandleBech32ToHex(req bech32bindingtypes.Bech32ToHex) (*bech32bindingtypes.Bech32ToHexResponse, error) {
	bech32Prefix := strings.SplitN(req.Address, "1", 2)[0]
	if bech32Prefix == req.Address {
		return nil, fmt.Errorf("invalid bech32 address: %s", req.Address)
	}

	addressBz, err := sdk.GetFromBech32(req.Address, bech32Prefix)
	if err != nil {
		return nil, err
	}

	// Check if the address is valid
	if err := sdk.VerifyAddressFormat(addressBz); err != nil {
		return nil, err
	}

	// Return the response
	return &bech32bindingtypes.Bech32ToHexResponse{
		Address: common.BytesToAddress(addressBz).String(),
	}, nil
}
