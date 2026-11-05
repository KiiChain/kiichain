package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	rewardstypes "github.com/kiichain/kiichain/v7/x/rewards/types"
)

// testRewardUpdate Tests a change to the reward schedule
func (s *IntegrationTestSuite) testRewardUpdate() {
	// Prep info
	var (
		valIdx = 0
		c      = s.chainA
		denom  = "akii"
	)
	// Amount 1k kii
	bigAkii, ok := math.NewIntFromString("1000000000000000000000")
	s.Require().True(ok)
	amount := sdk.NewCoin(denom, bigAkii)
	// Time
	now := time.Now()
	endTime := now.Add(time.Minute)
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)

	// Get initial balance
	initialBalance, err := getSpecificBalance(chainEndpoint, senderAddress.String(), denom)
	s.Require().NoError(err)

	// Get validator addresses
	validatorA, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	validatorB, err := s.chainA.validators[1].keyInfo.GetAddress()
	s.Require().NoError(err)
	valOperAddressA := sdk.ValAddress(validatorA.Bytes()).String()
	valOperAddressB := sdk.ValAddress(validatorB.Bytes()).String()

	// Get their rewards
	initialRewardsA, err := queryRewardFrom(chainEndpoint, validatorA.String(), valOperAddressA)
	s.Require().NoError(err)
	initialRewardsB, err := queryRewardFrom(chainEndpoint, validatorB.String(), valOperAddressB)
	s.Require().NoError(err)
	initialRewards := initialRewardsA.Rewards.Add(initialRewardsB.Rewards...)

	// 1. Fund pool via CLI
	s.fundRewardPool(c, valIdx, amount, senderAddress.String())

	// Check balance loss
	balance, err := getSpecificBalance(chainEndpoint, senderAddress.String(), denom)
	s.Require().NoError(err)
	s.Require().True(initialBalance.Sub(amount).IsGTE(balance))

	// Check pool change
	rewardResponse, err := queryRewardPool(chainEndpoint)
	s.Require().NoError(err)
	pool := rewardResponse.RewardPool.CommunityPool
	s.Require().False(pool.AmountOf(denom).IsZero())

	// 2. Create and pass proposal to change schedule
	s.passScheduleProposal(chainEndpoint, amount, senderAddress.String(), endTime)

	// Query changes
	scheduleResponse, err := queryReleaseSchedule(chainEndpoint)
	s.Require().NoError(err)
	schedule := scheduleResponse.ReleaseSchedule
	s.Require().Equal(schedule.TotalAmount, amount)
	s.Require().True(schedule.Active)

	// Wait time for blocks
	time.Sleep(time.Second * 10)

	// Check schedule change
	scheduleResponse, err = queryReleaseSchedule(chainEndpoint)
	s.Require().NoError(err)
	finalSchedule := scheduleResponse.ReleaseSchedule
	s.T().Logf("Scheduled amt before %s vs after %s", schedule.ReleasedAmount.Amount.String(), finalSchedule.ReleasedAmount.Amount.String())
	s.Require().True(schedule.ReleasedAmount.Amount.LT(finalSchedule.ReleasedAmount.Amount))
	s.Require().True(schedule.Active)

	// 3. Check that the sum of validator rewards is at least initial + the released amt
	finalRewardsA, err := queryRewardFrom(chainEndpoint, validatorA.String(), valOperAddressA)
	s.Require().NoError(err)
	finalRewardsB, err := queryRewardFrom(chainEndpoint, validatorB.String(), valOperAddressB)
	s.Require().NoError(err)
	// Total
	initialAkii := initialRewards.AmountOf(denom)
	initialPlusScheduled := initialAkii.Add(schedule.ReleasedAmount.Amount.ToLegacyDec())
	finalAkii := finalRewardsB.Rewards.AmountOf(denom).Add(finalRewardsA.Rewards.AmountOf(denom))
	s.T().Logf("Reward amt before %s vs after %s", initialAkii.String(), initialPlusScheduled.String())
	s.Require().True(finalAkii.GT(initialPlusScheduled))
}

