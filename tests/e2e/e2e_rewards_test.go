package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	rewardstypes "github.com/kiichain/kiichain/v7/x/rewards/types"
)

// rewardDistributedEvent is a parsed reward_distributed event from a block.
type rewardDistributedEvent struct {
	Amount        sdk.Coin
	TotalReleased sdk.Coin
	InflationRate math.LegacyDec
	BondedRatio   math.LegacyDec
}

// testRewards walks the x/rewards lifecycle:
//  1. default params ship with emissions off (supply_base = 0)
//  2. a funded pool does not emit while supply_base = 0
//  3. gov sets supply_base > 0; each block releases exactly
//     clamp(curve(bondedRatio)) × supply_base / blocks_per_year from the pool
//     into fee_collector and validators accrue rewards
//  4. gov sets supply_base back to 0; emissions stop and funds stay in the pool
func (s *IntegrationTestSuite) testRewards() {
	var (
		valIdx = 0
		c      = s.chainA
		denom  = "akii"
	)
	fundAmount, ok := math.NewIntFromString("1000000000000000000000") // 1000 KII
	s.Require().True(ok)
	amount := sdk.NewCoin(denom, fundAmount)
	supplyBase, ok := math.NewIntFromString("1000000000000000000000000") // 1e24
	s.Require().True(ok)

	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[c.id][valIdx].GetHostPort("1317/tcp"))
	rpcClient, err := rpchttp.New(fmt.Sprintf("tcp://%s", s.valResources[c.id][valIdx].GetHostPort("26657/tcp")), "/websocket")
	s.Require().NoError(err)

	senderAddress, err := c.validators[valIdx].keyInfo.GetAddress()
	s.Require().NoError(err)

	s.Run("default params have emissions disabled", func() {
		params, err := queryRewardsParams(chainEndpoint)
		s.Require().NoError(err)
		s.Require().Equal(denom, params.TokenDenom)
		s.Require().True(params.SupplyBase.IsZero(), "supply_base should default to 0, got %s", params.SupplyBase)
		s.Require().NoError(params.ValidateBasic())
	})

	s.Run("funded pool does not emit while supply_base is zero", func() {
		initialBalance, err := getSpecificBalance(chainEndpoint, senderAddress.String(), denom)
		s.Require().NoError(err)

		s.fundRewardPool(c, valIdx, amount, senderAddress.String())

		balance, err := getSpecificBalance(chainEndpoint, senderAddress.String(), denom)
		s.Require().NoError(err)
		s.Require().True(initialBalance.Sub(amount).IsGTE(balance))

		pool, err := queryRewardPool(chainEndpoint)
		s.Require().NoError(err)
		s.Require().True(pool.Balance.Amount.GTE(fundAmount), "pool balance %s should hold the funded amount", pool.Balance)

		s.requireNoEmissions(c, rpcClient, chainEndpoint, valIdx)
	})

	s.Run("legacy v1beta1 reward-pool reports the module balance", func() {
		height := int64(s.getLatestBlockHeight(c, valIdx))
		pool, err := queryRewardPoolAtHeight(chainEndpoint, height)
		s.Require().NoError(err)

		body, err := httpGetAtHeight(fmt.Sprintf("%s/kiichain/rewards/v1beta1/reward-pool", chainEndpoint), height)
		s.Require().NoError(err)
		s.Require().Contains(string(body), pool.Balance.Amount.String())
	})

	s.Run("supply_base > 0 releases the curve amount every block", func() {
		validatorA, err := c.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		validatorB, err := c.validators[1].keyInfo.GetAddress()
		s.Require().NoError(err)
		valOperAddressA := sdk.ValAddress(validatorA.Bytes()).String()
		valOperAddressB := sdk.ValAddress(validatorB.Bytes()).String()

		initialRewardsA, err := queryRewardFrom(chainEndpoint, validatorA.String(), valOperAddressA)
		s.Require().NoError(err)
		initialRewardsB, err := queryRewardFrom(chainEndpoint, validatorB.String(), valOperAddressB)
		s.Require().NoError(err)
		initialAkii := initialRewardsA.Rewards.AmountOf(denom).Add(initialRewardsB.Rewards.AmountOf(denom))

		initialPool, err := queryRewardPool(chainEndpoint)
		s.Require().NoError(err)

		s.passRewardsParamsProposal(chainEndpoint, senderAddress.String(), supplyBase)

		params, err := queryRewardsParams(chainEndpoint)
		s.Require().NoError(err)
		s.Require().Equal(supplyBase.String(), params.SupplyBase.String())

		// Let a few blocks emit, then verify the latest one exactly.
		s.waitUntilPassedHeight(c, valIdx, s.getLatestBlockHeight(c, valIdx)+2)
		height := int64(s.getLatestBlockHeight(c, valIdx))

		event, found := s.findRewardDistributedEvent(rpcClient, height)
		s.Require().True(found, "expected a reward_distributed event at height %d", height)

		// BeginBlock at `height` reads the state committed at `height-1`.
		expectedRatio := s.queryBondedRatioAtHeight(chainEndpoint, height-1, denom)
		s.Require().True(expectedRatio.Equal(event.BondedRatio),
			"event bonded_ratio %s != staking bonded ratio %s", event.BondedRatio, expectedRatio)

		expectedAmount, expectedInflation := rewardstypes.CalculateReward(event.BondedRatio, params)
		s.Require().True(expectedAmount.IsPositive(), "curve should emit at bonded ratio %s", event.BondedRatio)
		s.Require().True(expectedInflation.Equal(event.InflationRate),
			"event inflation_rate %s != expected %s", event.InflationRate, expectedInflation)
		s.Require().True(expectedInflation.GTE(params.InflationMin) && expectedInflation.LTE(params.InflationMax))
		s.Require().Equal(expectedAmount.String(), event.Amount.String())

		// The released amount leaves the pool and is added to total_released.
		before, err := queryRewardPoolAtHeight(chainEndpoint, height-1)
		s.Require().NoError(err)
		after, err := queryRewardPoolAtHeight(chainEndpoint, height)
		s.Require().NoError(err)
		s.Require().Equal(event.Amount.Amount.String(), before.Balance.Amount.Sub(after.Balance.Amount).String(),
			"pool balance should drop by exactly the released amount")
		s.Require().Equal(event.TotalReleased.String(), after.RewardPool.TotalReleased.String())
		releasedThisBlock := after.RewardPool.TotalReleased.Amount
		if !before.RewardPool.TotalReleased.IsNil() {
			releasedThisBlock = releasedThisBlock.Sub(before.RewardPool.TotalReleased.Amount)
		}
		s.Require().Equal(event.Amount.Amount.String(), releasedThisBlock.String())

		// Everything that left the pool since funding is accounted for in total_released.
		poolDrop := initialPool.Balance.Amount.Sub(after.Balance.Amount)
		totalReleased := after.RewardPool.TotalReleased.Amount
		if !initialPool.RewardPool.TotalReleased.IsNil() {
			totalReleased = totalReleased.Sub(initialPool.RewardPool.TotalReleased.Amount)
		}
		s.Require().True(poolDrop.IsPositive())
		s.Require().Equal(poolDrop.String(), totalReleased.String())

		finalRewardsA, err := queryRewardFrom(chainEndpoint, validatorA.String(), valOperAddressA)
		s.Require().NoError(err)
		finalRewardsB, err := queryRewardFrom(chainEndpoint, validatorB.String(), valOperAddressB)
		s.Require().NoError(err)
		finalAkii := finalRewardsA.Rewards.AmountOf(denom).Add(finalRewardsB.Rewards.AmountOf(denom))
		s.T().Logf("Validator rewards before %s vs after %s", initialAkii, finalAkii)
		s.Require().True(finalAkii.GT(initialAkii))
	})

	s.Run("supply_base = 0 stops emissions and keeps funds in the pool", func() {
		s.passRewardsParamsProposal(chainEndpoint, senderAddress.String(), math.ZeroInt())

		params, err := queryRewardsParams(chainEndpoint)
		s.Require().NoError(err)
		s.Require().True(params.SupplyBase.IsZero())

		pool, err := queryRewardPool(chainEndpoint)
		s.Require().NoError(err)
		s.Require().True(pool.Balance.IsPositive(), "unreleased funds should stay in the pool")

		s.requireNoEmissions(c, rpcClient, chainEndpoint, valIdx)
	})
}

