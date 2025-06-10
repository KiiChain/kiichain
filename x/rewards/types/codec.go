package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces register interfaces into the app
func RegisterInterfaces(registry types.InterfaceRegistry) {
	// Register messages
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgFundPool{},
		&MsgExtendReward{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec registers the necessary x/rewards interfaces
// and concrete types on the provided LegacyAmino codec. These types are used
// for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register all your concrete types
	cdc.RegisterConcrete(&MsgUpdateParams{}, "rewards/update-params", nil)
	cdc.RegisterConcrete(&MsgFundPool{}, "rewards/fund-pool", nil)
	cdc.RegisterConcrete(&MsgExtendReward{}, "rewards/extend-reward", nil)
}
