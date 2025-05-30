package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
)

func TestExchangeRateLogic(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// Exchange rates to be stored
	const BtcUsd = "BTC/USD"
	const EthUsd = "ETC/USD"
	const AtomUsd = "ATOM/USD"

	btcUsdExchangeRate := math.LegacyNewDecWithPrec(100, int64(OracleDecPrecision)).MulInt64(1e6)
	ethUsdExchangeRate := math.LegacyNewDecWithPrec(200, int64(OracleDecPrecision)).MulInt64(1e6)
	atomUsdExchangeRate := math.LegacyNewDecWithPrec(300, int64(OracleDecPrecision)).MulInt64(1e6)

	// ***** First exchange rate insertion
	err := oracleKeeper.SetBaseExchangeRate(ctx, BtcUsd, btcUsdExchangeRate) // Set exchange rates on KVStore
	require.NoError(t, err)
	btcUsdRate, err := oracleKeeper.GetBaseExchangeRate(ctx, BtcUsd) // Get exchange rate from KVStore
	require.NoError(t, err, "Expected no error getting BTC/USD exchange rate")
	require.Equal(t, btcUsdExchangeRate, btcUsdRate.ExchangeRate, "Expected got the same exchange rate as ")
	require.Equal(t, math.ZeroInt(), btcUsdRate.LastUpdate) // There is no previous updates

	// simulate time pass
	ctx = ctx.WithBlockHeight(3) // Update block height
	ts := time.Now()
	ctx = ctx.WithBlockTime(ts) // Update block timestamp

	// ***** Second exchange rate insertion
	err = oracleKeeper.SetBaseExchangeRate(ctx, EthUsd, ethUsdExchangeRate) // Set exchange rates on KVStore
	require.NoError(t, err)
	ethUsdRate, err := oracleKeeper.GetBaseExchangeRate(ctx, EthUsd) // Get exchange rate from KVStore
	require.NoError(t, err)
	require.Equal(t, ethUsdExchangeRate, ethUsdRate.ExchangeRate)
	require.Equal(t, math.NewInt(3), ethUsdRate.LastUpdate)
	require.Equal(t, ts.UnixMilli(), ethUsdRate.LastUpdateTimestamp)

	// simulate time pass
	ctx = ctx.WithBlockHeight(15) // Update block height
	newTime := ts.Add(time.Hour)
	ctx = ctx.WithBlockTime(newTime) // Update block timestamp

	// ***** Third exchange rate insertion (using events)
	err = oracleKeeper.SetBaseExchangeRateWithEvent(ctx, AtomUsd, atomUsdExchangeRate) // Set exchange rates on KVStore
	require.NoError(t, err)
	atomUsdRate, err := oracleKeeper.GetBaseExchangeRate(ctx, AtomUsd) // Get exchange rate from KVStore

	// Create the event validation function
	eventValidation := func() bool {
		// Expected event
		expectedEvent := sdk.NewEvent(
			types.EventTypeExchangeRateUpdate,
			sdk.NewAttribute(types.AttributeKeyDenom, AtomUsd),
			sdk.NewAttribute(types.AttributeKeyExchangeRate, atomUsdExchangeRate.String()))

		// Read the current events
		events := ctx.EventManager().Events()
		for _, event := range events {
			if event.Type != expectedEvent.Type { // Search the expected event
				continue
			}

			// Iterate over the event
			for i, attr := range event.Attributes {
				if attr.Index != expectedEvent.Attributes[i].Index {
					return false
				}

				if attr.Key != expectedEvent.Attributes[i].Key {
					return false
				}

				if attr.Value != expectedEvent.Attributes[i].Value {
					return false
				}
			}
			return true
		}
		return false
	}

	// Validations
	require.NoError(t, err)
	require.Equal(t, atomUsdExchangeRate, atomUsdRate.ExchangeRate)
	require.Equal(t, math.NewInt(15), atomUsdRate.LastUpdate)
	require.Equal(t, newTime.UnixMilli(), atomUsdRate.LastUpdateTimestamp)
	require.True(t, eventValidation())

	// ***** First exchange rate elimination
	oracleKeeper.DeleteBaseExchangeRate(ctx, BtcUsd)
	_, err = oracleKeeper.GetBaseExchangeRate(ctx, BtcUsd)
	require.Error(t, err) // Validate error

	// test iteration function
	exchangeRateAmount := 0
	iterationHandler := func(denom string, exchangeRate types.OracleExchangeRate) (bool, error) {
		exchangeRateAmount++
		return false, nil
	}

	oracleKeeper.IterateBaseExchangeRates(ctx, iterationHandler)
	require.Equal(t, 2, exchangeRateAmount) // verify that iterate over all exchange rates elements
}

