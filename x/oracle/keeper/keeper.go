package keeper

import (
	"errors"
	"fmt"
	"sort"

	"cosmossdk.io/collections"
	corestoretypes "cosmossdk.io/core/store"
	cosmoserrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kiichain/kiichain/v1/x/oracle/types"
)

// Keeper manages the oracle module's state
type Keeper struct {
	cdc codec.BinaryCodec // Codec for binary serialization

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	StakingKeeper types.StakingKeeper

	// Schema of the module
	Schema                    collections.Schema
	Params                    collections.Item[types.Params]
	ExchangeRate              collections.Map[string, types.OracleExchangeRate]
	FeederDelegation          collections.Map[sdk.ValAddress, string]
	VotePenaltyCounter        collections.Map[sdk.ValAddress, types.VotePenaltyCounter]
	AggregateExchangeRateVote collections.Map[sdk.ValAddress, types.AggregateExchangeRateVote]
	VoteTarget                collections.Map[string, types.Denom]
	PriceSnapshot             collections.Map[int64, types.PriceSnapshot]
	SpamPreventionCounter     collections.Map[sdk.ValAddress, int64]

	// Authority is the governance module address
	authority string
}

// NewKeeper creates an oracle Keeper instance
func NewKeeper(cdc codec.BinaryCodec, storeService corestoretypes.KVStoreService,
	accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, stakingKeeper types.StakingKeeper,
	authority string,
) Keeper {
	// Ensure oracle module account is set
	addr := accountKeeper.GetModuleAddress(types.ModuleName)
	if addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// Ensure the authority address is valid
	if _, err := accountKeeper.AddressCodec().StringToBytes(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	// Start the schema builder
	sb := collections.NewSchemaBuilder(storeService)

	// Build the Keeper
	keeper := Keeper{
		cdc:                       cdc,
		accountKeeper:             accountKeeper,
		bankKeeper:                bankKeeper,
		StakingKeeper:             stakingKeeper,
		Params:                    collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		ExchangeRate:              collections.NewMap(sb, types.ExchangeRateKey, "exchange_rate", collections.StringKey, codec.CollValue[types.OracleExchangeRate](cdc)),
		FeederDelegation:          collections.NewMap(sb, types.FeederDelegationKey, "feeder_delegation", sdk.ValAddressKey, collections.StringValue),
		VotePenaltyCounter:        collections.NewMap(sb, types.VotePenaltyCounterKey, "vote_penalty_counter", sdk.ValAddressKey, codec.CollValue[types.VotePenaltyCounter](cdc)),
		AggregateExchangeRateVote: collections.NewMap(sb, types.AggregateExchangeRateVoteKey, "aggregate_exchange_rate_vote", sdk.ValAddressKey, codec.CollValue[types.AggregateExchangeRateVote](cdc)),
		VoteTarget:                collections.NewMap(sb, types.VoteTargetKey, "vote_target", collections.StringKey, codec.CollValue[types.Denom](cdc)),
		PriceSnapshot:             collections.NewMap(sb, types.PriceSnapshotKey, "price_snapshot", collections.Int64Key, codec.CollValue[types.PriceSnapshot](cdc)),
		SpamPreventionCounter:     collections.NewMap(sb, types.SpamPreventionCounter, "spam_prevention_counter", sdk.ValAddressKey, collections.Int64Value),

		authority: authority,
	}

	// Build the schema
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	// Set and return
	keeper.Schema = schema
	return keeper
}

// Logger is used to define custom Log for the module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// **************************** EXCHANGE RATE LOGIC ***************************

// GetBaseExchangeRate is used to get the exchange rate by denom from the KVStore
func (k Keeper) GetBaseExchangeRate(ctx sdk.Context, denom string) (types.OracleExchangeRate, error) {
	return k.ExchangeRate.Get(ctx, denom)
}

// SetBaseExchangeRate is used to set the exchange rate by denom on the KVStore
func (k Keeper) SetBaseExchangeRate(ctx sdk.Context, denom string, exchangeRate math.LegacyDec) {
	// Get the extra data
	currentHeight := math.NewInt(ctx.BlockHeight())
	blockTimestamp := ctx.BlockTime().UnixMilli()

	// Build the exchange rate object
	rate := types.OracleExchangeRate{
		ExchangeRate:        exchangeRate,
		LastUpdate:          currentHeight,
		LastUpdateTimestamp: blockTimestamp,
	}

	// Store the exchange rate
	k.ExchangeRate.Set(ctx, denom, rate)
}

// SetBaseExchangeRateWithEvent calls SetBaseExchangeRate and generate an event about that denom creation
func (k Keeper) SetBaseExchangeRateWithEvent(ctx sdk.Context, denom string, exchangeRate math.LegacyDec) {
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
	k.ExchangeRate.Remove(ctx, denom)
}

// IterateBaseExchangeRates iterate over the exchange rate list and perform vallback function
func (k Keeper) IterateBaseExchangeRates(ctx sdk.Context, handler func(denom string, exchangeRate types.OracleExchangeRate) (bool, error)) {
	k.ExchangeRate.Walk(ctx, nil, handler)
}

// ****************************************************************************

// **************************** Oracle Delegation Logic ***********************

// GetFeederDelegation returns the delegated address by validator address
func (k Keeper) GetFeederDelegation(ctx sdk.Context, valAddr sdk.ValAddress) sdk.AccAddress {
	// Get the account address
	accAddressString, err := k.FeederDelegation.Get(ctx, valAddr)
	// If the not found, return the val Address
	if errors.Is(err, collections.ErrNotFound) {
		return sdk.AccAddress(valAddr)
	}

	// Handle any other error
	if err != nil {
		panic(err) // FIX ME: Add proper error handling
	}

	// Marshal the address bytes to sdk.AccAddress
	accAddress, err := sdk.AccAddressFromBech32(accAddressString)
	if err != nil {
		panic(err) // FIX ME: Add proper error handling
	}

	// Return the account address
	return accAddress
}

// SetFeederDelegation set into the KVStore the address of an account delegated by the validator
func (k Keeper) SetFeederDelegation(ctx sdk.Context, valAddr sdk.ValAddress, delegatedFeeder sdk.AccAddress) {
	// Set the feeder delegation
	k.FeederDelegation.Set(ctx, valAddr, delegatedFeeder.String())
}

// IterateFeederDelegations iterate over the delegated list and perform vallback function
func (k Keeper) IterateFeederDelegations(ctx sdk.Context, handler func(valAddr sdk.ValAddress, delegatedFeeder string) (bool, error)) {
	// Iterate the FeederDelegation
	k.FeederDelegation.Walk(ctx, nil, handler)
}

// ValidateFeeder the feeder address whether is a validator or delegated address and if is allowed
// to feed the Oracle module price
func (k Keeper) ValidateFeeder(ctx sdk.Context, feederAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	// validate if the feeder addr is a delegated address, if so, validate if the registered bounded address
	// by that validator is the feeder address
	if !feederAddr.Equals(valAddr) {
		delegator := k.GetFeederDelegation(ctx, valAddr) // Get the delegated address by validator address
		if !delegator.Equals(feederAddr) {
			return cosmoserrors.Wrap(types.ErrNoVotingPermission, feederAddr.String())
		}
	}

	// Validate the feeder addr is a validator, if so, validate if is bonded (allowed to validate blocks)
	validator, err := k.StakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return cosmoserrors.Wrapf(stakingtypes.ErrNoValidatorFound, "validator %s not found", valAddr.String())
	}
	if valAddr == nil || !validator.IsBonded() {
		return cosmoserrors.Wrapf(stakingtypes.ErrNoValidatorFound, "validator %s is not active set", valAddr.String())
	}

	return nil
}

