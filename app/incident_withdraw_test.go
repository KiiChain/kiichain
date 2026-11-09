package kiichain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	kiichain "github.com/kiichain/kiichain/v7/app"
	"github.com/kiichain/kiichain/v7/app/blockedaddrs"
	kiihelpers "github.com/kiichain/kiichain/v7/app/helpers"
)

// slashSetup is a chain where a delegator redelegated from validator A to
// validator B and has pending rewards at B, with the incident restriction on
type slashSetup struct {
	app       *kiichain.KiichainApp
	height    int64
	blockTime time.Time
	srcCons   sdk.ConsAddress
	srcPower  int64
}

// setupRedelegationSlash builds the slash scenario. delegatorFor picks the
// delegator (given the funded genesis account) and withdrawFor its withdraw
// address, which is written directly to the store as pre-upgrade state.
func setupRedelegationSlash(
	t *testing.T,
	delegatorFor func(funder sdk.AccAddress) sdk.AccAddress,
	withdrawFor func(delegator sdk.AccAddress) sdk.AccAddress,
) slashSetup {
	t.Helper()

	pv := mock.NewPV()
	pk, err := pv.GetPubKey()
	require.NoError(t, err)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{tmtypes.NewValidator(pk, 1)})

	funderKey := secp256k1.GenPrivKey()
	funderAcc := authtypes.NewBaseAccount(funderKey.PubKey().Address().Bytes(), funderKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: funderAcc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100_000_000_000_000))),
	}
	app := kiihelpers.SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{funderAcc}, balance)
	_, err = app.Commit()
	require.NoError(t, err)

	height := app.LastBlockHeight() + 1
	blockTime := time.Now().UTC()
	ctx := app.NewUncachedContext(false, cmtproto.Header{Height: height, ChainID: app.ChainID(), Time: blockTime})
	funder := funderAcc.GetAddress()

	// Validator A is the genesis validator, create validator B
	srcVal := sdk.ValAddress(pk.Address())
	src, err := app.StakingKeeper.GetValidator(ctx, srcVal)
	require.NoError(t, err)

	operator := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	selfBond := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000_000_000))
	require.NoError(t, app.BankKeeper.SendCoins(ctx, funder, operator, sdk.NewCoins(selfBond)))
	dstVal := sdk.ValAddress(operator)
	createMsg, err := stakingtypes.NewMsgCreateValidator(
		dstVal.String(), ed25519.GenPrivKey().PubKey(), selfBond, stakingtypes.Description{Moniker: "B"},
		stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDecWithPrec(2, 1), math.LegacyNewDecWithPrec(1, 2)),
		math.OneInt(),
	)
	require.NoError(t, err)
	_, err = stakingkeeper.NewMsgServerImpl(app.StakingKeeper).CreateValidator(ctx, createMsg)
	require.NoError(t, err)

	// Delegate to A and redelegate to B in the same block
	delegator := delegatorFor(funder)
	amt := math.NewInt(10_000_000)
	if !delegator.Equals(funder) {
		require.NoError(t, app.BankKeeper.SendCoins(ctx, funder, delegator, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))))
	}
	require.NoError(t, app.DistrKeeper.SetDelegatorWithdrawAddr(ctx, delegator, withdrawFor(delegator)))
	shares, err := app.StakingKeeper.Delegate(ctx, delegator, amt, stakingtypes.Unbonded, src, true)
	require.NoError(t, err)
	_, err = app.StakingKeeper.BeginRedelegation(ctx, delegator, srcVal, dstVal, shares)
	require.NoError(t, err)

	// Rewards accrue at B
	reward := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000)))
	require.NoError(t, app.BankKeeper.SendCoinsFromAccountToModule(ctx, funder, distrtypes.ModuleName, reward))
	dstValidator, err := app.StakingKeeper.GetValidator(ctx, dstVal)
	require.NoError(t, err)
	require.NoError(t, app.DistrKeeper.AllocateTokensToValidator(ctx, dstValidator, sdk.NewDecCoinsFromCoins(reward...)))

	// The test genesis helper skips the slashing pubkey and signing info a live chain has
	srcValidator, err := app.StakingKeeper.GetValidator(ctx, srcVal)
	require.NoError(t, err)
	srcCons, err := srcValidator.GetConsAddr()
	require.NoError(t, err)
	srcPub, err := srcValidator.ConsPubKey()
	require.NoError(t, err)
	require.NoError(t, app.SlashingKeeper.AddPubkey(ctx, srcPub))
	require.NoError(t, app.SlashingKeeper.SetValidatorSigningInfo(ctx, srcCons,
		slashingtypes.NewValidatorSigningInfo(srcCons, 0, 0, time.Unix(0, 0).UTC(), false, 0)))

	// Same call the v7.4.0 upgrade handler makes on mainnet
	blockedaddrs.Enable(ctx, app.GetKey(banktypes.StoreKey))

	return slashSetup{
		app:       app,
		height:    height,
		blockTime: blockTime,
		srcCons:   srcCons,
		srcPower:  srcValidator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)),
	}
}

