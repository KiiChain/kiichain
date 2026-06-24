package cosmos

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
	protov2 "google.golang.org/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	antetestutil "github.com/cosmos/cosmos-sdk/x/auth/ante/testutil"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// feeTxStub is a minimal sdk.FeeTx used to drive checkDeductFee without
// building a full transaction or bootstrapping the app.
type feeTxStub struct {
	fee        sdk.Coins
	gas        uint64
	feePayer   []byte
	feeGranter []byte
}

func (f feeTxStub) GetMsgs() []sdk.Msg                    { return nil }
func (f feeTxStub) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (f feeTxStub) GetGas() uint64                        { return f.gas }
func (f feeTxStub) GetFee() sdk.Coins                     { return f.fee }
func (f feeTxStub) FeePayer() []byte                      { return f.feePayer }
func (f feeTxStub) FeeGranter() []byte                    { return f.feeGranter }

// feeAbstractionStub is a stub of the fee abstraction keeper that returns a
// fixed converted fee (or error) regardless of input.
type feeAbstractionStub struct {
	converted sdk.Coins
	err       error
}

func (f feeAbstractionStub) ConvertNativeFee(_ sdk.Context, _ sdk.AccAddress, _ sdk.Coins) (sdk.Coins, error) {
	return f.converted, f.err
}

// TestCheckDeductFeeInternal exercises the post-conversion feegrant flow in
// checkDeductFee with fully mocked keepers. It runs without the "test" build
// tag (no app bootstrap), so the CI coverage job actually records it.
func TestCheckDeductFeeInternal(t *testing.T) {
	payer := sdk.AccAddress([]byte("payeraddress00000001"))
	granter := sdk.AccAddress([]byte("granteraddress000001"))

	nativeFee := sdk.NewCoins(sdk.NewInt64Coin("akii", 1000))
	convertedFee := sdk.NewCoins(sdk.NewInt64Coin("uconv", 500))

	testCases := []struct {
		name        string
		feeGranter  []byte
		converted   sdk.Coins
		convertErr  error
		grantErr    error
		deductErr   error
		errContains string
	}{
		{
			name:       "success - converts, consumes grant and deducts",
			feeGranter: granter,
			converted:  convertedFee,
		},
		{
			name:        "fail - conversion error is returned",
			feeGranter:  granter,
			convertErr:  errors.New("conversion boom"),
			errContains: "conversion boom",
		},
		{
			name:        "fail - feegrant denied is wrapped",
			feeGranter:  granter,
			converted:   convertedFee,
			grantErr:    errors.New("fee-grant not found"),
			errContains: "does not allow to pay fees",
		},
		{
			name:        "fail - deduction error is returned",
			feeGranter:  nil,
			converted:   convertedFee,
			deductErr:   errors.New("insufficient funds"),
			errContains: "insufficient funds",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			accountKeeper := antetestutil.NewMockAccountKeeper(ctrl)
			feegrantKeeper := antetestutil.NewMockFeegrantKeeper(ctrl)
			bankKeeper := authtestutil.NewMockBankKeeper(ctrl)

			deductFrom := payer
			if tc.feeGranter != nil {
				deductFrom = sdk.AccAddress(tc.feeGranter)
			}
			acc := authtypes.NewBaseAccountWithAddress(deductFrom)

			accountKeeper.EXPECT().GetModuleAddress(authtypes.FeeCollectorName).Return(sdk.AccAddress([]byte("feecollector00000001"))).AnyTimes()
			accountKeeper.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Return(acc).AnyTimes()
			feegrantKeeper.EXPECT().UseGrantedFees(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(tc.grantErr).AnyTimes()
			bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(tc.deductErr).AnyTimes()

			dfd := NewDeductFeeDecorator(
				accountKeeper,
				bankKeeper,
				feegrantKeeper,
				feeAbstractionStub{converted: tc.converted, err: tc.convertErr},
				func(sdk.Context, sdk.Tx) (sdk.Coins, int64, error) { return nil, 0, nil },
			)

			ctx := sdk.Context{}.WithEventManager(sdk.NewEventManager())
			tx := feeTxStub{fee: nativeFee, gas: 1_000_000, feePayer: payer, feeGranter: tc.feeGranter}

			err := dfd.checkDeductFee(ctx, tx, nativeFee)
			if tc.errContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
