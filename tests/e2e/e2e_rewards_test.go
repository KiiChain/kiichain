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
		c     = s.chainA
		denom = "akii"
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
	s.fundRewardPool(c, amount, senderAddress.String())

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
// It sends the transaction through the first validator
func (s *IntegrationTestSuite) fundRewardPool(c *chain, amount sdk.Coin, sender string) {
	const valIdx = 0

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

// testRewardScheduleReplaceLive covers the mainnet corner case where the next schedule is
// proposed while the current one is still live: nothing waits for a previous schedule to
// end. Schedule A (mainnet phase 1 amount) is started and, as soon as it is releasing,
// replaced through governance by schedule B (proposal 17 amount, 767,993 KII).
//
// It checks that A was still live when B's proposal went in, that B takes over with its
// released amount reset, releases exactly its total and deactivates, and that A's
// unreleased remainder stays in the pool.
func (s *IntegrationTestSuite) testRewardScheduleReplaceLive() {
	var (
		valIdx = 0
		c      = s.chainA
		denom  = akiiDenom
	)
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[c.id][0].GetHostPort("1317/tcp"))
	senderAddress, err := c.validators[valIdx].keyInfo.GetAddress()
	s.Require().NoError(err)

	scheduleA := sdk.NewCoin(denom, mustNewIntFromString("327545000000000000000000")) // 327,545 KII (phase 1)
	scheduleB := sdk.NewCoin(denom, mustNewIntFromString("767993000000000000000000")) // 767,993 KII (phase 2, proposal 17)

	// 1. Fund both schedules up front so B can be proposed as soon as A is live
	s.fundRewardPool(c, scheduleA.Add(scheduleB), senderAddress.String())

	// 2. Start A; any schedule left by a previous test is replaced while live as well
	s.passScheduleProposal(chainEndpoint, scheduleA, senderAddress.String(), time.Now().Add(5*time.Minute))
	s.waitScheduleReleasing(chainEndpoint, scheduleA)

	// 3. Snapshot while A is live. The pool is queried before the schedule, so A's released
	// amount below is >= what it had released at the pool snapshot
	poolPre := s.rewardPoolAmount(chainEndpoint)
	liveA := s.requireScheduleLive(chainEndpoint, scheduleA)
	s.T().Logf("schedule A live: released %s / %s, pool %s", liveA.ReleasedAmount, liveA.TotalAmount, poolPre)

	// 4. Replace A with B while A is live
	switchStart := time.Now()
	endTimeB := switchStart.Add(60 * time.Second)
	s.passScheduleProposal(chainEndpoint, scheduleB, senderAddress.String(), endTimeB)
	s.requireScheduleTookOver(chainEndpoint, scheduleB, endTimeB, switchStart)

	// 5. B releases exactly its total and deactivates
	s.waitScheduleCompleted(chainEndpoint, scheduleB)

	// 6. A's remainder stayed in the pool. With releasedA = what A paid out in total:
	//   poolFinal = poolPre + releasedA(at pool snapshot) - releasedA - totalB
	// A's leftover (totalA - releasedA) is positive iff
	//   poolFinal > poolPre + releasedA(at pool snapshot) - totalA - totalB
	// and releasedA(at pool snapshot) <= liveA.released, so this bound is sufficient
	poolFinal := s.rewardPoolAmount(chainEndpoint)
	bound := poolPre.Add(liveA.ReleasedAmount.Amount.ToLegacyDec()).
		Sub(scheduleA.Amount.ToLegacyDec()).Sub(scheduleB.Amount.ToLegacyDec())
	s.T().Logf("pool final %s, must be > %s and <= %s", poolFinal, bound, poolPre.Sub(scheduleB.Amount.ToLegacyDec()))
	s.Require().True(poolFinal.GT(bound), "schedule A's unreleased remainder did not stay in the pool")
	s.Require().True(poolFinal.LTE(poolPre.Sub(scheduleB.Amount.ToLegacyDec())), "pool paid out less than schedule B's total")
}

