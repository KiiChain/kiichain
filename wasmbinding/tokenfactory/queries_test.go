package tokenfactory_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v4/app/apptesting"
	"github.com/kiichain/kiichain/v4/wasmbinding/helpers"
	wasmbinding "github.com/kiichain/kiichain/v4/wasmbinding/tokenfactory"
)

// TestFullDenom tests the GetFullDenom function of the token factory
func TestFullDenom(t *testing.T) {
	actor := apptesting.RandomAccountAddress()

	specs := map[string]struct {
		addr         string
		subdenom     string
		expFullDenom string
		expErr       bool
	}{
		"valid address": {
			addr:         actor.String(),
			subdenom:     "subDenom1",
			expFullDenom: fmt.Sprintf("factory/%s/subDenom1", actor.String()),
		},
		"empty address": {
			addr:     "",
			subdenom: "subDenom1",
			expErr:   true,
		},
		"invalid address": {
			addr:     "invalid",
			subdenom: "subDenom1",
			expErr:   true,
		},
		"empty sub-denom": {
			addr:         actor.String(),
			subdenom:     "",
			expFullDenom: fmt.Sprintf("factory/%s/", actor.String()),
		},
		"valid sub-denom (contains underscore)": {
			addr:         actor.String(),
			subdenom:     "sub_denom",
			expFullDenom: fmt.Sprintf("factory/%s/sub_denom", actor.String()),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotFullDenom, gotErr := wasmbinding.GetFullDenom(spec.addr, spec.subdenom)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expFullDenom, gotFullDenom, "exp %s but got %s", spec.expFullDenom, gotFullDenom)
		})
	}
}

// TestDenomAdmin tests the GetTokenfactoryDenomAdmin function of the token factory
func TestDenomAdmin(t *testing.T) {
	addr := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, addr)

	// set token creation fee to zero to make testing easier
	tfParams := app.TokenFactoryKeeper.GetParams(ctx)
	tfParams.DenomCreationFee = sdk.NewCoins()
	if err := app.TokenFactoryKeeper.SetParams(ctx, tfParams); err != nil {
		t.Fatal(err)
	}

	// create a subdenom via the token factory
	admin := sdk.AccAddress([]byte("addr1_______________"))
	tfDenom, err := app.TokenFactoryKeeper.CreateDenom(ctx, admin.String(), "subdenom")
	require.NoError(t, err)
	require.NotEmpty(t, tfDenom)

	queryPlugin := wasmbinding.NewQueryPlugin(app.BankKeeper, &app.TokenFactoryKeeper)

	testCases := []struct {
		name        string
		denom       string
		expectErr   bool
		expectAdmin string
	}{
		{
			name:        "valid token factory denom",
			denom:       tfDenom,
			expectAdmin: admin.String(),
		},
		{
			name:        "invalid token factory denom",
			denom:       "uosmo",
			expectErr:   false,
			expectAdmin: "",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			resp, err := queryPlugin.GetTokenfactoryDenomAdmin(ctx, tc.denom)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tc.expectAdmin, resp.Admin)
			}
		})
	}
}

// TestGetTokenfactoryDenomsByCreator tests the GetTokenfactoryDenomsByCreator function with address validation
func TestGetTokenfactoryDenomsByCreator(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)
	queryPlugin := wasmbinding.NewQueryPlugin(
		app.BankKeeper,
		&app.TokenFactoryKeeper,
	)

	// set token creation fee to zero to make testing easier
	tfParams := app.TokenFactoryKeeper.GetParams(ctx)
	tfParams.DenomCreationFee = sdk.NewCoins()
	if err := app.TokenFactoryKeeper.SetParams(ctx, tfParams); err != nil {
		t.Fatal(err)
	}

	// Create some denoms for testing
	_, err := app.TokenFactoryKeeper.CreateDenom(ctx, actor.String(), "token1")
	require.NoError(t, err)
	_, err = app.TokenFactoryKeeper.CreateDenom(ctx, actor.String(), "token2")
	require.NoError(t, err)

	testCases := []struct {
		name          string
		creator       string
		expectErr     bool
		expectedCount int
	}{
		{
			name:          "valid creator address with denoms",
			creator:       actor.String(),
			expectErr:     false,
			expectedCount: 2,
		},
		{
			name:          "valid creator address with no denoms",
			creator:       apptesting.RandomAccountAddress().String(),
			expectErr:     false,
			expectedCount: 0,
		},
		{
			name:      "invalid creator address",
			creator:   "invalid_address",
			expectErr: true,
		},
		{
			name:      "empty creator address",
			creator:   "",
			expectErr: true,
		},
		{
			name:      "malformed bech32 address",
			creator:   "kii1invalid!!!",
			expectErr: true,
		},
		{
			name:      "valid bech32 but wrong chain prefix",
			creator:   "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := queryPlugin.GetTokenfactoryDenomsByCreator(ctx, tc.creator)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "invalid creator address")
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Len(t, resp.Denoms, tc.expectedCount)
				// Verify actual denoms content for creator with denoms
				if tc.expectedCount > 0 && tc.creator == actor.String() {
					expectedDenoms := []string{
						fmt.Sprintf("factory/%s/token1", actor.String()),
						fmt.Sprintf("factory/%s/token2", actor.String()),
					}
					require.ElementsMatch(t, expectedDenoms, resp.Denoms)
				}
			}
		})
	}
}