func TestParams(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// test default params
	defaultParams, err := oracleKeeper.Params.Get(ctx)
	require.NoError(t, err)

	err = oracleKeeper.Params.Set(ctx, defaultParams) // Set default params
	require.NoError(t, err)
	require.NotNil(t, defaultParams)

	// test custom params
	votePeriod := uint64(10)
	voteThreshold := math.LegacyNewDecWithPrec(33, 2) // 0.033
	rewardBand := math.LegacyNewDecWithPrec(1, 2)     // 0.01
	slashFraccion := math.LegacyNewDecWithPrec(1, 2)  // 0.01
	slashwindow := uint64(1000)
	minValPerWindow := math.LegacyNewDecWithPrec(1, 4) // 0.0001
	whiteList := types.DenomList{{Name: utils.MicroKiiDenom}, {Name: utils.MicroAtomDenom}}
	lookbackDuration := uint64(3600)

	params := types.Params{
		VotePeriod:        votePeriod,
		VoteThreshold:     voteThreshold,
		RewardBand:        rewardBand,
		Whitelist:         whiteList,
		SlashFraction:     slashFraccion,
		SlashWindow:       slashwindow,
		MinValidPerWindow: minValPerWindow,
		LookbackDuration:  lookbackDuration,
	}
	err = oracleKeeper.Params.Set(ctx, params)
	require.NoError(t, err)

	storedParams, err := oracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, slashFraccion)
	require.Equal(t, params, storedParams)
}

func TestDelegationLogic(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// ***** Get and set feeder delegator
	delegate := oracleKeeper.GetFeederDelegation(ctx, ValAddrs[0]) // supposed to received the same val addr
	require.Equal(t, Addrs[0], delegate)

	oracleKeeper.SetFeederDelegation(ctx, ValAddrs[0], Addrs[1]) // Delegate Val 0 -> Addr 1
	delegate = oracleKeeper.GetFeederDelegation(ctx, ValAddrs[0])
	require.Equal(t, Addrs[1], delegate)

	// ***** Iterate feeder delegator list
	var validators []sdk.ValAddress
	var delegates []sdk.AccAddress
	handler := func(valAddr sdk.ValAddress, delegatedFeeder string) (bool, error) {
		validators = append(validators, valAddr)

		// Parse the delegated feeder address to sdk.AccAddress
		delegatedFeederAcc, err := sdk.AccAddressFromBech32(delegatedFeeder)
		if err != nil {
			return true, err
		}

		delegates = append(delegates, delegatedFeederAcc)
		return false, nil
	}
	oracleKeeper.IterateFeederDelegations(ctx, handler)

	// Validation
	require.Equal(t, 1, len(delegates))
	require.Equal(t, 1, len(validators))
	require.Equal(t, Addrs[1], delegates[0]) // Validator 0 delegate to -> Addr 1
}

