package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/kiichain/kiichain3/x/oracle/types"
)

// Keeper manages the oracle module's state
type Keeper struct {
	cdc        codec.BinaryCodec // Codec for binary serialization
	storeKey   sdk.StoreKey      // storage key to access the module's state
	memKey     sdk.StoreKey
	paramSpace paramstypes.Subspace // Manages the module's parameters allowing dynamical settings

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	StakingKeeper types.StakingKeeper

	distrName string
}

// NewKeeper creates an oracle Keeper instance
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, memKey sdk.StoreKey, paramSpace paramstypes.Subspace,
	accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, StakingKeeper types.StakingKeeper,
	distrName string) Keeper {
	// Ensure oracle module account is set
	addr := accountKeeper.GetModuleAddress(types.ModuleName)
	if addr != nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// Ensure paramstore is properly initialized
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramSpace:    paramSpace,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		StakingKeeper: StakingKeeper,
		distrName:     distrName,
	}
}

// Logger is used to define custom Log for the module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// **************************** EXCHANGE RATE LOGIC ***************************
// SetBaseExchangeRate is used to get the exchange rate by denom from the KVStore
func (k Keeper) GetBaseExchangeRate(ctx sdk.Context, denom string) (sdk.Dec, sdk.Int, int64, error) {
	// Get ExchangeRate from KVStore
	store := ctx.KVStore(k.storeKey) // Get the oracle module's store
	byteData := store.Get(types.GetExchangeRateKey(denom))
	if byteData == nil {
		return sdk.ZeroDec(), sdk.ZeroInt(), 0, sdkerrors.Wrap(types.ErrUnknownDenom, denom)
	}

	// Decode ExchangeRate
	exchangeRate := &types.OracleExchangeRate{}
	k.cdc.MustUnmarshal(byteData, exchangeRate)
	return exchangeRate.ExchangeRate, exchangeRate.LastUpdate, exchangeRate.LastUpdateTimestamp, nil
}

// SetBaseExchangeRate is used to set the exchange rate by denom from the KVStore
func (k Keeper) SetBaseExchangeRate(ctx sdk.Context, denom string, exchangeRate sdk.Dec) {
	store := ctx.KVStore(k.storeKey) // Get the oracle module's store
	currentHeight := sdk.NewInt(ctx.BlockHeight())
	blockTimestamp := ctx.BlockTime().UnixMilli()

	rate := types.OracleExchangeRate{
		ExchangeRate:        exchangeRate,
		LastUpdate:          currentHeight,
		LastUpdateTimestamp: blockTimestamp,
	}

	byteData := k.cdc.MustMarshal(&rate)
	store.Set(types.GetExchangeRateKey(denom), byteData)
}

// SetBaseExchangeRateWithEvent calls SetBaseExchangeRate and generate an event about that denom creation
func (k Keeper) SetBaseExchangeRateWithEvent(ctx sdk.Context, denom string, exchangeRate sdk.Dec) {
	// Set exchange rate by denom
	k.SetBaseExchangeRate(ctx, denom, exchangeRate)

	// Create event
	event := sdk.NewEvent(
		types.EventTypeExchangeRateUpdate,
		sdk.NewAttribute(types.AttributeKeyDenom, denom),
		sdk.NewAttribute(types.AttributeKeyExchangeRate, exchangeRate.String()),
	)

	// Emit event
	ctx.EventManager().EmitEvent(event)
}

// DeleteBaseExchangeRate deletes an exchange rate by denom
func (k Keeper) DeleteBaseExchangeRate(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetExchangeRateKey(denom))
}

// IterateBaseExchangeRates iterate over the exchange rate list and perform vallback function
func (k Keeper) IterateBaseExchangeRates(ctx sdk.Context, handler func(denom string, exchangeRate types.OracleExchangeRate) bool) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ExchangeRateKey)
	defer iter.Close()

	// Iterate the whole exchangeRate list
	for ; iter.Valid(); iter.Next() {
		// Get denom and rate
		denom := string(iter.Key()[len(types.ExchangeRateKey)])
		rate := types.OracleExchangeRate{}
		k.cdc.MustUnmarshal(iter.Value(), &rate)

		if handler(denom, rate) {
			break
		}
	}
}

