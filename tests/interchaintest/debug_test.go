package interchaintest

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestDebugChain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
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

	// Wait for chain to start properly
	time.Sleep(10 * time.Second)

	// Test basic functionality - just check if we can query chain info
	t.Run("query chain info", func(t *testing.T) {
		t.Logf("Chain ID: %s", chain.Config().ChainID)
		t.Logf("RPC Address: %s", chain.GetRPCAddress())
		
		// Simple query that should work
		height, err := chain.Height(ctx)
		require.NoError(t, err)
		t.Logf("Current height: %d", height)
		require.Greater(t, height, uint64(0))
		
		t.Logf("Basic chain queries are working correctly")
	})
}