// requireNoEmissions waits for new blocks and checks that none of them emitted
// and that the pool balance and total_released did not change.
func (s *IntegrationTestSuite) requireNoEmissions(c *chain, rpcClient *rpchttp.HTTP, endpoint string, valIdx int) {
	start := int64(s.getLatestBlockHeight(c, valIdx))
	s.waitUntilPassedHeight(c, valIdx, int(start)+2)
	end := int64(s.getLatestBlockHeight(c, valIdx))

	for h := start + 1; h <= end; h++ {
		_, found := s.findRewardDistributedEvent(rpcClient, h)
		s.Require().False(found, "unexpected reward_distributed event at height %d", h)
	}

	before, err := queryRewardPoolAtHeight(endpoint, start)
	s.Require().NoError(err)
	after, err := queryRewardPoolAtHeight(endpoint, end)
	s.Require().NoError(err)
	s.Require().Equal(before.Balance.String(), after.Balance.String())
	s.Require().Equal(before.RewardPool.TotalReleased.String(), after.RewardPool.TotalReleased.String())
}

// findRewardDistributedEvent returns the reward_distributed event emitted at height, if any.
func (s *IntegrationTestSuite) findRewardDistributedEvent(rpcClient *rpchttp.HTTP, height int64) (rewardDistributedEvent, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := rpcClient.BlockResults(ctx, &height)
	s.Require().NoError(err)

	for _, ev := range res.FinalizeBlockEvents {
		if ev.Type != rewardstypes.EventTypeRewardDistributed {
			continue
		}
		attrs := make(map[string]string, len(ev.Attributes))
		for _, a := range ev.Attributes {
			attrs[a.Key] = a.Value
		}

		var out rewardDistributedEvent
		out.Amount, err = sdk.ParseCoinNormalized(attrs[rewardstypes.AttributeKeyAmount])
		s.Require().NoError(err)
		out.TotalReleased, err = sdk.ParseCoinNormalized(attrs[rewardstypes.AttributeKeyTotalReleased])
		s.Require().NoError(err)
		out.InflationRate, err = math.LegacyNewDecFromStr(attrs[rewardstypes.AttributeKeyInflationRate])
		s.Require().NoError(err)
		out.BondedRatio, err = math.LegacyNewDecFromStr(attrs[rewardstypes.AttributeKeyBondedRatio])
		s.Require().NoError(err)
		return out, true
	}
	return rewardDistributedEvent{}, false
}

