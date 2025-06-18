package wasmbinding

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	tfbinding "github.com/kiichain/kiichain/v2/wasmbinding/tokenfactory"
	tfbindingtypes "github.com/kiichain/kiichain/v2/wasmbinding/tokenfactory/types"
	"github.com/kiichain/kiichain/v2/wasmbinding/utils"
)

// KiichainMsg is the msg type for all cosmwasm bindings
type KiichainMsg struct {
	TokenFactory *tfbindingtypes.Msg `json:"token_factory,omitempty"`
}

// CustomMessageDecorator returns decorator for custom CosmWasm bindings messages
func CustomMessageDecorator(bank bankkeeper.Keeper, tokenFactory *tfbinding.CustomMessenger) func(wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &CustomMessenger{
			wrapped:      old,
			bank:         bank,
			tokenFactory: tokenFactory,
		}
	}
}

// CustomMessenger is a wrapper for the token factory message plugin
type CustomMessenger struct {
	wrapped      wasmkeeper.Messenger
	bank         bankkeeper.Keeper
	tokenFactory *tfbinding.CustomMessenger
}

// Ensure CustomMessenger implements the Messenger interface
var _ wasmkeeper.Messenger = (*CustomMessenger)(nil)

// DispatchMsg implements keeper.Messenger
func (m *CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, msgResponses [][]*types.Any, err error) {
	if msg.Custom != nil {
		// only handle the happy path where this is really creating / minting / swapping ...
		// leave everything else foro the wrapped version
		var contractMsg KiichainMsg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			return nil, nil, utils.EmptyMsgResp, errorsmod.Wrap(err, "error parsing message into KiichainMsg")
		}

		// Match the message
		switch {
		case contractMsg.TokenFactory != nil:
			// Call the token factory custom message handler
			return m.tokenFactory.DispatchMsg(ctx, contractAddr, contractIBCPortID, *contractMsg.TokenFactory)
		default:
			return nil, nil, utils.EmptyMsgResp, wasmvmtypes.UnsupportedRequest{Kind: "unknown kiichain msg variant"}
		}

	}
	return m.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}
