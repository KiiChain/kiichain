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

	rewardstypes "github.com/kiichain/kiichain/v2/x/rewards/types"
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
