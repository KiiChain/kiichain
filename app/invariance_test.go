package app_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	app "github.com/kiichain/kiichain3/app"
	"github.com/stretchr/testify/require"
)

func TestLightInvarianceChecks(t *testing.T) {
	tm := time.Now().UTC()
	valPub := secp256k1.GenPrivKey().PubKey()
	accounts := []sdk.AccAddress{
		sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
		sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
	}
	ukiiCoin := func(i int64) sdk.Coin { return sdk.NewCoin("ukii", sdk.NewInt(i)) }
	ukiiCoins := func(i int64) sdk.Coins { return sdk.NewCoins(ukiiCoin(i)) }
	for i, tt := range []struct {
		preUkii    []int64
		preWei     []int64
		preSupply  int64
		postUkii   []int64
		postWei    []int64
		postSupply int64
		success    bool
	}{
		{
			preUkii:    []int64{0, 0},
			preWei:     []int64{0, 0},
			preSupply:  5,
			postUkii:   []int64{1, 2},
			postWei:    []int64{0, 0},
			postSupply: 8,
			success:    true,
		},
		{
			preUkii:    []int64{2, 1},
			preWei:     []int64{0, 0},
			preSupply:  3,
			postUkii:   []int64{0, 0},
			postWei:    []int64{0, 0},
			postSupply: 0,
			success:    true,
		},
		{
			preUkii:    []int64{1, 0},
			preWei:     []int64{0, 0},
			preSupply:  10,
			postUkii:   []int64{0, 1},
			postWei:    []int64{0, 0},
			postSupply: 10,
			success:    true,
		},
		{
			preUkii:    []int64{1, 0},
			preWei:     []int64{0, 0},
			preSupply:  10,
			postUkii:   []int64{0, 0},
			postWei:    []int64{500_000_000_000, 500_000_000_000},
			postSupply: 10,
			success:    true,
		},
		{
			preUkii:    []int64{0, 0},
			preWei:     []int64{500_000_000_000, 500_000_000_000},
			preSupply:  10,
			postUkii:   []int64{1, 0},
			postWei:    []int64{0, 0},
			postSupply: 10,
			success:    true,
		},
		{
			preUkii:    []int64{0, 0},
			preWei:     []int64{1, 2},
			preSupply:  10,
			postUkii:   []int64{0, 0},
			postWei:    []int64{2, 1},
			postSupply: 10,
			success:    true,
		},
		{
			preUkii:    []int64{1, 0},
			preWei:     []int64{0, 0},
			preSupply:  10,
			postUkii:   []int64{1, 1},
			postWei:    []int64{0, 0},
			postSupply: 10,
			success:    false,
		},
		{
			preUkii:    []int64{1, 0},
			preWei:     []int64{0, 0},
			preSupply:  10,
			postUkii:   []int64{0, 0},
			postWei:    []int64{0, 0},
			postSupply: 10,
			success:    false,
		},
		{
			preUkii:    []int64{1, 0},
			preWei:     []int64{0, 0},
			preSupply:  10,
			postUkii:   []int64{0, 1},
			postWei:    []int64{500_000_000_000, 500_000_000_000},
			postSupply: 10,
			success:    false,
		},
		{
			preUkii:    []int64{1, 0},
			preWei:     []int64{500_000_000_000, 500_000_000_000},
			preSupply:  10,
			postUkii:   []int64{0, 1},
			postWei:    []int64{0, 0},
			postSupply: 10,
			success:    false,
		},
		{
			preUkii:    []int64{0, 0},
			preWei:     []int64{1, 2},
			preSupply:  10,
			postUkii:   []int64{0, 0},
			postWei:    []int64{2, 2},
			postSupply: 10,
			success:    false,
		},
		{
			preUkii:    []int64{0, 0},
			preWei:     []int64{1, 2},
			preSupply:  10,
			postUkii:   []int64{0, 0},
			postWei:    []int64{1, 1},
			postSupply: 10,
			success:    false,
		},
	} {
		fmt.Printf("Running test %d\n", i)
		testWrapper := app.NewTestWrapperWithSc(t, tm, valPub, false)
		a, ctx := testWrapper.App, testWrapper.Ctx
		for i := range tt.preUkii {
			if tt.preUkii[i] > 0 {
				a.BankKeeper.AddCoins(ctx, accounts[i], ukiiCoins(tt.preUkii[i]), false)
			}
			if tt.preWei[i] > 0 {
				a.BankKeeper.AddWei(ctx, accounts[i], sdk.NewInt(tt.preWei[i]))
			}
		}
		if tt.preSupply > 0 {
			a.BankKeeper.SetSupply(ctx, ukiiCoin(tt.preSupply))
		}
		a.SetDeliverStateToCommit()
		a.WriteState()
		a.GetWorkingHash() // flush to sc
		for i := range tt.postUkii {
			ukiiDiff := tt.postUkii[i] - tt.preUkii[i]
			if ukiiDiff > 0 {
				a.BankKeeper.AddCoins(ctx, accounts[i], ukiiCoins(ukiiDiff), false)
			} else if ukiiDiff < 0 {
				a.BankKeeper.SubUnlockedCoins(ctx, accounts[i], ukiiCoins(-ukiiDiff), false)
			}

			weiDiff := tt.postWei[i] - tt.preWei[i]
			if weiDiff > 0 {
				a.BankKeeper.AddWei(ctx, accounts[i], sdk.NewInt(weiDiff))
			} else if weiDiff < 0 {
				a.BankKeeper.SubWei(ctx, accounts[i], sdk.NewInt(-weiDiff))
			}
		}
		a.BankKeeper.SetSupply(ctx, ukiiCoin(tt.postSupply))
		a.SetDeliverStateToCommit()
		f := func() { a.LightInvarianceChecks(a.WriteState(), app.LightInvarianceConfig{SupplyEnabled: true}) }
		if tt.success {
			require.NotPanics(t, f)
		} else {
			require.Panics(t, f)
		}
		safeClose(a)
	}
}

// TODO: remove once snapshot manager can be closed gracefully in tests
func safeClose(a *app.App) {
	defer func() {
		_ = recover()
	}()
	a.Close()
}
