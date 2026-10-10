package keeper

import (
	"time"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	// Set rate limits, blacklists, and whitelists
	for _, rateLimit := range genState.RateLimits {
		k.SetRateLimit(ctx, rateLimit)
	}
	for _, denom := range genState.BlacklistedDenoms {
		k.AddDenomToBlacklist(ctx, denom)
	}
	for _, addressPair := range genState.WhitelistedAddressPairs {
		k.SetWhitelistedAddressPair(ctx, addressPair)
	}

	// Set pending sequence numbers - validating that they're in right format of {channelId}/{sequenceNumber}/{denom}
	for _, pendingPacketId := range genState.PendingSendPacketSequenceNumbers {
		channelOrClientId, sequence, denom, err := types.ParsePendingPacketId(pendingPacketId)
		if err != nil {
			if types.IsLegacyPendingPacketId(pendingPacketId) {
				continue
			}
			panic(err.Error())
		}
		if err := k.SetPendingSendPacket(ctx, channelOrClientId, sequence, denom); err != nil {
			panic(err)
		}
	}
	for _, pendingPacketId := range genState.PendingRecvPacketSequenceNumbers {
		channelOrClientId, sequence, denom, err := types.ParsePendingPacketId(pendingPacketId)
		if err != nil {
			if types.IsLegacyPendingPacketId(pendingPacketId) {
				continue
			}
			panic(err.Error())
		}
		if err := k.SetPendingReceivePacket(ctx, channelOrClientId, sequence, denom); err != nil {
			panic(err)
		}
	}

	// If the hour epoch has been initialized already (epoch number != 0), validate and then use it
	if genState.HourEpoch.EpochNumber > 0 {
		k.SetHourEpoch(ctx, genState.HourEpoch)
	} else {
		// If the hour epoch has not been initialized yet, set it so that the epoch number matches
		// the current hour and the start time is precisely on the hour
		genState.HourEpoch.EpochNumber = uint64(ctx.BlockTime().Hour()) //nolint:gosec
		genState.HourEpoch.EpochStartTime = ctx.BlockTime().Truncate(time.Hour)
		genState.HourEpoch.EpochStartHeight = ctx.BlockHeight()
		k.SetHourEpoch(ctx, genState.HourEpoch)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()

	pendingSendPackets, err := k.GetAllPendingSendPackets(ctx)
	if err != nil {
		panic(err)
	}
	pendingReceivePackets, err := k.GetAllPendingReceivePackets(ctx)
	if err != nil {
		panic(err)
	}

	genesis.Params = k.GetParams(ctx)
	genesis.RateLimits = k.GetAllRateLimits(ctx)
	genesis.BlacklistedDenoms = k.GetAllBlacklistedDenoms(ctx)
	genesis.WhitelistedAddressPairs = k.GetAllWhitelistedAddressPairs(ctx)
	genesis.PendingSendPacketSequenceNumbers = pendingSendPackets
	genesis.PendingRecvPacketSequenceNumbers = pendingReceivePackets
	genesis.HourEpoch = k.GetHourEpoch(ctx)

	return genesis
}
