package interchaintest

import (
	"context"
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testreporter"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestBasicChain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)
	client, network := interchaintest.DockerSetup(t)

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		&DefaultChainSpec,
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	chain := chains[0].(*cosmos.CosmosChain)

	// Setup Interchain
	ic := interchaintest.NewInterchain().
		AddChain(chain)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	faucetAddr, err := chain.Validators[0].KeyBech32(ctx, interchaintest.FaucetAccountKeyName, "acc")
	require.NoError(t, err)

	// Fund the faucet from validator
	err = chain.Validators[0].BankSend(ctx, "validator", ibc.WalletAmount{
		Address: faucetAddr,
		Amount: sdkmath.NewIntFromBigInt(new(big.Int).Mul(
			big.NewInt(100_000_000_000), // 100B tokens
			new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		)),
		Denom: Denom,
	})
	require.NoError(t, err)

	// Wait for transaction to process
	err = testutil.WaitForBlocks(ctx, 2, chain)
	require.NoError(t, err)

	// Use amount that faucet can afford with zero gas fees
	amt := sdkmath.NewInt(50_000_000_000_000) // 50T akii
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", amt,
		chain,
	)
	user := users[0]

	t.Run("validate funding", func(t *testing.T) {
		t.Logf("Querying balance for user: %s", user.FormattedAddress())
		bal, err := chain.BankQueryBalance(ctx, user.FormattedAddress(), chain.Config().Denom)
		require.NoError(t, err)
		t.Logf("Expected: %s, Got: %s", amt.String(), bal.String())
		require.EqualValues(t, amt, bal)
	})
}
