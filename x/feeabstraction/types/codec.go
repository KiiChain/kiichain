package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

const (
	MsgUpdateParamsName = "feeabstraction/update-params"
)

// RegisterInterfaces register all the proto interfaces into the app
func RegisterInterfaces(r codectypes.InterfaceRegistry) {
	// Register all the proto interfaces
	r.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)

	// Register on the message service
	msgservice.RegisterMsgServiceDesc(r, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec register the interface for legacy amino support
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register all the concrete types
	cdc.RegisterConcrete(&MsgUpdateParams{}, MsgUpdateParamsName, nil)
}