func TestValidateFeeder(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	stakingKeeper := init.StakingKeeper
	ctx := init.Ctx
	amount := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction) // staking power tokens
	msgServer := stakingkeeper.NewMsgServerImpl(&stakingKeeper)

	// Create validators
	val1Addr, val1PubKey := ValAddrs[0], ValPubKeys[0]
	val2Addr, val2PubKey := ValAddrs[1], ValPubKeys[1]
	_, err := msgServer.CreateValidator(ctx, NewTestMsgCreateValidator(val1Addr, val1PubKey, amount)) // Create validator
	require.NoError(t, err)
	_, err = msgServer.CreateValidator(ctx, NewTestMsgCreateValidator(val2Addr, val2PubKey, amount)) // Create validator
	require.NoError(t, err)
	_, err = stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	// Validate validator's bonded tokens
	stakingParams, err := stakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	bondDenomDefault := stakingParams.BondDenom
	reference := sdk.NewCoins(sdk.NewCoin(bondDenomDefault, InitTokens.Sub(amount))) // Create balance reference, Suppose to be 100 Kii
	balanceVal1 := init.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(val1Addr))
	balanceVal2 := init.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(val2Addr))

	// Get the validators
	val1, err := stakingKeeper.Validator(ctx, val1Addr)
	require.NoError(t, err)
	val2, err := stakingKeeper.Validator(ctx, val2Addr)
	require.NoError(t, err)

	// Get the bonded tokens for each validator
	bondedVal1 := val1.GetBondedTokens()
	bondedVal2 := val2.GetBondedTokens()

	// Validation
	require.Equal(t, reference, balanceVal1)
	require.Equal(t, reference, balanceVal2)
	require.Equal(t, amount, bondedVal1)
	require.Equal(t, amount, bondedVal2)

	// Validate Feeder when validators did not delegate
	require.NoError(t, oracleKeeper.ValidateFeeder(ctx, sdk.AccAddress(val1Addr), val1Addr))
	require.NoError(t, oracleKeeper.ValidateFeeder(ctx, sdk.AccAddress(val2Addr), val2Addr))

	// Delegate validator 1 to Val 2
	oracleKeeper.SetFeederDelegation(ctx, val1Addr, sdk.AccAddress(val2Addr))                // Delegate Val 1 to Val 2
	require.NoError(t, oracleKeeper.ValidateFeeder(ctx, sdk.AccAddress(val2Addr), val1Addr)) // Validate that Val2 is delegated by val1
	require.Error(t, oracleKeeper.ValidateFeeder(ctx, Addrs[2], val1Addr))
}

func TestMissCounter(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// ***** Get default voting information
	counter := oracleKeeper.GetVotePenaltyCounter(ctx, ValAddrs[0]) // Get the counter details of the val 0

	// Validation (everything must be zero, I haven't add voting information yet)
	require.Equal(t, uint64(0), counter.MissCount)
	require.Equal(t, uint64(0), counter.AbstainCount)
	require.Equal(t, uint64(0), counter.SuccessCount)
	require.Equal(t, uint64(0), oracleKeeper.GetMissCount(ctx, ValAddrs[0]))
	require.Equal(t, uint64(0), oracleKeeper.GetAbstainCount(ctx, ValAddrs[0]))
	require.Equal(t, uint64(0), oracleKeeper.GetSuccessCount(ctx, ValAddrs[0]))

	// ***** Set an specific voting information
	missCounter := uint64(10)
	abstainCounter := uint64(20)
	successCounter := uint64(30)
	oracleKeeper.SetVotePenaltyCounter(ctx, ValAddrs[0], missCounter, abstainCounter, successCounter) // Set the voting info

	// Validation
	counter = oracleKeeper.GetVotePenaltyCounter(ctx, ValAddrs[0]) // Get the counter details of the val 0
	require.Equal(t, missCounter, counter.MissCount)
	require.Equal(t, abstainCounter, counter.AbstainCount)
	require.Equal(t, successCounter, counter.SuccessCount)
	require.Equal(t, missCounter, oracleKeeper.GetMissCount(ctx, ValAddrs[0]))
	require.Equal(t, abstainCounter, oracleKeeper.GetAbstainCount(ctx, ValAddrs[0]))
	require.Equal(t, successCounter, oracleKeeper.GetSuccessCount(ctx, ValAddrs[0]))

	// ***** Test delete voting info
	oracleKeeper.DeleteVotePenaltyCounter(ctx, ValAddrs[0])

	// validation
	counter = oracleKeeper.GetVotePenaltyCounter(ctx, ValAddrs[0]) // Get the counter details of the val 0
	require.Equal(t, uint64(0), counter.MissCount)
	require.Equal(t, uint64(0), counter.AbstainCount)
	require.Equal(t, uint64(0), counter.SuccessCount)
	require.Equal(t, uint64(0), oracleKeeper.GetMissCount(ctx, ValAddrs[0]))
	require.Equal(t, uint64(0), oracleKeeper.GetAbstainCount(ctx, ValAddrs[0]))
	require.Equal(t, uint64(0), oracleKeeper.GetSuccessCount(ctx, ValAddrs[0]))

	// ***** Test increment function
	oracleKeeper.IncrementMissCount(ctx, ValAddrs[0])
	oracleKeeper.IncrementAbstainCount(ctx, ValAddrs[0])
	oracleKeeper.IncrementSuccessCount(ctx, ValAddrs[0])

	// validation
	counter = oracleKeeper.GetVotePenaltyCounter(ctx, ValAddrs[0]) // Get the counter details of the val 0
	require.Equal(t, uint64(1), counter.MissCount)
	require.Equal(t, uint64(1), counter.AbstainCount)
	require.Equal(t, uint64(1), counter.SuccessCount)
	require.Equal(t, uint64(1), oracleKeeper.GetMissCount(ctx, ValAddrs[0]))
	require.Equal(t, uint64(1), oracleKeeper.GetAbstainCount(ctx, ValAddrs[0]))
	require.Equal(t, uint64(1), oracleKeeper.GetSuccessCount(ctx, ValAddrs[0]))
}

