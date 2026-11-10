package v1beta1

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces keeps the mainnet v1beta1 messages unpackable so
// proposals that used the old type URLs still decode after the v1 bump.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgFundPool{},
		&MsgUpdateParams{},
		&MsgChangeSchedule{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec keeps the mainnet amino names.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateParams{}, "rewards/update-params", nil)
	cdc.RegisterConcrete(&MsgFundPool{}, "rewards/fund-pool", nil)
	cdc.RegisterConcrete(&MsgChangeSchedule{}, "rewards/change-schedule", nil)
}
