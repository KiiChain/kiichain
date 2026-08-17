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

// testRewardUpdate tests inflation-based emissions after funding the pool and setting supply_base
func (s *IntegrationTestSuite) testRewardUpdate() {
	var (
		valIdx = 0
		c      = s.chainA
		denom  = "akii"
	)
	bigAkii, ok := math.NewIntFromString("1000000000000000000000")
	s.Require().True(ok)
	amount := sdk.NewCoin(denom, bigAkii)
	supplyBase, ok := math.NewIntFromString("1000000000000000000000000") // 1e24
	s.Require().True(ok)

	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)

	initialBalance, err := getSpecificBalance(chainEndpoint, senderAddress.String(), denom)
	s.Require().NoError(err)

	validatorA, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	validatorB, err := s.chainA.validators[1].keyInfo.GetAddress()
	s.Require().NoError(err)
	valOperAddressA := sdk.ValAddress(validatorA.Bytes()).String()
	valOperAddressB := sdk.ValAddress(validatorB.Bytes()).String()

	initialRewardsA, err := queryRewardFrom(chainEndpoint, validatorA.String(), valOperAddressA)
	s.Require().NoError(err)
	initialRewardsB, err := queryRewardFrom(chainEndpoint, validatorB.String(), valOperAddressB)
	s.Require().NoError(err)
	initialRewards := initialRewardsA.Rewards.Add(initialRewardsB.Rewards...)

	s.fundRewardPool(c, valIdx, amount, senderAddress.String())

	balance, err := getSpecificBalance(chainEndpoint, senderAddress.String(), denom)
	s.Require().NoError(err)
	s.Require().True(initialBalance.Sub(amount).IsGTE(balance))

	rewardResponse, err := queryRewardPool(chainEndpoint)
	s.Require().NoError(err)
	pool := rewardResponse.RewardPool.CommunityPool
	s.Require().False(pool.AmountOf(denom).IsZero())
	initialPoolAmount := pool.AmountOf(denom)
	initialTotalReleased := rewardResponse.RewardPool.TotalReleased

	s.passRewardsParamsProposal(chainEndpoint, senderAddress.String(), supplyBase)

	time.Sleep(time.Second * 10)

	rewardResponse, err = queryRewardPool(chainEndpoint)
	s.Require().NoError(err)
	finalPool := rewardResponse.RewardPool
	s.T().Logf("Pool before %s vs after %s", initialPoolAmount.String(), finalPool.CommunityPool.AmountOf(denom).String())
	s.Require().True(finalPool.CommunityPool.AmountOf(denom).LT(initialPoolAmount))
	if !initialTotalReleased.IsNil() && !initialTotalReleased.IsZero() {
		s.Require().True(finalPool.TotalReleased.Amount.GT(initialTotalReleased.Amount))
	} else {
		s.Require().False(finalPool.TotalReleased.IsZero())
	}

	finalRewardsA, err := queryRewardFrom(chainEndpoint, validatorA.String(), valOperAddressA)
	s.Require().NoError(err)
	finalRewardsB, err := queryRewardFrom(chainEndpoint, validatorB.String(), valOperAddressB)
	s.Require().NoError(err)

	initialAkii := initialRewards.AmountOf(denom)
	finalAkii := finalRewardsB.Rewards.AmountOf(denom).Add(finalRewardsA.Rewards.AmountOf(denom))
	s.T().Logf("Reward amt before %s vs after %s", initialAkii.String(), finalAkii.String())
	s.Require().True(finalAkii.GT(initialAkii))
}

func queryRewardPool(endpoint string) (rewardstypes.QueryRewardPoolResponse, error) {
	var res rewardstypes.QueryRewardPoolResponse

	url := fmt.Sprintf("%s/kiichain/rewards/v1beta1/reward-pool", endpoint)

	body, err := httpGet(url)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func queryRewardFrom(endpoint string, address string, valoperAddress string) (types.QueryDelegationRewardsResponse, error) {
	var res types.QueryDelegationRewardsResponse

	url := fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/rewards/%s", endpoint, address, valoperAddress)

	body, err := httpGet(url)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

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

func (s *IntegrationTestSuite) passRewardsParamsProposal(chainEndpoint string, sender string, supplyBase math.Int) {
	s.writeRewardsParamsProposal(s.chainA, supplyBase)

	proposalCounter++
	submitGovFlags := []string{configFile(proposalUpdateRewardsParams)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	s.submitGovProposal(chainEndpoint, sender, proposalCounter, "UpdateParams", submitGovFlags, depositGovFlags, voteGovFlags, "vote")
}

func (s *IntegrationTestSuite) writeRewardsParamsProposal(c *chain, supplyBase math.Int) {
	body := `{
		"messages": [
                {
			"@type": "/kiichain.rewards.v1beta1.MsgUpdateParams",
            "authority": "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
            "params": {
                "token_denom": "akii",
                "goal_bonded": "0.670000000000000000",
                "inflation_min": "0.000000000000000000",
                "inflation_max": "0.200000000000000000",
                "supply_base": "%s",
                "inflation_rate_change": "0.130000000000000000"
            }
        }
    ],
    "metadata": "ipfs://CID",
    "deposit": "1000akii",
    "title": "Enable Rewards Emissions",
    "summary": "set supply_base to enable inflation-based emissions"
}`

	propMsgBody := fmt.Sprintf(body, supplyBase.String())

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalUpdateRewardsParams), []byte(propMsgBody))
	s.Require().NoError(err)
}