func TestMissCounterIterate(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// Set voting info
	missCounter := uint64(10)
	abstainCounter := uint64(20)
	successCounter := uint64(30)
	oracleKeeper.SetVotePenaltyCounter(ctx, ValAddrs[0], missCounter, abstainCounter, successCounter) // Set the voting info

	// The handler will iterate over
	handler := func(operator sdk.ValAddress, votePenaltyCounter types.VotePenaltyCounter) (bool, error) {
		missCount := votePenaltyCounter.MissCount
		abstainCount := votePenaltyCounter.AbstainCount
		successCount := votePenaltyCounter.SuccessCount

		// validation
		require.Equal(t, missCounter, missCount)
		require.Equal(t, abstainCounter, abstainCount)
		require.Equal(t, successCounter, successCount)
		return true, nil
	}

	oracleKeeper.IterateVotePenaltyCounters(ctx, handler)
}

func TestAggregateExchangeRateLogic(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// Create and set exchange rate
	exchangeRate := types.ExchangeRateTuples{
		{Denom: "BTC/USD", ExchangeRate: math.LegacyNewDec(1)},
		{Denom: "ETH/USD", ExchangeRate: math.LegacyNewDec(2)},
		{Denom: "ATOM/USD", ExchangeRate: math.LegacyNewDec(3)},
	}
	exchangeRateVote, err := types.NewAggregateExchangeRateVote(exchangeRate, ValAddrs[0])
	require.NoError(t, err)
	err = oracleKeeper.AggregateExchangeRateVote.Set(ctx, ValAddrs[0], exchangeRateVote)
	require.NoError(t, err)

	// Get the aggregated exchange rate and validate
	gotExchangeRate, err := oracleKeeper.AggregateExchangeRateVote.Get(ctx, ValAddrs[0])
	require.NoError(t, err)
	require.Equal(t, exchangeRate, gotExchangeRate.ExchangeRateTuples)
	require.Equal(t, ValAddrs[0].String(), gotExchangeRate.Voter)

	// Delete exchange rate
	err = oracleKeeper.AggregateExchangeRateVote.Remove(ctx, ValAddrs[0]) // delete exchange rate voting
	require.NoError(t, err)
	_, err = oracleKeeper.AggregateExchangeRateVote.Get(ctx, ValAddrs[0])
	require.Error(t, err)

	// Create and aggregate invalid exchange rate
	exchangeRate = types.ExchangeRateTuples{
		{Denom: "BTC/USD", ExchangeRate: math.LegacyNewDec(0)},
		{Denom: "ETH/USD", ExchangeRate: math.LegacyNewDec(-1)},
		{Denom: "ATOM/USD", ExchangeRate: math.LegacyNewDec(2)},
	}
	_, err = types.NewAggregateExchangeRateVote(exchangeRate, ValAddrs[0])
	require.Error(t, err)
	err = oracleKeeper.AggregateExchangeRateVote.Set(ctx, ValAddrs[0], exchangeRateVote)
	require.NoError(t, err)
}