// ****************************************************************************

// **************************** Oracle Delegation Logic ***********************
// GetFeederDelegation returns the delegated address by validator address
func (k Keeper) GetFeederDelegation(ctx sdk.Context, valAddr sdk.ValAddress) sdk.AccAddress {
	// Get delegator by validator Address
	store := ctx.KVStore(k.storeKey)
	byteData := store.Get(types.GetFeederDelegationKey(valAddr))
	if byteData == nil {
		// There is no delegater account, return the validator address
		return sdk.AccAddress(valAddr)
	}
	return sdk.AccAddress(byteData)
}

// SetFeederDelegation set into the KVStore the address of an account delegated by the validator
func (k Keeper) SetFeederDelegation(ctx sdk.Context, valAddr sdk.ValAddress, delegatedFeeder sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetFeederDelegationKey(valAddr), delegatedFeeder.Bytes())
}

// IterateFeederDelegations iterate over the delegated list and perform vallback function
func (k Keeper) IterateFeederDelegations(ctx sdk.Context, handler func(valAddr sdk.ValAddress, delegatedFeeder sdk.AccAddress) bool) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.FeederDelegationKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		valAddr := sdk.ValAddress(iter.Key()[2:])
		delegatedFeeder := sdk.AccAddress(iter.Value())

		// when handler returns true stop
		if handler(valAddr, delegatedFeeder) {
			break
		}
	}
}

// ****************************************************************************

// **************************** Miss counter logic ****************************
// GetVotePenaltyCounter returns the vote penalty counter data for an operator (validator or delegated address)
func (k Keeper) GetVotePenaltyCounter(ctx sdk.Context, operator sdk.ValAddress) types.VotePenaltyCounter {
	store := ctx.KVStore(k.storeKey) // Get oracle module's store
	byteData := store.Get(types.GetVotePenaltyCounterKey(operator))
	if byteData == nil {
		return types.VotePenaltyCounter{}
	}

	// Decode information
	votePenaltyCounter := types.VotePenaltyCounter{}
	k.cdc.MustUnmarshal(byteData, &votePenaltyCounter)
	return votePenaltyCounter
}

// SetVotePenaltyCounter add a penalty counter struct associated to an operator (validator or delegated address)
func (k Keeper) SetVotePenaltyCounter(ctx sdk.Context, operator sdk.ValAddress, missCount, abstainCount, successCount uint64) {
	// TODO: Add metrics on defer functions

	votePenaltyCounter := types.VotePenaltyCounter{
		MissCount:    missCount,
		AbstainCount: abstainCount,
		SuccessCount: successCount,
	}

	// Store info
	store := ctx.KVStore(k.storeKey)
	byteData := k.cdc.MustMarshal(&votePenaltyCounter)
	store.Set(types.GetVotePenaltyCounterKey(operator), byteData)
}

// IncrementMissCount increments the missing count to an specific operator address in the KVStore
func (k Keeper) IncrementMissCount(ctx sdk.Context, operator sdk.ValAddress) {
	currentPenaltyCounter := k.GetVotePenaltyCounter(ctx, operator)
	k.SetVotePenaltyCounter(ctx, operator, currentPenaltyCounter.MissCount+1, currentPenaltyCounter.AbstainCount, currentPenaltyCounter.SuccessCount)
}

// IncrementAbstainCount increments the abstain count to an specific operator address in the KVStore
func (k Keeper) IncrementAbstainCount(ctx sdk.Context, operator sdk.ValAddress) {
	currentPenaltyCounter := k.GetVotePenaltyCounter(ctx, operator)
	k.SetVotePenaltyCounter(ctx, operator, currentPenaltyCounter.MissCount, currentPenaltyCounter.AbstainCount+1, currentPenaltyCounter.SuccessCount)
}