// ****************************************************************************

// **************************** Miss counter logic ****************************

// GetVotePenaltyCounter returns the vote penalty counter data for an operator (validator or delegated address)
func (k Keeper) GetVotePenaltyCounter(ctx sdk.Context, operator sdk.ValAddress) types.VotePenaltyCounter {
	votePenaltyCounter, err := k.VotePenaltyCounter.Get(ctx, operator)
	// If not registered yet, return a default value
	if errors.Is(err, collections.ErrNotFound) {
		return types.VotePenaltyCounter{}
	}

	// Handle any other error
	if err != nil {
		panic(err) // FIX ME: Add proper error handling
	}
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

	// Set the penalty counter
	k.VotePenaltyCounter.Set(ctx, operator, votePenaltyCounter)
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
	k.VotePenaltyCounter.Remove(ctx, operator)
}

// IterateVotePenaltyCounters iterates over penalty counters in the store and perform vallback function
func (k Keeper) IterateVotePenaltyCounters(ctx sdk.Context, handler func(operator sdk.ValAddress, votePenaltyCounter types.VotePenaltyCounter) (bool, error)) {
	// Walk through the VotePenaltyCounter map
	k.VotePenaltyCounter.Walk(ctx, nil, handler)
}

// ****************************************************************************

// **************************** Aggregate Exchange Rate Vote logic ************