func TestIterateAggregateExchangeRateVotes(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// Aggregate exchange rates
	exchangeRate1 := types.ExchangeRateTuples{
		{Denom: "BTC/USD", ExchangeRate: math.LegacyNewDec(1)},
		{Denom: "ETH/USD", ExchangeRate: math.LegacyNewDec(2)},
		{Denom: "ATOM/USD", ExchangeRate: math.LegacyNewDec(3)},
	}
	exchangeRateVote1, err := types.NewAggregateExchangeRateVote(exchangeRate1, ValAddrs[0]) // Upload rates by val 0
	require.NoError(t, err)
	err = oracleKeeper.AggregateExchangeRateVote.Set(ctx, ValAddrs[0], exchangeRateVote1)
	require.NoError(t, err)

	exchangeRate2 := types.ExchangeRateTuples{
		{Denom: "BTC/USD", ExchangeRate: math.LegacyNewDec(4)},
		{Denom: "ETH/USD", ExchangeRate: math.LegacyNewDec(5)},
		{Denom: "ATOM/USD", ExchangeRate: math.LegacyNewDec(6)},
	}
	exchangeRateVote2, err := types.NewAggregateExchangeRateVote(exchangeRate2, ValAddrs[1]) // Upload rates by val 1
	require.NoError(t, err)
	err = oracleKeeper.AggregateExchangeRateVote.Set(ctx, ValAddrs[1], exchangeRateVote2)
	require.NoError(t, err)

	handler := func(voterAddr sdk.ValAddress, aggregateVote types.AggregateExchangeRateVote) (bool, error) {
		if voterAddr.Equals(ValAddrs[0]) {
			require.Equal(t, exchangeRateVote1, aggregateVote)
			require.Equal(t, exchangeRateVote1.Voter, voterAddr.String())
			return false, nil
		}

		require.Equal(t, exchangeRateVote2, aggregateVote)
		require.Equal(t, exchangeRateVote2.Voter, voterAddr.String())
		return false, nil
	}
	err = oracleKeeper.AggregateExchangeRateVote.Walk(ctx, nil, handler)
	require.NoError(t, err)
}

func TestRemoveExcessFeeds(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// Aggregate voting targets
	err := oracleKeeper.VoteTarget.Clear(ctx, nil)
	require.NoError(t, err)
	err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroAtomDenom, types.Denom{Name: utils.MicroAtomDenom})
	require.NoError(t, err)
	err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroEthDenom, types.Denom{Name: utils.MicroEthDenom})
	require.NoError(t, err)

	// Aggregate base exchange rate
	err = oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, math.LegacyNewDec(1))
	require.NoError(t, err)
	err = oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroEthDenom, math.LegacyNewDec(2))
	require.NoError(t, err)
	err = oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroKiiDenom, math.LegacyNewDec(3)) // extra denom
	require.NoError(t, err)

	// remove excess
	err = oracleKeeper.RemoveExcessFeeds(ctx)
	require.NoError(t, err)

	// Validate the successful erased of the extra denoms
	oracleKeeper.IterateBaseExchangeRates(ctx, func(denom string, exchangeRate types.OracleExchangeRate) (bool, error) {
		require.True(t, denom != utils.MicroKiiDenom)
		return false, nil
	})
}

func TestVoteTargetLogic(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// Set and Get Voting target
	err := oracleKeeper.VoteTarget.Clear(ctx, nil)
	require.NoError(t, err)
	voteTarget := map[string]types.Denom{
		utils.MicroKiiDenom:  {Name: utils.MicroKiiDenom},
		utils.MicroEthDenom:  {Name: utils.MicroEthDenom},
		utils.MicroUsdcDenom: {Name: utils.MicroUsdcDenom},
		utils.MicroAtomDenom: {Name: utils.MicroAtomDenom},
	}

	for denom := range voteTarget {
		err = oracleKeeper.VoteTarget.Set(ctx, denom, voteTarget[denom])
		require.NoError(t, err)
		gottenDenom, err := oracleKeeper.VoteTarget.Get(ctx, denom)
		require.NoError(t, err)
		require.Equal(t, voteTarget[denom], gottenDenom)
	}

	// Test iterate function
	handler := func(denom string, denomInfo types.Denom) (bool, error) {
		require.Equal(t, voteTarget[denom], denomInfo)
		return false, nil
	}
	err = oracleKeeper.VoteTarget.Walk(ctx, nil, handler)
	require.NoError(t, err)

	// Test delete all targets
	err = oracleKeeper.VoteTarget.Clear(ctx, nil)
	require.NoError(t, err)
	for denom := range voteTarget {
		_, err := oracleKeeper.VoteTarget.Get(ctx, denom)
		require.Error(t, err)
	}
}