// IncrementSuccessCount increments the success count to an specific operator address in the KVStore
func (k Keeper) IncrementSuccessCount(ctx sdk.Context, operator sdk.ValAddress) {
	currentPenaltyCounter := k.GetVotePenaltyCounter(ctx, operator)
	k.SetVotePenaltyCounter(ctx, operator, currentPenaltyCounter.MissCount, currentPenaltyCounter.AbstainCount, currentPenaltyCounter.SuccessCount+1)
}

// GetMissCount increments the missing count to an specific operator address in the KVStore
func (k Keeper) GetMissCount(ctx sdk.Context, operator sdk.ValAddress) uint64 {
	currentPenaltyCounter := k.GetVotePenaltyCounter(ctx, operator)
	return currentPenaltyCounter.MissCount
}

// GetAbstainCount increments the missing count to an specific operator address in the KVStore
func (k Keeper) GetAbstainCount(ctx sdk.Context, operator sdk.ValAddress) uint64 {
	currentPenaltyCounter := k.GetVotePenaltyCounter(ctx, operator)
	return currentPenaltyCounter.AbstainCount
}

// GetSuccessCount increments the missing count to an specific operator address in the KVStore
func (k Keeper) GetSuccessCount(ctx sdk.Context, operator sdk.ValAddress) uint64 {
	currentPenaltyCounter := k.GetVotePenaltyCounter(ctx, operator)
	return currentPenaltyCounter.SuccessCount
}

// DeleteVotePanltyCounter deletes the operator's vote penalty counter element
func (k Keeper) DeleteVotePanltyCounter(ctx sdk.Context, operator sdk.ValAddress) {
	// TODO: Add metrics on defer functions

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetVotePenaltyCounterKey(operator))
}

// IterateVotePenaltyCounters iterates over penalty counters in the store and perform vallback function
func (k Keeper) IterateVotePenaltyCounters(ctx sdk.Context, handler func(operator sdk.ValAddress, votePenaltyCounter types.VotePenaltyCounter) bool) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.VotePenaltyCounterKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		operator := sdk.ValAddress(iter.Key()[2:])

		votePenaltyCounter := types.VotePenaltyCounter{}
		k.cdc.MustUnmarshal(iter.Value(), &votePenaltyCounter)

		if handler(operator, votePenaltyCounter) {
			break
		}
	}
}

// ****************************************************************************

// **************************** Aggregate Exchange Rate Vote logic ************
// GetAggregateExchangeRateVote returns the exchange rate voted from the store by an specific voter
func (k Keeper) GetAggregateExchangeRateVote(ctx sdk.Context, voter sdk.ValAddress) (types.AggregateExchangeRateVote, error) {
	store := ctx.KVStore(k.storeKey)
	byteData := store.Get(types.GetAggregateExchangeRateVoteKey(voter))
	if byteData == nil {
		err := sdkerrors.Wrap(types.ErrNoAggregateVote, voter.String())
		return types.AggregateExchangeRateVote{}, err // Return custom error
	}

	// Decode information
	aggregateVote := types.AggregateExchangeRateVote{}
	k.cdc.MustUnmarshal(byteData, &aggregateVote)
	return aggregateVote, nil
}

// SetAggregateExchangeRateVote adds an oracle exchange rate vote to the KVStore
func (k Keeper) SetAggregateExchangeRateVote(ctx sdk.Context, voter sdk.ValAddress, vote types.AggregateExchangeRateVote) {
	store := ctx.KVStore(k.storeKey)
	byteData := k.cdc.MustMarshal(&vote)
	store.Set(types.GetAggregateExchangeRateVoteKey(voter), byteData)
}

// DeleteAggregateExchangeRateVote deletes an oracle exchange rate vote from the store
func (k Keeper) DeleteAggregateExchangeRateVote(ctx sdk.Context, voter sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetAggregateExchangeRateVoteKey(voter))
}

