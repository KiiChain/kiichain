package interchaintest

import (
	"context"
	"testing"

	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestTokenFactory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx := context.Background()
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)
	client, network := interchaintest.DockerSetup(t)

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		&DefaultChainSpec,
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	chain := chains[0].(*cosmos.CosmosChain)

	ic := interchaintest.NewInterchain().AddChain(chain)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
	}))

	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", GenesisFundsAmount, chain, chain)
	user := users[0]
	user2 := users[1]

	uaddr := user.FormattedAddress()
	uaddr2 := user2.FormattedAddress()

	node := chain.GetNode()

	tfDenom, _, err := node.TokenFactoryCreateDenom(ctx, user, "ictestdenom", 5_000_000)
	t.Log("TF Denom: ", tfDenom)
	require.NoError(t, err)

	t.Run("Mint TF Denom to user", func(t *testing.T) {
		node.TokenFactoryMintDenom(ctx, user.FormattedAddress(), tfDenom, 100)
		if balance, err := chain.GetBalance(ctx, uaddr, tfDenom); err != nil {
			t.Fatal(err)
		} else if balance.Int64() != 100 {
			t.Fatal("balance not 100")
		}
	})

	t.Run("Mint TF Denom to another user", func(t *testing.T) {
		node.TokenFactoryMintDenomTo(ctx, user.FormattedAddress(), tfDenom, 70, user2.FormattedAddress())
		if balance, err := chain.GetBalance(ctx, uaddr2, tfDenom); err != nil {
			t.Fatal(err)
		} else if balance.Int64() != 70 {
			t.Fatal("balance not 70")
		}
	})

	t.Run("Change admin to uaddr2", func(t *testing.T) {
		_, err = node.TokenFactoryChangeAdmin(ctx, user.KeyName(), tfDenom, uaddr2)
		require.NoError(t, err)
	})

	t.Run("Validate new admin address", func(t *testing.T) {
		res, err := chain.TokenFactoryQueryAdmin(ctx, tfDenom)
		require.NoError(t, err)
		require.EqualValues(t, res.AuthorityMetadata.Admin, uaddr2, "admin not uaddr2. Did not properly transfer.")
	})

	t.Cleanup(func() {
		_ = ic.Close()
	})
}