// queryBondedRatioAtHeight mirrors x/staking BondedRatio: bonded tokens / bond denom supply.
func (s *IntegrationTestSuite) queryBondedRatioAtHeight(endpoint string, height int64, denom string) math.LegacyDec {
	body, err := httpGetAtHeight(fmt.Sprintf("%s/cosmos/staking/v1beta1/pool", endpoint), height)
	s.Require().NoError(err)
	var poolRes stakingtypes.QueryPoolResponse
	s.Require().NoError(cdc.UnmarshalJSON(body, &poolRes))

	body, err = httpGetAtHeight(fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/by_denom?denom=%s", endpoint, denom), height)
	s.Require().NoError(err)
	var supplyRes banktypes.QuerySupplyOfResponse
	s.Require().NoError(cdc.UnmarshalJSON(body, &supplyRes))

	s.Require().True(supplyRes.Amount.IsPositive())
	return math.LegacyNewDecFromInt(poolRes.Pool.BondedTokens).QuoInt(supplyRes.Amount.Amount)
}

func queryRewardsParams(endpoint string) (rewardstypes.Params, error) {
	var res rewardstypes.QueryParamsResponse

	body, err := httpGet(fmt.Sprintf("%s/kiichain/rewards/v1/params", endpoint))
	if err != nil {
		return res.Params, err
	}
	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res.Params, err
	}
	return res.Params, nil
}

func queryRewardPool(endpoint string) (rewardstypes.QueryRewardPoolResponse, error) {
	var res rewardstypes.QueryRewardPoolResponse

	body, err := httpGet(fmt.Sprintf("%s/kiichain/rewards/v1/reward-pool", endpoint))
	if err != nil {
		return res, err
	}
	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryRewardPoolAtHeight(endpoint string, height int64) (rewardstypes.QueryRewardPoolResponse, error) {
	var res rewardstypes.QueryRewardPoolResponse

	body, err := httpGetAtHeight(fmt.Sprintf("%s/kiichain/rewards/v1/reward-pool", endpoint), height)
	if err != nil {
		return res, err
	}
	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

// httpGetAtHeight queries the REST gateway against the state committed at height.
func httpGetAtHeight(url string, height int64) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil) //nolint:noctx // only used during tests
	if err != nil {
		return nil, err
	}
	req.Header.Set(grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(height, 10))

	resp, err := http.DefaultClient.Do(req) //nolint:gosec // only used during tests
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s at height %d: status %d: %s", url, height, resp.StatusCode, body)
	}
	return body, nil
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
	moduleAddr := authtypes.NewModuleAddress(rewardstypes.ModuleName).String()
	s.execBankSend(c, valIdx, sender, moduleAddr, amount.String(), standardFees.String(), false)
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
			"@type": "/kiichain.rewards.v1.MsgUpdateParams",
            "authority": "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
            "params": {
                "token_denom": "akii",
                "goal_bonded": "0.670000000000000000",
                "inflation_min": "0.000000000000000000",
                "inflation_max": "0.200000000000000000",
                "supply_base": "%s",
                "inflation_rate_change": "0.130000000000000000",
                "blocks_per_year": 15778800
            }
        }
    ],
    "metadata": "ipfs://CID",
    "deposit": "100akii",
    "title": "Update Rewards Emissions",
    "summary": "set supply_base to %s"
}`

	propMsgBody := fmt.Sprintf(body, supplyBase.String(), supplyBase.String())

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalUpdateRewardsParams), []byte(propMsgBody))
	s.Require().NoError(err)
}