// IterateAggregateExchangeRateVotes iterates over exchange rate votes in the store and perform vallback function
func (k Keeper) IterateAggregateExchangeRateVotes(ctx sdk.Context, handler func(voterAddr sdk.ValAddress, aggregateVote types.AggregateExchangeRateVote) bool) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.AggregateExchangeRateVoteKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		voterAddr := sdk.ValAddress(iter.Key()[2:])

		aggregateVote := types.AggregateExchangeRateVote{}
		k.cdc.MustUnmarshal(iter.Value(), &aggregateVote)
		if handler(voterAddr, aggregateVote) {
			break
		}
	}
}

// ValidateFeeder the feeder address whether is a validator or delegated address and if is allowed
// to feed the Oracle module price
func (k Keeper) ValidateFeeder(ctx sdk.Context, feederAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	// validate if the feeder addr is a delegated address, if so, validate if the registered bounded address
	// by that validator is the feeder address
	if !feederAddr.Equals(valAddr) {
		delegator := k.GetFeederDelegation(ctx, valAddr) // Get the delegated address by validator address
		if !delegator.Equals(feederAddr) {
			return sdkerrors.Wrap(types.ErrNoVotingPermission, feederAddr.String())
		}
	}

	// Validate the feeder addr is a validator, if so, validate if is bonded (allowed to validate blocks)
	validator := k.StakingKeeper.Validator(ctx, valAddr)
	if valAddr == nil || !validator.IsBonded() {
		return sdkerrors.Wrapf(stakingtypes.ErrNoValidatorFound, "validator %s is not active set", valAddr.String())
	}

	return nil
}

// ****************************************************************************

// **************************** Vote Target logic *****************************
func (k Keeper) GetVoteTarget(ctx sdk.Context, denom string) (types.Denom, error) {
	store := ctx.KVStore(k.storeKey)
	byteData := store.Get(types.GetVoteTargetKey(denom))
	if byteData == nil {
		err := sdkerrors.Wrap(types.ErrNoVoteTarget, denom)
		return types.Denom{}, err // Return custom error
	}

	voteTarget := types.Denom{}
	k.cdc.MustUnmarshal(byteData, &voteTarget)

	return voteTarget, nil
}

// SetVoteTarget adds an denom exchange rate to the KVStore
func (k Keeper) SetVoteTarget(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	byteData := k.cdc.MustMarshal(&types.Denom{Name: denom})
	store.Set(types.GetVoteTargetKey(denom), byteData)
}

// IterateVoteTargets iterates over denoms in the store and perform vallback function
func (k Keeper) IterateVoteTargets(ctx sdk.Context, handler func(denom string, denomInfo types.Denom) bool) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.VoteTargetKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		denom := types.ExtractDenomFromVoteTargetKey(iter.Key()) // Get the specific rate in string

		denomInfo := types.Denom{}
		k.cdc.MustUnmarshal(iter.Value(), &denomInfo)
		if handler(denom, denomInfo) {
			break
		}
	}
}

// DeleteVoteTargets deletes all elements on VoteTargetKey prefix
func (k Keeper) DeleteVoteTargets(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	for _, key := range k.getAllKeysForPrefix(store, types.VoteTargetKey) {
		store.Delete(key)
	}
}

// getAllKeysForPrefix helper function, returns an array with the elements inside a prefix
func (k Keeper) getAllKeysForPrefix(store sdk.KVStore, prefix []byte) [][]byte {
	keys := [][]byte{}
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key()) // Add the key of the elements inside the prefix
	}
	return keys
}

// ****************************************************************************

// **************************** Price Snapshot logic **************************
// GetPriceSnapshot returns the exchange rate prices stored by a defined timestamp
func (k Keeper) GetPriceSnapshot(ctx sdk.Context, timestamp int64) types.PriceSnapshot {
	store := ctx.KVStore(k.storeKey)
	snapshotBytes := store.Get(types.GetPriceSnapshotKey(uint64(timestamp)))
	if snapshotBytes == nil {
		return types.PriceSnapshot{} // Empty response
	}

	// Decode information
	priceSnapshot := types.PriceSnapshot{}
	k.cdc.MustUnmarshal(snapshotBytes, &priceSnapshot)
	return priceSnapshot
}

