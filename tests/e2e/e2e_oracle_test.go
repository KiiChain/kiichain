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

	oracletypes "github.com/kiichain/kiichain/v4/x/oracle/types"
)

const (
	BlocksPerPeriod = 15
	SlashWindow     = 48960
)

var Fee = sdk.NewCoin(akiiDenom, math.NewInt(500000000))

// Test the following:
// - Feeless TXs
// - Double Feeless TXs sending
// - Slashing
// - Voting
// - Feeder address

// testFeelessTx tests the capability of sending oracle votes without fees
// This also tests the double-feeless TXs sending
func (s *IntegrationTestSuite) testFeelessTx() {
	// Take the chain endpoint
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	// Get the first validator information
	validatorA := s.chainA.validators[0]
	voterAddr, _ := validatorA.keyInfo.GetAddress()
	validatorAddress := sdk.ValAddress(voterAddr).String()

	// Check if the oracle parameters are set correctly
	s.checkAndUpdateOracleParams()

	// Get the balance for the key
	balance, err := getSpecificBalance(chainEndpoint, voterAddr.String(), akiiDenom)
	s.Require().NoError(err, "failed to get balance for %s", voterAddr.String())

	// Vote on the exchange rate
	s.execAggregateVote(s.chainA, 0, "1000akii", validatorAddress, voterAddr.String(), kiichainHomePath, Fee.String(), nil)

	// The balance should be the same as before, since the vote is fee-less
	balanceAfterFirstVote, err := getSpecificBalance(chainEndpoint, voterAddr.String(), akiiDenom)
	s.Require().NoError(err, "failed to get balance for %s after voting", voterAddr.String())
	s.Require().Equal(balance.Amount, balanceAfterFirstVote.Amount, "balance should remain the same after fee-less vote")

	// If we vote again, the balance should change
	s.execAggregateVote(s.chainA, 0, "1000akii", validatorAddress, voterAddr.String(), kiichainHomePath, Fee.String(), nil)

	// Get the new balance after the second vote
	balanceAfterSecondVote, err := getSpecificBalance(chainEndpoint, voterAddr.String(), akiiDenom)
	s.Require().NoError(err, "failed to get balance for %s after second vote", voterAddr.String())
	s.Require().True(balanceAfterSecondVote.Amount.LT(balanceAfterFirstVote.Amount), "new balance should be less than the previous balance after second vote")
}

// testFeeder tests the feeder address functionality
// This test will check if the feeder address can be set and used correctly
func (s *IntegrationTestSuite) testFeeder() {
	// Check if the oracle parameters are set correctly
	s.checkAndUpdateOracleParams()

	// Get the validatorA
	validatorA := s.chainA.validators[0]
	voterAddr, _ := validatorA.keyInfo.GetAddress()
	validatorAddress := sdk.ValAddress(voterAddr).String()

	// Try to vote with a new address
	otherVoter := s.chainA.genesisAccounts[3]
	otherVoterAddr, _ := otherVoter.keyInfo.GetAddress()

	// Try to vote, but should fail with unauthorized voter error
	s.execAggregateVote(
		s.chainA,
		0,
		"1000akii",
		validatorAddress,
		otherVoterAddr.String(),
		kiichainHomePath,
		Fee.String(),
		s.execValidationWithError(s.chainA, 1, "unauthorized voter"),
	)

	// Now set the feeder address for validatorB
	s.T().Logf("Setting feeder address for validator %s to %s", validatorAddress, otherVoterAddr.String())
	s.execSetFeeder(
		s.chainA,
		0,
		otherVoterAddr.String(),
		voterAddr.String(),
		kiichainHomePath,
		Fee.String(),
		nil,
	)

	// Now the otherVoter should be able to vote
	s.T().Logf("Voting with feeder address %s", otherVoterAddr.String())
	s.execAggregateVote(
		s.chainA,
		0,
		"1000akii",
		validatorAddress,
		otherVoterAddr.String(),
		kiichainHomePath,
		Fee.String(),
		nil,
	)
}

// testSlash tests the slashing functionality of the oracle module
// This checks the slash window and the successful slash of a validator
func (s *IntegrationTestSuite) testSlash() {
	// Take the chain endpoint
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	// Get the first validator information
	validatorA := s.chainA.validators[0]
	voterAddr, _ := validatorA.keyInfo.GetAddress()
	validatorAddress := sdk.ValAddress(voterAddr).String()

	// Get the penalty counter for the validator
	queryPenaltyCounter, err := queryPenaltyCounter(chainEndpoint, validatorAddress)
	s.Require().NoError(err, "failed to query penalty counter for validator %s", validatorAddress)

	// Both the success and the abstain votes should be above zero
	s.Require().Greater(queryPenaltyCounter.VotePenaltyCounter.SuccessCount, uint64(0), "success penalty counter should be greater than zero")
	s.Require().Greater(queryPenaltyCounter.VotePenaltyCounter.AbstainCount, uint64(0), "abstain penalty counter should be greater than zero")
}

