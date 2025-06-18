package e2e

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	tokenfactorytypes "github.com/kiichain/kiichain/v2/x/tokenfactory/types"
)

func (s *IntegrationTestSuite) testTokenFactory() {
	s.Run("create_token_mint_and_burn", func() {
		var (
			err           error
			valIdx        = 0
			c             = s.chainA
			chainEndpoint = fmt.Sprintf("http://%s", s.valResources[c.id][valIdx].GetHostPort("1317/tcp"))
		)

		// Define one admin and two other accounts
		admin, _ := c.genesisAccounts[1].keyInfo.GetAddress()
		bob, _ := c.genesisAccounts[2].keyInfo.GetAddress()
		charlie, _ := c.genesisAccounts[3].keyInfo.GetAddress()

		adminAddress := admin.String()
		bobAddress := bob.String()
		charlieAddress := charlie.String()

		// Get denom name
		newDenom := "upanda"
		fullDenom := builFullDenom(adminAddress, newDenom)

		// Setup amounts
		adminAmount := math.NewInt(1000000000)
		bobAmount := math.NewInt(5000)
		burnAmount := math.NewInt(2000)

		// Create denom
		s.createDenom(s.chainA, adminAddress, newDenom)
		// Add funds to alice and bob
		s.mintDenom(s.chainA, adminAddress, sdk.NewCoin(fullDenom, adminAmount))
		s.mintDenomTo(s.chainA, adminAddress, bobAddress, sdk.NewCoin(fullDenom, bobAmount))

		var initialAdminBalance,
			initialBobBalance,
			initialCharlieBalance,
			laterBobBalance sdk.Coin

		// Get balances of admin and other accounts
		s.Require().Eventually(
			func() bool {
				// Admin should have a bunch
				initialAdminBalance, err = getSpecificBalance(chainEndpoint, adminAddress, fullDenom)
				s.Require().NoError(err)
				s.Require().Equal(adminAmount, initialAdminBalance.Amount)

				// Bob should have some balance
				initialBobBalance, err = getSpecificBalance(chainEndpoint, bobAddress, fullDenom)
				s.Require().NoError(err)
				s.Require().Equal(bobAmount, initialBobBalance.Amount)

				// Charlie should have no balance
				initialCharlieBalance, err = getSpecificBalance(chainEndpoint, charlieAddress, fullDenom)
				s.Require().NoError(err)
				s.Require().Zero(initialCharlieBalance.Amount)

				return true
			},
			10*time.Second,
			5*time.Second,
		)

		// Burn from bob, expected to fail
		s.burnDenomFrom(c, adminAddress, bobAddress, sdk.NewCoin(fullDenom, burnAmount))

		// Verify change
		s.Require().Eventually(
			func() bool {
				// Bob should have same coin since it should fail
				laterBobBalance, err = getSpecificBalance(chainEndpoint, bobAddress, fullDenom)
				s.Require().NoError(err)
				s.Require().Equal(bobAmount, laterBobBalance.Amount)

				return true
			},
			10*time.Second,
			5*time.Second,
		)
	})
}

// createDenomAndMint uses tokenfactory module to create a specific denom under a given
// admin and mint a given amount of that new currency
func (s *IntegrationTestSuite) createDenom(c *chain, admin, denom string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	createDenomCmd := []string{
		kiichaindBinary,
		txCommand,
		tokenfactorytypes.ModuleName,
		"create-denom",
		denom,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, admin),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, "300000000000akii"),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "auto"),
		fmt.Sprintf("--%s=%s", flags.FlagGasAdjustment, "1.5"),
		// "--gas=250000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	s.T().Logf("Creating denom %s with admin %s on chain %s ", denom, admin, s.chainA.id)

	s.executeKiichainTxCommand(ctx, c, createDenomCmd, 0, s.defaultExecValidation(c, 0))
}

func (s *IntegrationTestSuite) mintDenom(c *chain, admin string, amount sdk.Coin) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	mintCmd := []string{
		kiichaindBinary,
		txCommand,
		tokenfactorytypes.ModuleName,
		"mint",
		amount.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, admin),
		fmt.Sprintf("--%s=%s", flags.FlagFees, standardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--gas=250000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.T().Logf("Minting %s to %s", amount.String(), admin)
	s.executeKiichainTxCommand(ctx, c, mintCmd, 0, s.defaultExecValidation(c, 0))
}

func (s *IntegrationTestSuite) mintDenomTo(c *chain, admin, receiver string, amount sdk.Coin) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	mintCmd := []string{
		kiichaindBinary,
		txCommand,
		tokenfactorytypes.ModuleName,
		"mint-to",
		receiver,
		amount.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, admin),
		fmt.Sprintf("--%s=%s", flags.FlagFees, standardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--gas=250000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.T().Logf("Minting %s to %s", amount.String(), admin)
	s.executeKiichainTxCommand(ctx, c, mintCmd, 0, s.defaultExecValidation(c, 0))
}

func (s *IntegrationTestSuite) burnDenomFrom(c *chain, admin, receiver string, amount sdk.Coin) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	mintCmd := []string{
		kiichaindBinary,
		txCommand,
		tokenfactorytypes.ModuleName,
		"burn-from",
		receiver,
		amount.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, admin),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagFees, standardFees.String()),
		"--gas=250000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.T().Logf("Burning %s from %s", amount.String(), receiver)
	// burnFrom is disabled and should fail
	s.executeKiichainTxCommand(ctx, c, mintCmd, 0, s.expectErrExecValidation(c, 0, true))
}

func builFullDenom(creator, denom string) string {
	return fmt.Sprintf("factory/%s/%s", creator, denom)
}
