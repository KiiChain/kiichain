package keeper

import (
	"fmt"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		paramstore   paramtypes.Subspace
		authority    string
		Schema       collections.Schema

		// PendingSendPackets stores packets whose send flow was applied and may need
		// to be reverted on timeout or error acknowledgement. The key order is
		// (channelId, denom, sequence) to support efficient channel+denom range resets.
		PendingSendPackets collections.KeySet[collections.Triple[string, string, uint64]]
		// PendingReceivePackets stores packets whose receive flow was applied and may
		// need to be reverted when an async acknowledgement fails. The key order is
		// (channelId, denom, sequence) to support efficient channel+denom range resets.
		PendingReceivePackets collections.KeySet[collections.Triple[string, string, uint64]]

		bankKeeper    types.BankKeeper
		channelKeeper types.ChannelKeeper
		clientKeeper  types.ClientKeeper
		ics4Wrapper   types.ICS4Wrapper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	ps paramtypes.Subspace,
	authority string,
	bankKeeper types.BankKeeper,
	channelKeeper types.ChannelKeeper,
	clientKeeper types.ClientKeeper,
	ics4Wrapper types.ICS4Wrapper,
) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	pendingPacketKeyCodec := collections.TripleKeyCodec(collections.StringKey, collections.StringKey, collections.Uint64Key)
	k := Keeper{
		cdc:          cdc,
		storeService: storeService,
		paramstore:   ps,
		authority:    authority,

		PendingSendPackets:    collections.NewKeySet(sb, types.PendingSendPacketsKey, "pending_send_packets", pendingPacketKeyCodec),
		PendingReceivePackets: collections.NewKeySet(sb, types.PendingReceivePacketsKey, "pending_receive_packets", pendingPacketKeyCodec),

		bankKeeper:    bankKeeper,
		channelKeeper: channelKeeper,
		clientKeeper:  clientKeeper,
		ics4Wrapper:   ics4Wrapper,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return &k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetICS4Wrapper sets the ICS4 wrapper used to send packets.
func (k *Keeper) SetICS4Wrapper(wrapper types.ICS4Wrapper) {
	k.ics4Wrapper = wrapper
}

// SetIBCKeepers allows us to set the relevant IBC keepers post dependency
// injection, as IBC doesn't support dependency injection yet.
func (k *Keeper) SetIBCKeepers(channelKeeper types.ChannelKeeper, clientKeeper types.ClientKeeper, ics4Wrapper types.ICS4Wrapper) {
	k.channelKeeper = channelKeeper
	k.clientKeeper = clientKeeper
	k.ics4Wrapper = ics4Wrapper
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
