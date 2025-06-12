package tokenfactory_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v2/app/apptesting"
	"github.com/kiichain/kiichain/v2/wasmbinding/helpers"
	wasmbinding "github.com/kiichain/kiichain/v2/wasmbinding/tokenfactory"
	bindingtypes "github.com/kiichain/kiichain/v2/wasmbinding/tokenfactory/types"
	"github.com/kiichain/kiichain/v2/x/tokenfactory/types"
)

// TestCreateDenom tests the CreateDenom function
func TestCreateDenom(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	// Fund actor with 100 base denom creation fees
	actorAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	helpers.FundAccount(t, ctx, app, actor, actorAmount)

	specs := map[string]struct {
		createDenom *bindingtypes.CreateDenom
		expErr      bool
	}{
		"valid sub-denom": {
			createDenom: &bindingtypes.CreateDenom{
				Subdenom: "MOON",
			},
		},
		"empty sub-denom": {
			createDenom: &bindingtypes.CreateDenom{
				Subdenom: "",
			},
			expErr: false,
		},
		"invalid sub-denom": {
			createDenom: &bindingtypes.CreateDenom{
				Subdenom: "sub-denom_2",
			},
			expErr: false,
		},
		"null create denom": {
			createDenom: nil,
			expErr:      true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			_, gotErr := wasmbinding.PerformCreateDenom(&app.TokenFactoryKeeper, app.BankKeeper, ctx, actor, spec.createDenom)
			// then
			if spec.expErr {
				t.Logf("validate_msg_test got error: %v", gotErr)
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

// TestChangeAdmin tests the ChangeAdmin function
func TestChangeAdmin(t *testing.T) {
	const validDenom = "validdenom"

	tokenCreator := apptesting.RandomAccountAddress()

	specs := map[string]struct {
		actor       sdk.AccAddress
		changeAdmin *bindingtypes.ChangeAdmin

		expErrMsg string
	}{
		"valid": {
			changeAdmin: &bindingtypes.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: helpers.RandomBech32AccountAddress(),
			},
			actor: tokenCreator,
		},
		"typo in factory in denom name": {
			changeAdmin: &bindingtypes.ChangeAdmin{
				Denom:           fmt.Sprintf("facory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: helpers.RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: "denom prefix is incorrect. Is: facory.  Should be: factory: invalid denom",
		},
		"invalid address in denom": {
			changeAdmin: &bindingtypes.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", helpers.RandomBech32AccountAddress(), validDenom),
				NewAdminAddress: helpers.RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: "failed changing admin from message: unauthorized account",
		},
		"other denom name in 3 part name": {
			changeAdmin: &bindingtypes.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), "invalid denom"),
				NewAdminAddress: helpers.RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: fmt.Sprintf("invalid denom: factory/%s/invalid denom", tokenCreator.String()),
		},
		"empty denom": {
			changeAdmin: &bindingtypes.ChangeAdmin{
				Denom:           "",
				NewAdminAddress: helpers.RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: "invalid denom: ",
		},
		"empty address": {
			changeAdmin: &bindingtypes.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: "",
			},
			actor:     tokenCreator,
			expErrMsg: "address from bech32: empty address string is not allowed",
		},
		"creator is a different address": {
			changeAdmin: &bindingtypes.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: helpers.RandomBech32AccountAddress(),
			},
			actor:     apptesting.RandomAccountAddress(),
			expErrMsg: "failed changing admin from message: unauthorized account",
		},
		"change to the same address": {
			changeAdmin: &bindingtypes.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: tokenCreator.String(),
			},
			actor: tokenCreator,
		},
		"nil binding": {
			actor:     tokenCreator,
			expErrMsg: "invalid request: changeAdmin is nil - original request: ",
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// Setup
			app, ctx := helpers.SetupCustomApp(t, tokenCreator)

			// Fund actor with 100 base denom creation fees
			actorAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
			helpers.FundAccount(t, ctx, app, tokenCreator, actorAmount)

			_, err := wasmbinding.PerformCreateDenom(&app.TokenFactoryKeeper, app.BankKeeper, ctx, tokenCreator, &bindingtypes.CreateDenom{
				Subdenom: validDenom,
			})
			require.NoError(t, err)

			err = wasmbinding.ChangeAdmin(&app.TokenFactoryKeeper, ctx, spec.actor, spec.changeAdmin)
			if len(spec.expErrMsg) > 0 {
				require.Error(t, err)
				actualErrMsg := err.Error()
				require.Equal(t, spec.expErrMsg, actualErrMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}

// TestMint tests the minting of tokens
func TestMint(t *testing.T) {
	creator := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, creator)

	// Fund actor with 100 base denom creation fees
	tokenCreationFeeAmt := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	helpers.FundAccount(t, ctx, app, creator, tokenCreationFeeAmt)

	// Create denoms for valid mint tests
	validDenom := bindingtypes.CreateDenom{
		Subdenom: "MOON",
	}
	_, err := wasmbinding.PerformCreateDenom(&app.TokenFactoryKeeper, app.BankKeeper, ctx, creator, &validDenom)
	require.NoError(t, err)

	emptyDenom := bindingtypes.CreateDenom{
		Subdenom: "",
	}
	_, err = wasmbinding.PerformCreateDenom(&app.TokenFactoryKeeper, app.BankKeeper, ctx, creator, &emptyDenom)
	require.NoError(t, err)

	validDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), validDenom.Subdenom)
	emptyDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), emptyDenom.Subdenom)

	lucky := apptesting.RandomAccountAddress()

	// lucky was broke
	balances := app.BankKeeper.GetAllBalances(ctx, lucky)
	require.Empty(t, balances)

	amount, ok := sdkmath.NewIntFromString("8080")
	require.True(t, ok)

	specs := map[string]struct {
		mint   *bindingtypes.MintTokens
		expErr bool
	}{
		"valid mint": {
			mint: &bindingtypes.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
		},
		"empty sub-denom": {
			mint: &bindingtypes.MintTokens{
				Denom:         emptyDenomStr,
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
			expErr: false,
		},
		"nonexistent sub-denom": {
			mint: &bindingtypes.MintTokens{
				Denom:         fmt.Sprintf("factory/%s/%s", creator.String(), "SUN"),
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"invalid sub-denom": {
			mint: &bindingtypes.MintTokens{
				Denom:         "sub-denom_2",
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"zero amount": {
			mint: &bindingtypes.MintTokens{
				Denom:         validDenomStr,
				Amount:        sdkmath.ZeroInt(),
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"negative amount": {
			mint: &bindingtypes.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount.Neg(),
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"empty recipient": {
			mint: &bindingtypes.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount,
				MintToAddress: "",
			},
			expErr: true,
		},
		"invalid recipient": {
			mint: &bindingtypes.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount,
				MintToAddress: "invalid",
			},
			expErr: true,
		},
		"null mint": {
			mint:   nil,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotErr := wasmbinding.PerformMint(&app.TokenFactoryKeeper, app.BankKeeper, ctx, creator, spec.mint)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

// TestBurn tests the burning of tokens
func TestBurn(t *testing.T) {
	creator := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, creator)

	// Fund actor with 100 base denom creation fees
	tokenCreationFeeAmt := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	helpers.FundAccount(t, ctx, app, creator, tokenCreationFeeAmt)

	// Create denoms for valid burn tests
	validDenom := bindingtypes.CreateDenom{
		Subdenom: "MOON",
	}
	_, err := wasmbinding.PerformCreateDenom(&app.TokenFactoryKeeper, app.BankKeeper, ctx, creator, &validDenom)
	require.NoError(t, err)

	emptyDenom := bindingtypes.CreateDenom{
		Subdenom: "",
	}
	_, err = wasmbinding.PerformCreateDenom(&app.TokenFactoryKeeper, app.BankKeeper, ctx, creator, &emptyDenom)
	require.NoError(t, err)

	lucky := apptesting.RandomAccountAddress()

	// lucky was broke
	balances := app.BankKeeper.GetAllBalances(ctx, lucky)
	require.Empty(t, balances)

	validDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), validDenom.Subdenom)
	emptyDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), emptyDenom.Subdenom)
	mintAmount, ok := sdkmath.NewIntFromString("8080")
	require.True(t, ok)

	// Check if burn is enabled from non admins
	capabilities := app.TokenFactoryKeeper.GetEnabledCapabilities()
	burnFromEnabled := types.IsCapabilityEnabled(capabilities, types.EnableBurnFrom)

	specs := map[string]struct {
		burn   *bindingtypes.BurnTokens
		expErr bool
	}{
		"valid burn": {
			burn: &bindingtypes.BurnTokens{
				Denom:  validDenomStr,
				Amount: mintAmount,
			},
			expErr: false,
		},
		"non admin address": {
			burn: &bindingtypes.BurnTokens{
				Denom:           validDenomStr,
				Amount:          mintAmount,
				BurnFromAddress: lucky.String(),
			},
			expErr: !burnFromEnabled,
		},
		"empty sub-denom": {
			burn: &bindingtypes.BurnTokens{
				Denom:  emptyDenomStr,
				Amount: mintAmount,
			},
			expErr: false,
		},
		"invalid sub-denom": {
			burn: &bindingtypes.BurnTokens{
				Denom:  "sub-denom_2",
				Amount: mintAmount,
			},
			expErr: true,
		},
		"non-minted denom": {
			burn: &bindingtypes.BurnTokens{
				Denom:  fmt.Sprintf("factory/%s/%s", creator.String(), "SUN"),
				Amount: mintAmount,
			},
			expErr: true,
		},
		"zero amount": {
			burn: &bindingtypes.BurnTokens{
				Denom:  validDenomStr,
				Amount: sdkmath.ZeroInt(),
			},
			expErr: true,
		},
		"negative amount": {
			burn:   nil,
			expErr: true,
		},
		"null burn": {
			burn: &bindingtypes.BurnTokens{
				Denom:  validDenomStr,
				Amount: mintAmount.Neg(),
			},
			expErr: true,
		},
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// Mint valid denom str and empty denom string for burn test
			mintBinding := &bindingtypes.MintTokens{
				Denom:         validDenomStr,
				Amount:        mintAmount,
				MintToAddress: creator.String(),
			}
			err := wasmbinding.PerformMint(&app.TokenFactoryKeeper, app.BankKeeper, ctx, creator, mintBinding)
			require.NoError(t, err)

			emptyDenomMintBinding := &bindingtypes.MintTokens{
				Denom:         emptyDenomStr,
				Amount:        mintAmount,
				MintToAddress: creator.String(),
			}
			err = wasmbinding.PerformMint(&app.TokenFactoryKeeper, app.BankKeeper, ctx, creator, emptyDenomMintBinding)
			require.NoError(t, err)

			// when
			gotErr := wasmbinding.PerformBurn(&app.TokenFactoryKeeper, ctx, creator, spec.burn)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