// GetAggregateExchangeRateVote returns the exchange rate voted from the store by an specific voter
func (k Keeper) GetAggregateExchangeRateVote(ctx sdk.Context, voter sdk.ValAddress) (types.AggregateExchangeRateVote, error) {
	return k.AggregateExchangeRateVote.Get(ctx, voter)
}

// SetAggregateExchangeRateVote adds an oracle exchange rate vote to the KVStore
func (k Keeper) SetAggregateExchangeRateVote(ctx sdk.Context, voter sdk.ValAddress, vote types.AggregateExchangeRateVote) {
	k.AggregateExchangeRateVote.Set(ctx, voter, vote)
}

// DeleteAggregateExchangeRateVote deletes an oracle exchange rate vote from the store
func (k Keeper) DeleteAggregateExchangeRateVote(ctx sdk.Context, voter sdk.ValAddress) {
	k.AggregateExchangeRateVote.Remove(ctx, voter)
}

// IterateAggregateExchangeRateVotes iterates over exchange rate votes in the store and perform vallback function
func (k Keeper) IterateAggregateExchangeRateVotes(ctx sdk.Context, handler func(voterAddr sdk.ValAddress, aggregateVote types.AggregateExchangeRateVote) (bool, error)) {
	// Iterate the AggregateExchangeRateVote map
	k.AggregateExchangeRateVote.Walk(ctx, nil, handler)
}

