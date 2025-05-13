package wasmbinding

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	tfbinding "github.com/kiichain/kiichain/v1/wasmbinding/tokenfactory"
	bindingtypes "github.com/kiichain/kiichain/v1/wasmbinding/tokenfactory/types"
)

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
		var contractMsg bindingtypes.Msg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			return nil, nil, bindingtypes.EmptyMsgResp, errorsmod.Wrap(err, "token factory msg")
		}

		// Match the message
		switch {
		case contractMsg.CreateDenom != nil:
			return m.tokenFactory.CreateDenom(ctx, contractAddr, contractMsg.CreateDenom)
		case contractMsg.MintTokens != nil:
			return m.tokenFactory.MintTokens(ctx, contractAddr, contractMsg.MintTokens)
		case contractMsg.ChangeAdmin != nil:
			return m.tokenFactory.ChangeAdmin(ctx, contractAddr, contractMsg.ChangeAdmin)
		case contractMsg.BurnTokens != nil:
			return m.tokenFactory.BurnTokens(ctx, contractAddr, contractMsg.BurnTokens)
		case contractMsg.SetMetadata != nil:
			return m.tokenFactory.SetMetadata(ctx, contractAddr, contractMsg.SetMetadata)
		case contractMsg.ForceTransfer != nil:
			return m.tokenFactory.ForceTransfer(ctx, contractAddr, contractMsg.ForceTransfer)
		default:
			return nil, nil, bindingtypes.EmptyMsgResp, wasmvmtypes.UnsupportedRequest{Kind: "unknown token factory msg variant"}
		}

	}
	return m.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}