func TestPriceSnapshotLogic(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// Snapshot Data
	exchangeRate1 := types.OracleExchangeRate{
		ExchangeRate:        math.LegacyNewDec(1),
		LastUpdate:          math.NewInt(1),
		LastUpdateTimestamp: 1,
	}
	exchangeRate2 := types.OracleExchangeRate{
		ExchangeRate:        math.LegacyNewDec(2),
		LastUpdate:          math.NewInt(2),
		LastUpdateTimestamp: 2,
	}
	snapshotItem1 := types.NewPriceSnapshotItem(utils.MicroKiiDenom, exchangeRate1)
	snapshotItem2 := types.NewPriceSnapshotItem(utils.MicroEthDenom, exchangeRate2)
	snapshot1 := types.NewPriceSnapshot(1, types.PriceSnapshotItems{snapshotItem1, snapshotItem1})
	snapshot2 := types.NewPriceSnapshot(2, types.PriceSnapshotItems{snapshotItem2, snapshotItem2})

	// test set and get snapshot data
	err := oracleKeeper.PriceSnapshot.Set(ctx, snapshot1.SnapshotTimestamp, snapshot1) // Set snapshot 1
	require.NoError(t, err)
	err = oracleKeeper.PriceSnapshot.Set(ctx, snapshot2.SnapshotTimestamp, snapshot2) // Set snapshot 2
	require.NoError(t, err)

	gottenSnapshot1 := oracleKeeper.GetPriceSnapshot(ctx, 1)
	gottenSnapshot2 := oracleKeeper.GetPriceSnapshot(ctx, 2)
	require.Equal(t, snapshot1, gottenSnapshot1) // validate
	require.Equal(t, snapshot2, gottenSnapshot2) // validate

	// test iterate functions
	iteration := int64(1)
	handler := func(_ int64, snapshot types.PriceSnapshot) (bool, error) {
		require.Equal(t, iteration, snapshot.SnapshotTimestamp)
		iteration++
		return false, nil
	}
	err = oracleKeeper.PriceSnapshot.Walk(ctx, nil, handler)
	require.NoError(t, err)

	iteration = int64(2)
	handlerReverse := func(snapshot types.PriceSnapshot) (bool, error) {
		require.Equal(t, iteration, snapshot.SnapshotTimestamp)
		iteration--
		return false, nil
	}
	err = oracleKeeper.IteratePriceSnapshotsReverse(ctx, handlerReverse)
	require.NoError(t, err)

	// test delete snapshot
	expected := types.PriceSnapshot{}
	err = oracleKeeper.PriceSnapshot.Remove(ctx, 1)
	require.NoError(t, err)
	result := oracleKeeper.GetPriceSnapshot(ctx, 1) // Expected empty struct
	require.Equal(t, expected, result)
}