// checkAndUpdateOracleParams checks if the oracle parameters are set correctly and updates them if necessary
func (s *IntegrationTestSuite) checkAndUpdateOracleParams() {
	// Get the chain endpoint
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	// Query the oracle parameters
	params, err := queryOracleParameters(chainEndpoint)
	s.Require().NoError(err, "failed to query oracle parameters")

	// Check if the parameters are set under the expected values
	expectedWhitelist := []string{"akii"}
	if len(params.Params.Whitelist) == 1 && params.Params.Whitelist[0].Name == expectedWhitelist[0] {
		s.T().Log("Oracle parameters are already set correctly, skipping update")
		return
	}

	// If not, write a proposal to update the oracle parameters
	s.T().Log("Oracle parameters are not set correctly, writing proposal to update them")

	// Get the chain and update its information
	c := s.chainA
	s.writeOracleParamChangeProposal(c)
	proposalCounter++
	submitGovFlags := []string{configFile(proposalUpdateRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	// Get the validator address as the submitter
	validatorA := s.chainA.validators[0]
	validatorAAddr, _ := validatorA.keyInfo.GetAddress()

	// Submit the proposal to update the oracle parameters
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "oracle.MsgUpdateParams", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	// Log that the param was updated
	s.T().Logf("Oracle parameters updated successfully on chain %s", c.id)

	// Wait for the proposal to be processed and the parameters to be updated
	s.Require().Eventually(
		func() bool {
			s.T().Logf("After ParamUpdate proposal")

			params, err := queryOracleParameters(chainEndpoint)
			s.Require().NoError(err, "failed to query oracle parameters after proposal")
			s.Require().Equal(len(params.Params.Whitelist), 1, "expected one whitelist entry after proposal")
			s.Require().Equal(params.Params.Whitelist[0].Name, "akii", "expected whitelist entry to be 'akii' after proposal")

			return true
		},
		15*time.Second,
		5*time.Second,
	)

	// Get the current height
	currentHeight := s.getLatestBlockHeight(c, 0)
	next := ((currentHeight/BlocksPerPeriod)+1)*BlocksPerPeriod + 1

	// Wait for the height, but log first with the current height
	s.T().Logf("Waiting for the next block height %d on chain %s. Current block %d", next, c.id, currentHeight)
	s.waitUntilPassedHeight(s.chainA, 0, next)
}

// writeOracleParamChangeProposal writes a proposal to update the oracle parameters
func (s *IntegrationTestSuite) writeOracleParamChangeProposal(c *chain) {
	template := `
	{
		"messages": [
			{
				"@type": "/kiichain.oracle.v1beta1.MsgUpdateParams",
				"authority": "%s",
				"params": {
					"vote_period": "%d",
					"vote_threshold": "0.667000000000000000",
					"reward_band": "0.020000000000000000",
					"whitelist": [
						{
							"name": "akii"
						}
					],
					"slash_fraction": "0.050000000000000000",
					"slash_window": "%d",
					"min_valid_per_window": "0.050000000000000000",
					"lookback_duration": "3600"
				}
			}
		],
		"metadata": "ipfs://CID",
		"deposit": "100akii",
		"title": "Param Update",
		"summary": "Param Update",
		"expedited": false
	}`
	propMsgBody := fmt.Sprintf(
		template,
		govAuthority,
		BlocksPerPeriod,
		SlashWindow,
	)

	// Write the proposal to a file
	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalUpdateRateLimitAtomFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

// execAggregateVote executes an aggregate vote transaction on the oracle module
func (s *IntegrationTestSuite) execAggregateVote(c *chain, valIdx int, vote, validator, senderAddr, home, gasPrices string, validation func([]byte, []byte) bool) { //nolint:unparam
	// Build the context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Build the send command to the kiichaind binary
	s.T().Logf("Executing kiichaind tx oracle aggregate vote %s", c.id)
	kiichaindCommand := []string{
		kiichaindBinary,
		txCommand,
		oracletypes.ModuleName,
		"aggregate-vote",
		vote,
		validator,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, senderAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, gasPrices),
		"--gas=300000",
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	// Execute the command
	s.executeKiichainTxCommand(ctx, c, kiichaindCommand, valIdx, validation)
	// Log the result
	s.T().Logf("Executed kiichaind tx oracle aggregate vote %s successfully", c.id)
}

// execSetFeeder executes a set feeder transaction on the oracle module
func (s *IntegrationTestSuite) execSetFeeder(c *chain, valIdx int, feeder, senderAddr, home, gasPrices string, validation func([]byte, []byte) bool) {
	// Build the context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Build the send command to the kiichaind binary
	s.T().Logf("Executing kiichaind tx oracle set feeder %s", c.id)
	kiichaindCommand := []string{
		kiichaindBinary,
		txCommand,
		oracletypes.ModuleName,
		"set-feeder",
		feeder,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, senderAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, gasPrices),
		"--gas=300000",
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	// Execute the command
	s.executeKiichainTxCommand(ctx, c, kiichaindCommand, valIdx, validation)
	// Log the result
	s.T().Logf("Executed kiichaind tx oracle set feeder %s successfully", c.id)
}

// queryOracleParameters queries the oracle parameters from the given endpoint
func queryOracleParameters(endpoint string) (oracletypes.QueryParamsResponse, error) {
	// Create a new codec for unmarshalling
	var res oracletypes.QueryParamsResponse

	// Make the HTTP GET request to the endpoint
	body, err := httpGet(fmt.Sprintf("%s/kiichain/oracle/v1beta1/params", endpoint))
	if err != nil {
		return oracletypes.QueryParamsResponse{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	// Unmarshal the JSON response into the response struct
	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return oracletypes.QueryParamsResponse{}, err
	}
	return res, nil
}

// queryPenaltyCounter queries the penalty counter for a specific validator
func queryPenaltyCounter(endpoint string, validatorAddress string) (oracletypes.QueryVotePenaltyCounterResponse, error) {
	// Create a new codec for unmarshalling
	var res oracletypes.QueryVotePenaltyCounterResponse

	// Make the HTTP GET request to the endpoint
	body, err := httpGet(fmt.Sprintf("%s/kiichain/oracle/v1beta1/validators/%s/vote_penalty_counter", endpoint, validatorAddress))
	if err != nil {
		return oracletypes.QueryVotePenaltyCounterResponse{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	// Unmarshal the JSON response into the response struct
	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return oracletypes.QueryVotePenaltyCounterResponse{}, err
	}
	return res, nil
}