// finalizeWithDoubleSign commits the setup block, then finalizes the next block
// with double-sign evidence for validator A, which slashes the redelegation
func finalizeWithDoubleSign(t *testing.T, s slashSetup) error {
	t.Helper()
	_, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.height, Time: s.blockTime})
	require.NoError(t, err)
	_, err = s.app.Commit()
	require.NoError(t, err)

	_, err = s.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: s.height + 1,
		Time:   s.blockTime.Add(5 * time.Second),
		Misbehavior: []abci.Misbehavior{{
			Type:             abci.MisbehaviorType_DUPLICATE_VOTE,
			Validator:        abci.Validator{Address: s.srcCons, Power: s.srcPower},
			Height:           s.height,
			Time:             s.blockTime,
			TotalVotingPower: s.srcPower,
		}},
	})
	if err != nil {
		return err
	}
	_, err = s.app.Commit()
	require.NoError(t, err)

	ctx := s.app.NewUncachedContext(false, cmtproto.Header{Height: s.height + 2})
	require.True(t, s.app.SlashingKeeper.IsTombstoned(ctx, s.srcCons), "double-sign evidence must be processed")
	return nil
}

// TestIncidentAddressRejectedAsWithdrawAddress tests that an incident address
// cannot be set as a distribution withdraw address
func TestIncidentAddressRejectedAsWithdrawAddress(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t)
	delegator := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	for _, addr := range blockedaddrs.SortedAttackerAddresses() {
		err := app.DistrKeeper.SetWithdrawAddr(ctx, delegator, sdk.MustAccAddressFromBech32(addr))
		require.ErrorIs(t, err, sdkerrors.ErrUnauthorized, addr)
	}
}

// TestSlashWithIncidentRewardRecipientDoesNotHalt tests that a slash paying
// rewards to an incident address inside BeginBlocker does not fail the block
func TestSlashWithIncidentRewardRecipientDoesNotHalt(t *testing.T) {
	incident := sdk.MustAccAddressFromBech32(blockedaddrs.SortedAttackerAddresses()[0])

	testCases := []struct {
		name         string
		delegatorFor func(funder sdk.AccAddress) sdk.AccAddress
		withdrawFor  func(delegator sdk.AccAddress) sdk.AccAddress
	}{
		{
			name:         "regular withdraw address",
			delegatorFor: func(funder sdk.AccAddress) sdk.AccAddress { return funder },
			withdrawFor:  func(delegator sdk.AccAddress) sdk.AccAddress { return delegator },
		},
		{
			name:         "incident withdraw address set before the upgrade",
			delegatorFor: func(funder sdk.AccAddress) sdk.AccAddress { return funder },
			withdrawFor:  func(sdk.AccAddress) sdk.AccAddress { return incident },
		},
		{
			name:         "incident address as the delegator",
			delegatorFor: func(sdk.AccAddress) sdk.AccAddress { return incident },
			withdrawFor:  func(delegator sdk.AccAddress) sdk.AccAddress { return delegator },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := setupRedelegationSlash(t, tc.delegatorFor, tc.withdrawFor)
			require.NoError(t, finalizeWithDoubleSign(t, s))
		})
	}
}