// RemoveExcessFeeds deletes the exchange rates added to the KVStore but not require on the whitelist
func (k Keeper) RemoveExcessFeeds(ctx sdk.Context) {
	// get exchange rates stored on the KVStore
	excessActives := make(map[string]struct{})
	k.IterateBaseExchangeRates(ctx, func(denom string, exchangeRate types.OracleExchangeRate) (bool, error) {
		excessActives[denom] = struct{}{}
		return false, nil
	})

	// Get voting target
	k.VoteTarget.Walk(ctx, nil, func(denom string, denomInfo types.Denom) (bool, error) {
		// Remove vote targets from actives
		delete(excessActives, denom)
		return false, nil
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

// **************************** Price Snapshot logic **************************

// GetPriceSnapshot returns the exchange rate prices stored by a defined timestamp
func (k Keeper) GetPriceSnapshot(ctx sdk.Context, timestamp int64) types.PriceSnapshot {
	// Get the price snapshot
	priceSnapshot, err := k.PriceSnapshot.Get(ctx, timestamp)

	// If not found, return an empty snapshot
	if errors.Is(err, collections.ErrNotFound) {
		return types.PriceSnapshot{}
	}

	// Handle any other error
	if err != nil {
		panic(err) // FIX ME: Add proper error handling
	}

	return priceSnapshot
}

// AddPriceSnapshot stores the snapshot on the KVStore and deletes snapshots older than the lookBackDuration
// defined on the params
func (k Keeper) AddPriceSnapshot(ctx sdk.Context, snapshot types.PriceSnapshot) error {
	// Get the module params
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	lookBackDuration := params.LookbackDuration

	// Add snapshot on the KVStore
	err = k.PriceSnapshot.Set(ctx, snapshot.SnapshotTimestamp, snapshot)
	if err != nil {
		return err
	}

	// Delete the snapshot that it's timestamps is older that the LookbackDuration
	var timestampsToDelete []int64

	k.IteratePriceSnapshots(ctx, func(_ int64, snapshot types.PriceSnapshot) (bool, error) {
		// If the snapshot is too old, mark it for deletion
		if snapshot.SnapshotTimestamp+int64(lookBackDuration) < ctx.BlockTime().Unix() {
			timestampsToDelete = append(timestampsToDelete, snapshot.SnapshotTimestamp)
			return false, nil // Continue iteration
		}

		// If a valid snapshot is found, stop iterating
		return true, nil
	})

	// Delete all marked old snapshots
	for _, timeToDelete := range timestampsToDelete {
		err = k.PriceSnapshot.Remove(ctx, timeToDelete)
		if err != nil {
			return err
		}
	}
	return nil
}

// IteratePriceSnapshots iterates over the snapshot list and execute the handler
func (k Keeper) IteratePriceSnapshots(ctx sdk.Context, handler func(timestamp int64, snapshot types.PriceSnapshot) (bool, error)) {
	// Iterate the PriceSnapshot map
	k.PriceSnapshot.Walk(ctx, nil, handler)
}

// IteratePriceSnapshotsReverse REVERSE iterates over the snapshot list and execute the handler
func (k Keeper) IteratePriceSnapshotsReverse(ctx sdk.Context, handler func(snapshot types.PriceSnapshot) (bool, error)) error {
	// Iterate the PriceSnapshot map in reverse order
	iterator, err := k.PriceSnapshot.IterateRaw(ctx, nil, nil, collections.OrderDescending)
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Take the value
		val, err := iterator.Value()
		if err != nil {
			return err
		}

		// Handle the value
		stop, err := handler(val)
		if err != nil {
			return err
		}

		// If stop, stop
		if stop {
			break
		}
	}

	return nil
}

// ****************************************************************************

// **************************** Spam Prevention Counter logic *****************

// SetSpamPreventionCounter stores the block heigh by the validator as an anti voting spam mechanism
func (k Keeper) SetSpamPreventionCounter(ctx sdk.Context, valAddr sdk.ValAddress) error {
	// Get the height of the current block
	height := ctx.BlockHeight()

	// Set the spam prevention counter
	return k.SpamPreventionCounter.Set(ctx, valAddr, height)
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

	var timeTraversed int64                        // last time analyzed
	twapByDenom := make(map[string]math.LegacyDec) // Here I will store the calculated twap by denom
	durationByDenom := make(map[string]int64)      // Here I will the analyzed duration by denom

	// get targets exchange rate
	targetsMap := make(map[string]struct{}) // here I store the collected targets from the KVStore
	err = k.VoteTarget.Walk(ctx, nil, func(denom string, denomInfo types.Denom) (bool, error) {
		targetsMap[denom] = struct{}{} // Store the active targets
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	// Iterate the complete snapshots list from the most recent to the oldest
	err = k.IteratePriceSnapshotsReverse(ctx, func(snapshot types.PriceSnapshot) (stop bool, err error) {
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
				twapByDenom[denom] = math.LegacyZeroDec()
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
		return stop, err
	})
	if err != nil {
		return nil, err
	}

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
		denomTwap := math.LegacyZeroDec()
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
	// Get the params
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Check the lookback seconds against the params
	lookBackDuration := params.LookbackDuration
	if lookBackSeconds > lookBackDuration || lookBackSeconds == 0 {
		return types.ErrInvalidTwapLookback
	}
	return nil
}