// queryReleaseSchedule returns schedule information from the chain
func queryReleaseSchedule(endpoint string) (rewardstypes.QueryReleaseScheduleResponse, error) {
	var res rewardstypes.QueryReleaseScheduleResponse

	// Construct the full URL
	url := fmt.Sprintf("%s/kiichain/rewards/v1beta1/release-schedule", endpoint)

	// Make HTTP GET request
	body, err := httpGet(url)
	if err != nil {
		return res, err
	}

	// Unmarshal JSON response
	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

// queryRewardPool returns reward pool information from the chain
func queryRewardPool(endpoint string) (rewardstypes.QueryRewardPoolResponse, error) {
	var res rewardstypes.QueryRewardPoolResponse

	// Construct the full URL
	url := fmt.Sprintf("%s/kiichain/rewards/v1beta1/reward-pool", endpoint)

	// Make HTTP GET request
	body, err := httpGet(url)
	if err != nil {
		return res, err
	}

	// Unmarshal JSON response
	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

// queryRewardFrom returns reward from validator from the chain
func queryRewardFrom(endpoint string, address string, valoperAddress string) (types.QueryDelegationRewardsResponse, error) {
	var res types.QueryDelegationRewardsResponse

	// Construct the full URL
	url := fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/rewards/%s", endpoint, address, valoperAddress)

	// Make HTTP GET request
	body, err := httpGet(url)
	if err != nil {
		return res, err
	}

	// Unmarshal JSON response
	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

// fundRewardPool adds funds to the rewards pool
func (s *IntegrationTestSuite) fundRewardPool(c *chain, valIdx int, amount sdk.Coin, sender string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	kiichainCommand := []string{
		kiichaindBinary,
		txCommand,
		rewardstypes.ModuleName,
		"fund-pool",
		amount.String(),
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, "300000000akii"),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "5000000"),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeKiichainTxCommand(ctx, c, kiichainCommand, valIdx, s.defaultExecValidation(c, valIdx))
}

// passScheduleProposal passes a storage proposal
func (s *IntegrationTestSuite) passScheduleProposal(chainEndpoint string, amount sdk.Coin, sender string, endTime time.Time) {
	// Write proposal
	s.writeScheduleProposal(s.chainA, amount, endTime)

	// Create command
	proposalCounter++
	submitGovFlags := []string{configFile(proposalAddSchedule)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	// Create and pass proposal
	s.submitGovProposal(chainEndpoint, sender, proposalCounter, "ChangeSchedule", submitGovFlags, depositGovFlags, voteGovFlags, "vote")
}

// writeScheduleProposal stores a file with the change schedule proposal
func (s *IntegrationTestSuite) writeScheduleProposal(c *chain, amount sdk.Coin, endTime time.Time) {
	body := `{
		"messages": [
                {
			"@type": "/kiichain.rewards.v1beta1.MsgChangeSchedule",
            "authority": "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
            "schedule": {
                "total_amount": {
                    "denom": "%s",
                    "amount": "%s"
                },
                "released_amount": {
                    "denom": "%s",
                    "amount": "0"
                },
                "end_time": "%s",
                "last_release_time": "0001-01-01T00:00:00Z",
                "active": true
            }
        }
    ],
    "metadata": "ipfs://CID",
    "deposit": "1000akii",
    "title": "Add Schedule",
    "summary": "initial schedule"
}`

	propMsgBody := fmt.Sprintf(body, amount.Denom, amount.Amount.String(), amount.Denom, endTime.UTC().Format(time.RFC3339Nano))

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalAddSchedule), []byte(propMsgBody))
	s.Require().NoError(err)
}

// testRewardScheduleTransition simulates replacing a live release schedule with a new,
// larger live schedule through governance (the mainnet round 1 -> round 2 transition):
//   - Schedule A (round 1 amount) is set and starts releasing
//   - While A is still active, a proposal swaps it for schedule B (round 2 amount)
//   - B must take over immediately: new total/end time, released amount reset
//   - A's unreleased remainder must stay in the pool (not paid out, not lost)
//   - B must release exactly its total and then deactivate
//
// Durations are compressed to minutes; the amounts are the real mainnet ones.
func (s *IntegrationTestSuite) testRewardScheduleTransition() {
	var (
		valIdx = 0
		c      = s.chainA
		denom  = akiiDenom
	)
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[c.id][0].GetHostPort("1317/tcp"))
	senderAddress, err := c.validators[valIdx].keyInfo.GetAddress()
	s.Require().NoError(err)

	scheduleA := sdk.NewCoin(denom, mustNewIntFromString("327545000000000000000000")) // 327,545 KII (round 1)
	scheduleB := sdk.NewCoin(denom, mustNewIntFromString("771000000000000000000000")) // 771,000 KII (round 2)

	// 1. Fund the pool with enough to cover both schedules
	// (ChangeSchedule requires the pool to hold the full new total_amount)
	s.fundRewardPool(c, valIdx, scheduleA.Add(scheduleB), senderAddress.String())

	// 2. Start schedule A, long enough to still be live when B's proposal passes
	endTimeA := time.Now().Add(4 * time.Minute)
	s.passScheduleProposal(chainEndpoint, scheduleA, senderAddress.String(), endTimeA)

	s.Require().Eventually(func() bool {
		res, err := queryReleaseSchedule(chainEndpoint)
		s.Require().NoError(err)
		sch := res.ReleaseSchedule
		return sch.Active && sch.TotalAmount.Equal(scheduleA) && sch.ReleasedAmount.Amount.IsPositive()
	}, time.Minute, time.Second, "schedule A never started releasing")

	// 3. Snapshot A while live. The schedule is queried before the pool, so the pool
	// snapshot height is >= the schedule snapshot height (used for the bounds below)
	res, err := queryReleaseSchedule(chainEndpoint)
	s.Require().NoError(err)
	liveA := res.ReleaseSchedule
	s.Require().True(liveA.Active)
	s.Require().True(liveA.ReleasedAmount.Amount.LT(scheduleA.Amount), "schedule A finished before the transition")

	poolRes, err := queryRewardPool(chainEndpoint)
	s.Require().NoError(err)
	poolPre := poolRes.RewardPool.CommunityPool.AmountOf(denom)
	s.T().Logf("schedule A live: released %s / %s, pool %s", liveA.ReleasedAmount, liveA.TotalAmount, poolPre)

	// 4. Replace A with B while A is still active
	switchStart := time.Now()
	endTimeB := switchStart.Add(90 * time.Second)
	s.passScheduleProposal(chainEndpoint, scheduleB, senderAddress.String(), endTimeB)

	res, err = queryReleaseSchedule(chainEndpoint)
	s.Require().NoError(err)
	newB := res.ReleaseSchedule
	s.Require().True(newB.Active)
	s.Require().True(newB.TotalAmount.Equal(scheduleB), "total amount not replaced: %s", newB.TotalAmount)
	s.Require().True(newB.EndTime.Sub(endTimeB).Abs() < time.Second,
		"end time not replaced: %s", newB.EndTime)
	// Released amount was reset: B can only have released for the few blocks since it was set
	s.Require().True(newB.ReleasedAmount.Amount.LT(scheduleB.Amount.QuoRaw(2)),
		"released amount was not reset for schedule B: %s", newB.ReleasedAmount)
	s.Require().True(newB.LastReleaseTime.After(switchStart.Add(-5*time.Second)),
		"last release time %s predates the transition", newB.LastReleaseTime)

	// 5. Wait for B to release everything and deactivate
	var finalB rewardstypes.ReleaseSchedule
	s.Require().Eventually(func() bool {
		res, err := queryReleaseSchedule(chainEndpoint)
		s.Require().NoError(err)
		finalB = res.ReleaseSchedule
		return !finalB.Active
	}, 3*time.Minute, 2*time.Second, "schedule B never completed")
	s.Require().True(finalB.ReleasedAmount.Equal(scheduleB), "schedule B released %s, expected %s", finalB.ReleasedAmount, scheduleB)

	// 6. Pool accounting. With X the pool before A started releasing:
	//   poolPre   = X - releasedA(at snapshot)
	//   poolFinal = X - releasedA(at switch) - totalB
	// and releasedA(snapshot) <= releasedA(switch) <= totalA, so:
	//   poolPre + liveA.released - totalA - totalB <= poolFinal <= poolPre - totalB
	poolRes, err = queryRewardPool(chainEndpoint)
	s.Require().NoError(err)
	poolFinal := poolRes.RewardPool.CommunityPool.AmountOf(denom)

	totalA := scheduleA.Amount.ToLegacyDec()
	totalB := scheduleB.Amount.ToLegacyDec()
	upper := poolPre.Sub(totalB)
	lower := poolPre.Add(liveA.ReleasedAmount.Amount.ToLegacyDec()).Sub(totalA).Sub(totalB)
	s.T().Logf("pool final %s, expected in [%s, %s]", poolFinal, lower, upper)
	s.Require().True(poolFinal.LTE(upper), "pool paid out less than schedule B's total")
	s.Require().True(poolFinal.GTE(lower), "pool paid out more than A's cap + B's total")

	// A ran for well under half of its window before being replaced, so at least half
	// of A must still be sitting in the pool (it was not drained by the transition)
	s.Require().True(poolFinal.GTE(upper.Sub(totalA.QuoInt64(2))),
		"schedule A's unreleased remainder did not stay in the pool")
}