// SetPriceSnapshot stores the snapshot on the KVStore DO NOT USE IT - USE
func (k Keeper) SetPriceSnapshot(ctx sdk.Context, snapshot types.PriceSnapshot) {
	store := ctx.KVStore(k.storeKey)
	byteData := k.cdc.MustMarshal(&snapshot)
	store.Set(types.GetPriceSnapshotKey(uint64(snapshot.SnapshotTimestamp)), byteData)
}

// AddPriceSnapshot stores the snapshot on the KVStore and deletes snapshots older than the lookBackDuration
// defined on the params
func (k Keeper) AddPriceSnapshot(ctx sdk.Context, snapshot types.PriceSnapshot) {
	params := k.GetParams(ctx) // Get the module Params
	lookBackDuration := params.LookbackDuration

	// Add snapshot on the KVStore
	k.SetPriceSnapshot(ctx, snapshot)

	// Delete the snapshot that it's timestamps is older that the LookbackDuration
	var timestampsToDelete []int64

	k.IteratePriceSnapshots(ctx, func(snapshot types.PriceSnapshot) bool {
		// If the snapshot is too old, mark it for deletion
		if snapshot.SnapshotTimestamp+lookBackDuration < ctx.BlockTime().Unix() {
			timestampsToDelete = append(timestampsToDelete, snapshot.SnapshotTimestamp)
			return false // Continue iteration
		}

		// If a valid snapshot is found, stop iterating
		return true
	})

	// Delete all marked old snapshots
	for _, timeToDelete := range timestampsToDelete {
		k.DeletePriceSnapshot(ctx, timeToDelete)
	}
}

// IteratePriceSnapshots iterates over the snapshot list and execute the handler
func (k Keeper) IteratePriceSnapshots(ctx sdk.Context, handler func(snapshot types.PriceSnapshot) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.PriceSnapshotKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.PriceSnapshot
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if handler(val) {
			break
		}
	}
}

// IteratePriceSnapshotsReverse REVERSE iterates over the snapshot list and execute the handler
func (k Keeper) IteratePriceSnapshotsReverse(ctx sdk.Context, handler func(snapshot types.PriceSnapshot) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStoreReversePrefixIterator(store, types.PriceSnapshotKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.PriceSnapshot
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if handler(val) {
			break
		}
	}
}

// DeletePriceSnapshot deletes an snapshot based by the given timestamp
func (k Keeper) DeletePriceSnapshot(ctx sdk.Context, timestamp int64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPriceSnapshotKey(uint64(timestamp)))
}

// ****************************************************************************

// **************************** Spam Prevention Counter logic *****************
// GetSpamPreventionCounter returns the stored block heigh by the validator (in that heigh the validator voted)
func (k Keeper) GetSpamPreventionCounter(ctx sdk.Context, valAddr sdk.ValAddress) int64 {
	store := ctx.KVStore(k.memKey) // Get oracle module's KVStore
	byteData := store.Get(types.GetSpamPreventionCounterKey(valAddr))
	if byteData == nil {
		return -1 // Return invalid counter
	}

	return int64(sdk.BigEndianToUint64(byteData)) // return the counter by validator address
}

// SetSpamPreventionCounter stores the block heigh by the validator as an anti voting spam mecanism
func (k Keeper) SetSpamPreventionCounter(ctx sdk.Context, valAddr sdk.ValAddress) {
	store := ctx.KVStore(k.memKey)

	height := ctx.BlockHeight() // Get current block height
	byteData := sdk.Uint64ToBigEndian(uint64(height))

	store.Set(types.GetSpamPreventionCounterKey(valAddr), byteData) // store the current block height
}