func TestAddPriceSnapshot(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// priceSnapshot initial data
	ctx = ctx.WithBlockTime(time.Unix(3500, 0)) // by default LookbackDuration is 3600
	exchangeRate1 := types.OracleExchangeRate{
		ExchangeRate:        math.LegacyNewDec(1),
		LastUpdate:          math.NewInt(1),
		LastUpdateTimestamp: 1,
	}
	exchangeRate2 := types.OracleExchangeRate{
		ExchangeRate:        math.LegacyNewDec(2),
		LastUpdate:          math.NewInt(2),
		LastUpdateTimestamp: 2,
	}
	snapshotItem1 := types.NewPriceSnapshotItem(utils.MicroKiiDenom, exchangeRate1)
	snapshotItem2 := types.NewPriceSnapshotItem(utils.MicroEthDenom, exchangeRate2)
	snapshot1 := types.NewPriceSnapshot(1, types.PriceSnapshotItems{snapshotItem1, snapshotItem2})
	snapshot2 := types.NewPriceSnapshot(2, types.PriceSnapshotItems{snapshotItem1, snapshotItem2})

	// Add snapshots (the function will not delete nothing)
	err := oracleKeeper.AddPriceSnapshot(ctx, snapshot1) // Add snapshots 1
	require.NoError(t, err)
	err = oracleKeeper.AddPriceSnapshot(ctx, snapshot2) // Add snapshots 2
	require.NoError(t, err)

	// Validate the 2 snapshots are on the KVStore
	data1 := oracleKeeper.GetPriceSnapshot(ctx, 1)
	data2 := oracleKeeper.GetPriceSnapshot(ctx, 2)
	require.Equal(t, snapshot1, data1)
	require.Equal(t, snapshot2, data2)

	// Update the block time (time is higher than the default param)
	ctx = ctx.WithBlockTime(time.Unix(3700, 0))

	// Create new snapshot
	exchangeRate3 := types.OracleExchangeRate{
		ExchangeRate:        math.LegacyNewDec(3),
		LastUpdate:          math.NewInt(4),
		LastUpdateTimestamp: 3,
	}
	snapshotItem3 := types.NewPriceSnapshotItem(utils.MicroKiiDenom, exchangeRate3)
	snapshot3 := types.NewPriceSnapshot(1000, types.PriceSnapshotItems{snapshotItem1, snapshotItem2, snapshotItem3})

	// Add snapshots (the function will delete the snapshot 1 and 2)
	err = oracleKeeper.AddPriceSnapshot(ctx, snapshot3) // Add snapshots 3
	require.NoError(t, err)

	// Validate the snapshot 1 and 2 were deleted
	data1 = oracleKeeper.GetPriceSnapshot(ctx, 1)
	data2 = oracleKeeper.GetPriceSnapshot(ctx, 2)
	data3 := oracleKeeper.GetPriceSnapshot(ctx, 1000)

	deletedSnapshot := types.NewPriceSnapshot(0, nil)
	require.Equal(t, deletedSnapshot, data1) // data1 is empty
	require.Equal(t, deletedSnapshot, data2) // data2 is empty
	require.Equal(t, snapshot3, data3)
}

func TestClearVoteTargets(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// Eliminate initial voting target
	err := oracleKeeper.VoteTarget.Clear(ctx, nil)
	require.NoError(t, err)

	// Aggregate voting targets
	err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroAtomDenom, types.Denom{Name: utils.MicroAtomDenom})
	require.NoError(t, err)
	err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroEthDenom, types.Denom{Name: utils.MicroEthDenom})
	require.NoError(t, err)

	// Validate the voting target were successfully added
	targets, err := oracleKeeper.GetVoteTargets(ctx)
	require.NoError(t, err)
	require.True(t, len(targets) == 2)

	// Clear voting targets
	err = oracleKeeper.VoteTarget.Clear(ctx, nil)
	require.NoError(t, err)

	// Validate empty voting targets
	targets, err = oracleKeeper.GetVoteTargets(ctx)
	require.NoError(t, err)
	require.True(t, len(targets) == 0)
}

func TestSpamPreventionLogic(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// test set and get spam prevention
	ctx = ctx.WithBlockHeight(100)                                 // Set an specific block height
	err := oracleKeeper.SetSpamPreventionCounter(ctx, ValAddrs[0]) // set spam on block 100 to val 0
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(200)
	err = oracleKeeper.SetSpamPreventionCounter(ctx, ValAddrs[1]) // set spam on block 200 to val 1
	require.NoError(t, err)

	spamVal1, err := oracleKeeper.SpamPreventionCounter.Get(ctx, ValAddrs[0]) // get smap list for val 0
	require.NoError(t, err)
	spamVal2, err := oracleKeeper.SpamPreventionCounter.Get(ctx, ValAddrs[1]) // get smap list for val 1
	require.NoError(t, err)

	// Validation
	require.Equal(t, int64(100), spamVal1)
	require.Equal(t, int64(200), spamVal2)
}
