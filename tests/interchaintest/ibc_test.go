package interchaintest

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	interchaintestrelayer "github.com/cosmos/interchaintest/v10/relayer"
	"github.com/cosmos/interchaintest/v10/testreporter"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

const (
	ibcPath = "ibc-path"
)

func TestIBCBasic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)
	client, network := interchaintest.DockerSetup(t)

	cs := &DefaultChainSpec
	cs.ModifyGenesis = cosmos.ModifyGenesis([]cosmos.GenesisKV{cosmos.NewGenesisKV("app_state.ratelimit.blacklisted_denoms", []string{})})

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		cs,
		&SecondDefaultChainSpec,
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	chainA, chainB := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	// Relayer Factory
	r := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel)),
		interchaintestrelayer.CustomDockerImage(RelayerRepo, RelayerVersion, "100:1000"),
		interchaintestrelayer.StartupFlags("--processor", "events", "--block-history", "200"),
	).Build(t, client, network)

	ic := interchaintest.NewInterchain().
		AddChain(chainA).
		AddChain(chainB).
		AddRelayer(r, "relayer")

	ic = ic.AddLink(interchaintest.InterchainLink{
		Chain1:  chainA,
		Chain2:  chainB,
		Relayer: r,
		Path:    ibcPath,
	})

	// Build interchain
	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
	}))

	require.NoError(t, testutil.WaitForBlocks(ctx, 5, chainA))

	// Fund faucet
	err = FundFaucet(ctx, chainA)
	require.NoError(t, err)
	err = FundFaucet(ctx, chainB)
	require.NoError(t, err)

	// Create and Fund User Wallets
	fundAmount := math.NewInt(10_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", fundAmount, chainA, chainB)
	userA := users[0]
	userB := users[1]

	userAInitial, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
	require.NoError(t, err)
	require.True(t, userAInitial.Equal(fundAmount))

	// Get Channel ID
	aInfo, err := r.GetChannels(ctx, eRep, chainA.Config().ChainID)
	require.NoError(t, err)
	aChannelID, err := getTransferChannel(aInfo)
	require.NoError(t, err)

	bInfo, err := r.GetChannels(ctx, eRep, chainB.Config().ChainID)
	require.NoError(t, err)
	bChannelID, err := getTransferChannel(bInfo)
	require.NoError(t, err)

	// Send Transaction
	amountToSend := math.NewInt(1_000_000)
	dstAddress := userB.FormattedAddress()
	transfer := ibc.WalletAmount{
		Address: dstAddress,
		Denom:   chainA.Config().Denom,
		Amount:  amountToSend,
	}

	_, err = chainA.SendIBCTransfer(ctx, aChannelID, userA.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	// relay MsgRecvPacket to chainB, then MsgAcknowledgement back to chainA
	require.NoError(t, r.Flush(ctx, eRep, ibcPath, aChannelID))

	// test source wallet has decreased funds
	expectedBal := userAInitial.Sub(amountToSend)
	aNewBal, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
	require.NoError(t, err)
	require.True(t, aNewBal.Equal(expectedBal))

	// Trace IBC Denom
	srcDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", bChannelID, chainA.Config().Denom))
	dstIbcDenom := srcDenomTrace.IBCDenom()

	// Test destination wallet has increased funds
	bNewBal, err := chainB.GetBalance(ctx, userB.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.True(t, bNewBal.Equal(amountToSend))
}
