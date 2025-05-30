package oracle_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/kiichain/kiichain/v1/x/oracle"
	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
)

func TestVoteAloneHandle(t *testing.T) {
	type testCase struct {
		name          string
		expectedError bool
		tx            sdk.Tx
	}

	// these are the test messages
	testOracleMsg := types.MsgAggregateExchangeRateVote{}
	testNoOracleMsg := banktypes.MsgSend{}
	testNoOracleMsg2 := banktypes.MsgSend{}

	// register oracle vote alone decorator
	decorator := oracle.NewVoteAloneDecorator()
	anteHandler := sdk.ChainAnteDecorators(decorator)

	testCases := []testCase{
		// ante handle wil continue this
		{
			name:          "only oracle votes",
			expectedError: false,
			tx:            oracle.NewTestTx([]sdk.Msg{&testOracleMsg}),
		},

		// ante handle will ignore this message
		{
			name:          "only non-oracle votes",
			expectedError: false,
			tx:            oracle.NewTestTx([]sdk.Msg{&testNoOracleMsg, &testNoOracleMsg2}),
		},

		// ante handle will return an error because the oracle message can not be with other messages
		{
			name:          "mixed messages",
			expectedError: true,
			tx:            oracle.NewTestTx([]sdk.Msg{&testOracleMsg, &testNoOracleMsg, &testNoOracleMsg2}),
		},
	}

	// Iterate cases
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			// Create context
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)

			// execute ante handler
			_, err := anteHandler(ctx, test.tx, false)
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSpammingPreventionHandle(t *testing.T) {
	// Prepare env
	input, _ := oracle.SetUp(t)
	ctx := input.Ctx
	oracleKeeper := input.OracleKeeper

	// Create test exchange rate
	randomAExchangeRate := math.LegacyNewDec(1700)
	exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom

	voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[0], keeper.ValAddrs[0])
	invalidVoteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[5], keeper.ValAddrs[2]) // addr 3 has not been delegated by val 2

	// Register anti spamming decorator
	spammingDecorator := oracle.NewSpammingPreventionDecorator(oracleKeeper)
	anteHandler := sdk.ChainAnteDecorators(spammingDecorator)

	// should skip the ante handler because the context is set as Recheck
	recheckCtx := ctx.WithIsReCheckTx(true)
	_, err := anteHandler(recheckCtx, oracle.NewTestTx([]sdk.Msg{voteMsg}), false)
	require.NoError(t, err)

	// should return error because the feeder is not valid
	checkCtx := ctx.WithIsCheckTx(true)
	require.True(t, checkCtx.IsCheckTx()) // validate ctx has IsCheckTx active
	_, err = anteHandler(checkCtx, oracle.NewTestTx([]sdk.Msg{invalidVoteMsg}), false)
	require.Error(t, err)

	// should return error because of feeder malform
	malformFeeder := voteMsg
	malformFeeder.Feeder = "kiifeeder"
	_, err = anteHandler(checkCtx, oracle.NewTestTx([]sdk.Msg{malformFeeder}), false)
	require.Error(t, err)

	// should return error because of validator malform
	malformVal := voteMsg
	malformVal.Feeder = "kiivalidator"
	_, err = anteHandler(checkCtx, oracle.NewTestTx([]sdk.Msg{malformVal}), false)
	require.Error(t, err)

	// should fail, no exchange rate on message
	exRate, _ := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{}, keeper.ValAddrs[0])
	err = oracleKeeper.AggregateExchangeRateVote.Set(ctx, keeper.ValAddrs[0], exRate)
	require.NoError(t, err)
	_, err = anteHandler(checkCtx, oracle.NewTestTx([]sdk.Msg{voteMsg}), false)
	require.Error(t, err)
}
