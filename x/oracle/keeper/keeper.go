package keeper

import (
	"fmt"
	"sort"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
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
	if addr == nil {
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

// GetBaseExchangeRate is used to get the exchange rate by denom from the KVStore
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

// SetBaseExchangeRate is used to set the exchange rate by denom on the KVStore
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
		denom := string(iter.Key()[len(types.ExchangeRateKey):])
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
func (k Keeper) DeleteVotePenaltyCounter(ctx sdk.Context, operator sdk.ValAddress) {
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

// RemoveExcessFeeds deletes the exchange rates added to the KVStore but not require on the whitelist
func (k Keeper) RemoveExcessFeeds(ctx sdk.Context) {
	// get exchange rates stored on the KVStore
	excessActives := make(map[string]struct{})
	k.IterateBaseExchangeRates(ctx, func(denom string, exchangeRate types.OracleExchangeRate) bool {
		excessActives[denom] = struct{}{}
		return false
	})

	// Get voting target
	k.IterateVoteTargets(ctx, func(denom string, denomInfo types.Denom) bool {
		// Remove vote targets from actives
		delete(excessActives, denom)
		return false
	})

	// at this point just left the excess exchange rates
	activesToClear := make([]string, 0, len(excessActives))
	for denom := range excessActives {
		activesToClear = append(activesToClear, denom)
	}
	sort.Strings(activesToClear)

	// delete the excess exchange rates
	for _, denom := range activesToClear {
		k.DeleteBaseExchangeRate(ctx, denom)
	}
}

// ****************************************************************************

// **************************** Vote Target logic *****************************

// GetVoteTarget returns the Denom info on the KVStore by denom string
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

// ClearVoteTargets deletes all voting target on the KVStore
func (k Keeper) ClearVoteTargets(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	for _, key := range k.getAllKeysForPrefix(store, types.VoteTargetKey) {
		store.Delete(key)
	}
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
		if snapshot.SnapshotTimestamp+int64(lookBackDuration) < ctx.BlockTime().Unix() {
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

// ****************************************************************************

// **************************** Helper Functions logic ************************

// CalculateTwaps calculate the twap to each exchange rate stored on the KVStore, the twap is a fundamental operation
// to avoid price manipulation using the historycal price and feeders input to calculate the current price
func (k Keeper) CalculateTwaps(ctx sdk.Context, lookBackSeconds uint64) (types.OracleTwaps, error) {
	oracleTwaps := types.OracleTwaps{}
	currentTime := ctx.BlockTime().Unix()                  // timestamp time unit
	err := k.ValidateLookBackSeconds(ctx, lookBackSeconds) // validate the input lookback
	if err != nil {
		return oracleTwaps, err
	}

	var timeTraversed int64                   // last time analyzed
	twapByDenom := make(map[string]sdk.Dec)   // Here I will store the calculated twap by denom
	durationByDenom := make(map[string]int64) // Here I will the analyzed duration by denom

	// get targets exchange rate
	targetsMap := make(map[string]struct{}) // here I store the collected targets from the KVStore
	k.IterateVoteTargets(ctx, func(denom string, denomInfo types.Denom) bool {
		targetsMap[denom] = struct{}{} // Store the active targets
		return false
	})

	// Iterate the complete snapshots list from the most recent to the oldest
	k.IteratePriceSnapshotsReverse(ctx, func(snapshot types.PriceSnapshot) (stop bool) {
		stop = false

		// Check if the current snapshot is older than the lookBack time
		// currentTime - lookBackSeconds is the end time until I will calculate the twap
		snapshotTimestamp := snapshot.SnapshotTimestamp
		if currentTime-int64(lookBackSeconds) > snapshotTimestamp { // If this happened, means the snapshot is older than the lookback period
			snapshotTimestamp = currentTime - int64(lookBackSeconds)
			stop = true // Stop iteration
		}

		timeTraversed = currentTime - snapshotTimestamp // time between current block and the analized snapshot

		snapshotPriceItems := snapshot.PriceSnapshotItems // Get the current snapshot data (an array of denom with its exchange rate)
		for _, priceItem := range snapshotPriceItems {    // Iterate the aray of data
			// Get snapshot denom and check if its valid (is a target denom)
			denom := priceItem.Denom
			_, ok := targetsMap[denom]
			if !ok {
				continue // The denom that is not tergeted does not care
			}

			// Check if the twap by denom exist, if so initialize the average with 0
			_, exist := twapByDenom[denom]
			if !exist {
				twapByDenom[denom] = sdk.ZeroDec()
				durationByDenom[denom] = 0
			}

			// Calculate the twap by denom
			twapAverageByDenom := twapByDenom[denom] // current twap by denom
			denomDuration := durationByDenom[denom]  // current analyzed time by denom

			durationDifference := timeTraversed - denomDuration                                    // difference between current time and the
			exchangeRate := priceItem.OracleExchangeRate.ExchangeRate                              // exchange rate on the snapshot
			twapAverageByDenom = twapAverageByDenom.Add(exchangeRate.MulInt64(durationDifference)) // multiply the snapshot by the duration

			twapByDenom[denom] = twapAverageByDenom // update the twap by denom with the result
			durationByDenom[denom] = timeTraversed  // update the analized time by denom
		}
		return stop
	})

	// Order the exchange rates with its twaps (just to have an order)
	denomKeys := make([]string, 0, len(twapByDenom))
	for k := range twapByDenom {
		denomKeys = append(denomKeys, k)
	}
	sort.Strings(denomKeys)

	// iterate over all denoms with TWAP data
	for _, denomKey := range denomKeys {
		// divide the twap sum by denom duration
		denomTimeWeightedSum := twapByDenom[denomKey] // Get twap
		denomDuration := durationByDenom[denomKey]    // Get duration

		// validate divide by zero
		denomTwap := sdk.ZeroDec()
		if denomDuration != 0 {
			denomTwap = denomTimeWeightedSum.QuoInt64(denomDuration)
		}

		denomOracleTwap := types.OracleTwap{
			Denom:           denomKey,
			Twap:            denomTwap,
			LookbackSeconds: denomDuration,
		}

		// Store on the calculated twaps list
		oracleTwaps = append(oracleTwaps, denomOracleTwap)
	}

	if len(oracleTwaps) == 0 {
		return oracleTwaps, types.ErrNoTwapData
	}

	return oracleTwaps, nil
}

// ValidateLookBackSeconds validates the input lookbackseconds, must be lower or equan than the param lookback (because there are not longer
// data than the param lookback param)
func (k Keeper) ValidateLookBackSeconds(ctx sdk.Context, lookBackSeconds uint64) error {
	lookBackDuration := k.LookbackDuration(ctx)
	if lookBackSeconds > lookBackDuration || lookBackSeconds == 0 {
		return types.ErrInvalidTwapLookback
	}
	return nil
}
