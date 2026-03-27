package keepers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	kiichain "github.com/kiichain/kiichain/v7/app"
	helpers "github.com/kiichain/kiichain/v7/app/helpers"
)

// createBondedValidator funds a new validator address, creates the validator,
// runs EndBlocker to move it to the bonded set, and registers signing info.
// Returns the val address and consensus address
func createBondedValidator(t *testing.T, app *kiichain.KiichainApp, ctx sdk.Context, stakedAmount math.Int) (sdk.ValAddress, sdk.ConsAddress) {
	t.Helper()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	valAddr := sdk.ValAddress(pubKey.Address())
	consAddr := sdk.ConsAddress(pubKey.Address())

	bondDenom, err := app.StakingKeeper.BondDenom(ctx)
	require.NoError(t, err)

	selfBond := sdk.NewCoins(sdk.NewCoin(bondDenom, stakedAmount))
	require.NoError(t, app.BankKeeper.MintCoins(ctx, "tokenfactory", selfBond))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, "tokenfactory", sdk.AccAddress(valAddr), selfBond))

	msgServer := stakingkeeper.NewMsgServerImpl(app.StakingKeeper)
	createMsg, err := stakingtypes.NewMsgCreateValidator(
		valAddr.String(),
		pubKey,
		sdk.NewCoin(bondDenom, stakedAmount),
		stakingtypes.Description{Moniker: "test-validator"},
		stakingtypes.NewCommissionRates(
			math.LegacyNewDecWithPrec(5, 2),
			math.LegacyNewDecWithPrec(20, 2),
			math.LegacyNewDecWithPrec(1, 2),
		),
		math.OneInt(),
	)
	require.NoError(t, err)
	_, err = msgServer.CreateValidator(ctx, createMsg)
	require.NoError(t, err)

	_, err = app.StakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr, ctx.BlockHeight(), 0, time.Unix(0, 0), false, 0,
	)
	require.NoError(t, app.SlashingKeeper.SetValidatorSigningInfo(ctx, consAddr, signingInfo))

	return valAddr, consAddr
}

// TestSlashHooks_SupplyUnchangedAfterSlash verifies that when
// the staking keeper slashes a validator, the total token supply is
// unchanged
func TestSlashHooks_SupplyUnchangedAfterSlash(t *testing.T) {
	app := helpers.Setup(t)
	ctx := app.NewUncachedContext(true, tmtypes.Header{
		Height:  1,
		ChainID: "testing",
		Time:    time.Now().UTC(),
	})

	bondDenom, err := app.StakingKeeper.BondDenom(ctx)
	require.NoError(t, err)

	stakedAmount := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	valAddr, consAddr := createBondedValidator(t, app, ctx, stakedAmount)

	supplyBefore := app.BankKeeper.GetSupply(ctx, bondDenom).Amount

	validator, err := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.NoError(t, err)
	powerReduction := app.StakingKeeper.PowerReduction(ctx)
	consensusPower := validator.GetConsensusPower(powerReduction)

	slashFactor := math.LegacyNewDecWithPrec(5, 2) // 5%
	slashed, err := app.StakingKeeper.Slash(ctx, consAddr, 0, consensusPower, slashFactor)
	require.NoError(t, err)
	require.True(t, slashed.IsPositive(), "expected a non-zero slash amount")

	// Supply must be unchanged: the burn is offset by the hook's mint
	require.Equal(t, supplyBefore, app.BankKeeper.GetSupply(ctx, bondDenom).Amount,
		"total supply must be unchanged after slash")

	// Community pool must have received exactly the slashed amount
	feePool, err := app.DistrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	communityPool := feePool.CommunityPool.AmountOf(bondDenom)
	require.Equal(t, math.LegacyNewDecFromInt(slashed), communityPool,
		"community pool must receive exactly the slash amount")
}

// TestSlashHooks_ZeroSlashFactor verifies no tokens move and supply is
// unchanged when the slash factor is zero
func TestSlashHooks_ZeroSlashFactor(t *testing.T) {
	app := helpers.Setup(t)
	ctx := app.NewUncachedContext(true, tmtypes.Header{
		Height:  1,
		ChainID: "testing",
		Time:    time.Now().UTC(),
	})

	bondDenom, err := app.StakingKeeper.BondDenom(ctx)
	require.NoError(t, err)

	stakedAmount := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	_, consAddr := createBondedValidator(t, app, ctx, stakedAmount)

	supplyBefore := app.BankKeeper.GetSupply(ctx, bondDenom).Amount

	validator, err := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.NoError(t, err)
	powerReduction := app.StakingKeeper.PowerReduction(ctx)
	consensusPower := validator.GetConsensusPower(powerReduction)

	slashed, err := app.StakingKeeper.Slash(ctx, consAddr, 0, consensusPower, math.LegacyZeroDec())
	require.NoError(t, err)
	require.True(t, slashed.IsZero(), "expected zero slash for zero factor")

	require.Equal(t, supplyBefore, app.BankKeeper.GetSupply(ctx, bondDenom).Amount,
		"supply must be unchanged when slash factor is zero")
}