// testRewardScheduleTransition simulates the phase 1 -> phase 2 transition with funding in
// two phases and exact pool accounting:
//   - Fund x and set schedule A for exactly x
//   - While A is still releasing, fund 2x and replace A with schedule B for 2x
//
// It checks that B takes over, releases exactly 2x and deactivates, and that the pool ends
// with exactly A's unreleased remainder.
func (s *IntegrationTestSuite) testRewardScheduleTransition() {
	var (
		valIdx = 0
		c      = s.chainA
		denom  = akiiDenom
	)
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[c.id][0].GetHostPort("1317/tcp"))
	senderAddress, err := c.validators[valIdx].keyInfo.GetAddress()
	s.Require().NoError(err)

	scheduleA := sdk.NewCoin(denom, mustNewIntFromString("327545000000000000000000")) // x = 327,545 KII (phase 1)
	scheduleB := sdk.NewCoin(denom, scheduleA.Amount.MulRaw(2))                       // 2x = 655,090 KII (phase 2)

	// 0. Wait for any schedule left by a previous test to finish, so the pool only
	// changes through this test's funding and schedules (needed for exact accounting)
	s.Require().Eventually(func() bool {
		res, err := queryReleaseSchedule(chainEndpoint)
		s.Require().NoError(err)
		return !res.ReleaseSchedule.Active
	}, 2*time.Minute, 2*time.Second, "previous release schedule never finished")
	pool0 := s.rewardPoolAmount(chainEndpoint)

	// 1. Fund x and set schedule A for exactly x
	s.fundRewardPool(c, scheduleA, senderAddress.String())
	s.Require().True(s.rewardPoolAmount(chainEndpoint).Equal(pool0.Add(scheduleA.Amount.ToLegacyDec())),
		"pool was not funded with schedule A's amount")

	s.passScheduleProposal(chainEndpoint, scheduleA, senderAddress.String(), time.Now().Add(5*time.Minute))
	liveA := s.waitScheduleReleasing(chainEndpoint, scheduleA)
	s.T().Logf("schedule A live: released %s / %s", liveA.ReleasedAmount, liveA.TotalAmount)

	// 2. Fund 2x while A is live and replace A with schedule B for 2x
	s.fundRewardPool(c, scheduleB, senderAddress.String())

	switchStart := time.Now()
	endTimeB := switchStart.Add(90 * time.Second)
	s.passScheduleProposal(chainEndpoint, scheduleB, senderAddress.String(), endTimeB)
	s.requireScheduleTookOver(chainEndpoint, scheduleB, endTimeB, switchStart)

	// 3. Wait for B to release everything and deactivate
	s.waitScheduleCompleted(chainEndpoint, scheduleB)

	// 4. Pool accounting, exact since nothing else touched the pool:
	//   poolFinal = pool0 + x + 2x - releasedA - 2x  =>  releasedA = pool0 + x - poolFinal
	// releasedA is what A actually paid out before B replaced it
	poolFinal := s.rewardPoolAmount(chainEndpoint)
	releasedA := pool0.Add(scheduleA.Amount.ToLegacyDec()).Sub(poolFinal)
	leftoverA := scheduleA.Amount.ToLegacyDec().Sub(releasedA)
	s.T().Logf("pool final %s: schedule A released %s before the switch, %s left in the pool", poolFinal, releasedA, leftoverA)

	// A paid out at least what was observed while it was live...
	s.Require().True(releasedA.GTE(liveA.ReleasedAmount.Amount.ToLegacyDec()),
		"schedule A released %s, less than the %s observed while live", releasedA, liveA.ReleasedAmount)
	// ...but was replaced before finishing, and its unreleased remainder is still in the pool
	s.Require().True(leftoverA.IsPositive(), "schedule A's unreleased remainder did not stay in the pool")
}

// rewardPoolAmount returns the reward pool's akii balance
func (s *IntegrationTestSuite) rewardPoolAmount(endpoint string) math.LegacyDec {
	res, err := queryRewardPool(endpoint)
	s.Require().NoError(err)
	return res.RewardPool.CommunityPool.AmountOf(akiiDenom)
}

// waitScheduleReleasing waits until the schedule for total is active and has released something
func (s *IntegrationTestSuite) waitScheduleReleasing(endpoint string, total sdk.Coin) rewardstypes.ReleaseSchedule {
	var schedule rewardstypes.ReleaseSchedule
	s.Require().Eventually(func() bool {
		res, err := queryReleaseSchedule(endpoint)
		s.Require().NoError(err)
		schedule = res.ReleaseSchedule
		return schedule.Active && schedule.TotalAmount.Equal(total) && schedule.ReleasedAmount.Amount.IsPositive()
	}, time.Minute, time.Second, "schedule for %s never started releasing", total)
	return schedule
}

// requireScheduleLive asserts the current schedule is the one for total and has not ended
func (s *IntegrationTestSuite) requireScheduleLive(endpoint string, total sdk.Coin) rewardstypes.ReleaseSchedule {
	res, err := queryReleaseSchedule(endpoint)
	s.Require().NoError(err)
	schedule := res.ReleaseSchedule
	s.Require().True(schedule.Active && schedule.TotalAmount.Equal(total) && schedule.ReleasedAmount.Amount.LT(total.Amount),
		"schedule for %s is not live: %v", total, schedule)
	return schedule
}

// requireScheduleTookOver waits for the schedule for total to replace the previous one and
// asserts it started fresh. ChangeSchedule stores it with a zero LastReleaseTime, which
// BeginBlocker initializes in a later block, so this polls for that first
func (s *IntegrationTestSuite) requireScheduleTookOver(endpoint string, total sdk.Coin, endTime, switchStart time.Time) {
	var schedule rewardstypes.ReleaseSchedule
	s.Require().Eventually(func() bool {
		res, err := queryReleaseSchedule(endpoint)
		s.Require().NoError(err)
		schedule = res.ReleaseSchedule
		return schedule.TotalAmount.Equal(total) && !schedule.LastReleaseTime.IsZero()
	}, 30*time.Second, time.Second, "schedule for %s was not set or never initialized", total)
	s.Require().True(schedule.Active)
	s.Require().True(schedule.EndTime.Sub(endTime).Abs() < time.Second, "end time not replaced: %s", schedule.EndTime)
	// Released amount was reset: it can only cover the few blocks since the schedule was set
	s.Require().True(schedule.ReleasedAmount.Amount.LT(total.Amount.QuoRaw(2)),
		"released amount was not reset: %s", schedule.ReleasedAmount)
	s.Require().True(schedule.LastReleaseTime.After(switchStart),
		"last release time %s predates the transition", schedule.LastReleaseTime)
}

// waitScheduleCompleted waits for the current schedule to deactivate and asserts it released exactly total
func (s *IntegrationTestSuite) waitScheduleCompleted(endpoint string, total sdk.Coin) {
	var schedule rewardstypes.ReleaseSchedule
	s.Require().Eventually(func() bool {
		res, err := queryReleaseSchedule(endpoint)
		s.Require().NoError(err)
		schedule = res.ReleaseSchedule
		return !schedule.Active
	}, 3*time.Minute, 2*time.Second, "schedule for %s never completed", total)
	s.Require().True(schedule.ReleasedAmount.Equal(total), "schedule released %s, expected %s", schedule.ReleasedAmount, total)
}
