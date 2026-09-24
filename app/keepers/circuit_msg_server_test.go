package keepers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type mockCircuitBreaker struct {
	denied map[string]bool
	err    error
}

func (m mockCircuitBreaker) IsAllowed(_ context.Context, typeURL string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return !m.denied[typeURL], nil
}

func denyAll(msgs ...sdk.Msg) mockCircuitBreaker {
	denied := make(map[string]bool, len(msgs))
	for _, msg := range msgs {
		denied[sdk.MsgTypeURL(msg)] = true
	}
	return mockCircuitBreaker{denied: denied}
}

type mockStakingMsgServer struct {
	stakingtypes.UnimplementedMsgServer
	called string
}

func (m *mockStakingMsgServer) CreateValidator(context.Context, *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	m.called = "CreateValidator"
	return &stakingtypes.MsgCreateValidatorResponse{}, nil
}
func (m *mockStakingMsgServer) EditValidator(context.Context, *stakingtypes.MsgEditValidator) (*stakingtypes.MsgEditValidatorResponse, error) {
	m.called = "EditValidator"
	return &stakingtypes.MsgEditValidatorResponse{}, nil
}
func (m *mockStakingMsgServer) Delegate(context.Context, *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error) {
	m.called = "Delegate"
	return &stakingtypes.MsgDelegateResponse{}, nil
}
func (m *mockStakingMsgServer) BeginRedelegate(context.Context, *stakingtypes.MsgBeginRedelegate) (*stakingtypes.MsgBeginRedelegateResponse, error) {
	m.called = "BeginRedelegate"
	return &stakingtypes.MsgBeginRedelegateResponse{}, nil
}
func (m *mockStakingMsgServer) Undelegate(context.Context, *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error) {
	m.called = "Undelegate"
	return &stakingtypes.MsgUndelegateResponse{}, nil
}
func (m *mockStakingMsgServer) CancelUnbondingDelegation(context.Context, *stakingtypes.MsgCancelUnbondingDelegation) (*stakingtypes.MsgCancelUnbondingDelegationResponse, error) {
	m.called = "CancelUnbondingDelegation"
	return &stakingtypes.MsgCancelUnbondingDelegationResponse{}, nil
}
func (m *mockStakingMsgServer) UpdateParams(context.Context, *stakingtypes.MsgUpdateParams) (*stakingtypes.MsgUpdateParamsResponse, error) {
	m.called = "UpdateParams"
	return &stakingtypes.MsgUpdateParamsResponse{}, nil
}

type mockDistrMsgServer struct {
	distrtypes.UnimplementedMsgServer
	called string
}

func (m *mockDistrMsgServer) SetWithdrawAddress(context.Context, *distrtypes.MsgSetWithdrawAddress) (*distrtypes.MsgSetWithdrawAddressResponse, error) {
	m.called = "SetWithdrawAddress"
	return &distrtypes.MsgSetWithdrawAddressResponse{}, nil
}
func (m *mockDistrMsgServer) WithdrawDelegatorReward(context.Context, *distrtypes.MsgWithdrawDelegatorReward) (*distrtypes.MsgWithdrawDelegatorRewardResponse, error) {
	m.called = "WithdrawDelegatorReward"
	return &distrtypes.MsgWithdrawDelegatorRewardResponse{}, nil
}
func (m *mockDistrMsgServer) WithdrawValidatorCommission(context.Context, *distrtypes.MsgWithdrawValidatorCommission) (*distrtypes.MsgWithdrawValidatorCommissionResponse, error) {
	m.called = "WithdrawValidatorCommission"
	return &distrtypes.MsgWithdrawValidatorCommissionResponse{}, nil
}
func (m *mockDistrMsgServer) FundCommunityPool(context.Context, *distrtypes.MsgFundCommunityPool) (*distrtypes.MsgFundCommunityPoolResponse, error) {
	m.called = "FundCommunityPool"
	return &distrtypes.MsgFundCommunityPoolResponse{}, nil
}
func (m *mockDistrMsgServer) UpdateParams(context.Context, *distrtypes.MsgUpdateParams) (*distrtypes.MsgUpdateParamsResponse, error) {
	m.called = "UpdateParams"
	return &distrtypes.MsgUpdateParamsResponse{}, nil
}
func (m *mockDistrMsgServer) CommunityPoolSpend(context.Context, *distrtypes.MsgCommunityPoolSpend) (*distrtypes.MsgCommunityPoolSpendResponse, error) {
	m.called = "CommunityPoolSpend"
	return &distrtypes.MsgCommunityPoolSpendResponse{}, nil
}
func (m *mockDistrMsgServer) DepositValidatorRewardsPool(context.Context, *distrtypes.MsgDepositValidatorRewardsPool) (*distrtypes.MsgDepositValidatorRewardsPoolResponse, error) {
	m.called = "DepositValidatorRewardsPool"
	return &distrtypes.MsgDepositValidatorRewardsPoolResponse{}, nil
}

type mockGovMsgServer struct {
	govv1.UnimplementedMsgServer
	called string
}

func (m *mockGovMsgServer) SubmitProposal(context.Context, *govv1.MsgSubmitProposal) (*govv1.MsgSubmitProposalResponse, error) {
	m.called = "SubmitProposal"
	return &govv1.MsgSubmitProposalResponse{}, nil
}
func (m *mockGovMsgServer) ExecLegacyContent(context.Context, *govv1.MsgExecLegacyContent) (*govv1.MsgExecLegacyContentResponse, error) {
	m.called = "ExecLegacyContent"
	return &govv1.MsgExecLegacyContentResponse{}, nil
}
func (m *mockGovMsgServer) Vote(context.Context, *govv1.MsgVote) (*govv1.MsgVoteResponse, error) {
	m.called = "Vote"
	return &govv1.MsgVoteResponse{}, nil
}
func (m *mockGovMsgServer) VoteWeighted(context.Context, *govv1.MsgVoteWeighted) (*govv1.MsgVoteWeightedResponse, error) {
	m.called = "VoteWeighted"
	return &govv1.MsgVoteWeightedResponse{}, nil
}
func (m *mockGovMsgServer) Deposit(context.Context, *govv1.MsgDeposit) (*govv1.MsgDepositResponse, error) {
	m.called = "Deposit"
	return &govv1.MsgDepositResponse{}, nil
}
func (m *mockGovMsgServer) UpdateParams(context.Context, *govv1.MsgUpdateParams) (*govv1.MsgUpdateParamsResponse, error) {
	m.called = "UpdateParams"
	return &govv1.MsgUpdateParamsResponse{}, nil
}
func (m *mockGovMsgServer) CancelProposal(context.Context, *govv1.MsgCancelProposal) (*govv1.MsgCancelProposalResponse, error) {
	m.called = "CancelProposal"
	return &govv1.MsgCancelProposalResponse{}, nil
}

type mockSlashingMsgServer struct {
	slashingtypes.UnimplementedMsgServer
	called string
}

func (m *mockSlashingMsgServer) Unjail(context.Context, *slashingtypes.MsgUnjail) (*slashingtypes.MsgUnjailResponse, error) {
	m.called = "Unjail"
	return &slashingtypes.MsgUnjailResponse{}, nil
}
func (m *mockSlashingMsgServer) UpdateParams(context.Context, *slashingtypes.MsgUpdateParams) (*slashingtypes.MsgUpdateParamsResponse, error) {
	m.called = "UpdateParams"
	return &slashingtypes.MsgUpdateParamsResponse{}, nil
}

func TestEnsureMsgAllowed(t *testing.T) {
	msg := &stakingtypes.MsgDelegate{}
	require.NoError(t, ensureMsgAllowed(context.Background(), mockCircuitBreaker{}, msg))

	err := ensureMsgAllowed(context.Background(), mockCircuitBreaker{err: errors.New("boom")}, msg)
	require.EqualError(t, err, "boom")

	err = ensureMsgAllowed(context.Background(), denyAll(msg), msg)
	require.ErrorContains(t, err, "circuit breaker disables execution")
}

func TestCircuitStakingMsgServer(t *testing.T) {
	inner := &mockStakingMsgServer{}
	allow := newCircuitStakingMsgServer(inner, mockCircuitBreaker{})
	deny := newCircuitStakingMsgServer(inner, denyAll(
		&stakingtypes.MsgCreateValidator{},
		&stakingtypes.MsgEditValidator{},
		&stakingtypes.MsgDelegate{},
		&stakingtypes.MsgBeginRedelegate{},
		&stakingtypes.MsgUndelegate{},
		&stakingtypes.MsgCancelUnbondingDelegation{},
		&stakingtypes.MsgUpdateParams{},
	))

	cases := []struct {
		name string
		run  func(stakingtypes.MsgServer) error
	}{
		{"CreateValidator", func(s stakingtypes.MsgServer) error {
			_, err := s.CreateValidator(context.Background(), &stakingtypes.MsgCreateValidator{})
			return err
		}},
		{"EditValidator", func(s stakingtypes.MsgServer) error {
			_, err := s.EditValidator(context.Background(), &stakingtypes.MsgEditValidator{})
			return err
		}},
		{"Delegate", func(s stakingtypes.MsgServer) error {
			_, err := s.Delegate(context.Background(), &stakingtypes.MsgDelegate{})
			return err
		}},
		{"BeginRedelegate", func(s stakingtypes.MsgServer) error {
			_, err := s.BeginRedelegate(context.Background(), &stakingtypes.MsgBeginRedelegate{})
			return err
		}},
		{"Undelegate", func(s stakingtypes.MsgServer) error {
			_, err := s.Undelegate(context.Background(), &stakingtypes.MsgUndelegate{})
			return err
		}},
		{"CancelUnbondingDelegation", func(s stakingtypes.MsgServer) error {
			_, err := s.CancelUnbondingDelegation(context.Background(), &stakingtypes.MsgCancelUnbondingDelegation{})
			return err
		}},
		{"UpdateParams", func(s stakingtypes.MsgServer) error {
			_, err := s.UpdateParams(context.Background(), &stakingtypes.MsgUpdateParams{})
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/allow", func(t *testing.T) {
			inner.called = ""
			require.NoError(t, tc.run(allow))
			require.Equal(t, tc.name, inner.called)
		})
		t.Run(tc.name+"/deny", func(t *testing.T) {
			inner.called = ""
			require.Error(t, tc.run(deny))
			require.Empty(t, inner.called)
		})
	}
}

func TestCircuitDistrMsgServer(t *testing.T) {
	inner := &mockDistrMsgServer{}
	allow := newCircuitDistrMsgServer(inner, mockCircuitBreaker{})
	deny := newCircuitDistrMsgServer(inner, denyAll(
		&distrtypes.MsgSetWithdrawAddress{},
		&distrtypes.MsgWithdrawDelegatorReward{},
		&distrtypes.MsgWithdrawValidatorCommission{},
		&distrtypes.MsgFundCommunityPool{},
		&distrtypes.MsgUpdateParams{},
		&distrtypes.MsgCommunityPoolSpend{},
		&distrtypes.MsgDepositValidatorRewardsPool{},
	))

	cases := []struct {
		name string
		run  func(distrtypes.MsgServer) error
	}{
		{"SetWithdrawAddress", func(s distrtypes.MsgServer) error {
			_, err := s.SetWithdrawAddress(context.Background(), &distrtypes.MsgSetWithdrawAddress{})
			return err
		}},
		{"WithdrawDelegatorReward", func(s distrtypes.MsgServer) error {
			_, err := s.WithdrawDelegatorReward(context.Background(), &distrtypes.MsgWithdrawDelegatorReward{})
			return err
		}},
		{"WithdrawValidatorCommission", func(s distrtypes.MsgServer) error {
			_, err := s.WithdrawValidatorCommission(context.Background(), &distrtypes.MsgWithdrawValidatorCommission{})
			return err
		}},
		{"FundCommunityPool", func(s distrtypes.MsgServer) error {
			_, err := s.FundCommunityPool(context.Background(), &distrtypes.MsgFundCommunityPool{})
			return err
		}},
		{"UpdateParams", func(s distrtypes.MsgServer) error {
			_, err := s.UpdateParams(context.Background(), &distrtypes.MsgUpdateParams{})
			return err
		}},
		{"CommunityPoolSpend", func(s distrtypes.MsgServer) error {
			_, err := s.CommunityPoolSpend(context.Background(), &distrtypes.MsgCommunityPoolSpend{})
			return err
		}},
		{"DepositValidatorRewardsPool", func(s distrtypes.MsgServer) error {
			_, err := s.DepositValidatorRewardsPool(context.Background(), &distrtypes.MsgDepositValidatorRewardsPool{})
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/allow", func(t *testing.T) {
			inner.called = ""
			require.NoError(t, tc.run(allow))
			require.Equal(t, tc.name, inner.called)
		})
		t.Run(tc.name+"/deny", func(t *testing.T) {
			inner.called = ""
			require.Error(t, tc.run(deny))
			require.Empty(t, inner.called)
		})
	}
}

func TestCircuitGovMsgServer(t *testing.T) {
	inner := &mockGovMsgServer{}
	allow := newCircuitGovMsgServer(inner, mockCircuitBreaker{})
	deny := newCircuitGovMsgServer(inner, denyAll(
		&govv1.MsgSubmitProposal{},
		&govv1.MsgExecLegacyContent{},
		&govv1.MsgVote{},
		&govv1.MsgVoteWeighted{},
		&govv1.MsgDeposit{},
		&govv1.MsgUpdateParams{},
		&govv1.MsgCancelProposal{},
	))

	cases := []struct {
		name string
		run  func(govv1.MsgServer) error
	}{
		{"SubmitProposal", func(s govv1.MsgServer) error {
			_, err := s.SubmitProposal(context.Background(), &govv1.MsgSubmitProposal{})
			return err
		}},
		{"ExecLegacyContent", func(s govv1.MsgServer) error {
			_, err := s.ExecLegacyContent(context.Background(), &govv1.MsgExecLegacyContent{})
			return err
		}},
		{"Vote", func(s govv1.MsgServer) error {
			_, err := s.Vote(context.Background(), &govv1.MsgVote{})
			return err
		}},
		{"VoteWeighted", func(s govv1.MsgServer) error {
			_, err := s.VoteWeighted(context.Background(), &govv1.MsgVoteWeighted{})
			return err
		}},
		{"Deposit", func(s govv1.MsgServer) error {
			_, err := s.Deposit(context.Background(), &govv1.MsgDeposit{})
			return err
		}},
		{"UpdateParams", func(s govv1.MsgServer) error {
			_, err := s.UpdateParams(context.Background(), &govv1.MsgUpdateParams{})
			return err
		}},
		{"CancelProposal", func(s govv1.MsgServer) error {
			_, err := s.CancelProposal(context.Background(), &govv1.MsgCancelProposal{})
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/allow", func(t *testing.T) {
			inner.called = ""
			require.NoError(t, tc.run(allow))
			require.Equal(t, tc.name, inner.called)
		})
		t.Run(tc.name+"/deny", func(t *testing.T) {
			inner.called = ""
			require.Error(t, tc.run(deny))
			require.Empty(t, inner.called)
		})
	}
}

func TestCircuitSlashingMsgServer(t *testing.T) {
	inner := &mockSlashingMsgServer{}
	allow := newCircuitSlashingMsgServer(inner, mockCircuitBreaker{})
	deny := newCircuitSlashingMsgServer(inner, denyAll(
		&slashingtypes.MsgUnjail{},
		&slashingtypes.MsgUpdateParams{},
	))

	cases := []struct {
		name string
		run  func(slashingtypes.MsgServer) error
	}{
		{"Unjail", func(s slashingtypes.MsgServer) error {
			_, err := s.Unjail(context.Background(), &slashingtypes.MsgUnjail{})
			return err
		}},
		{"UpdateParams", func(s slashingtypes.MsgServer) error {
			_, err := s.UpdateParams(context.Background(), &slashingtypes.MsgUpdateParams{})
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/allow", func(t *testing.T) {
			inner.called = ""
			require.NoError(t, tc.run(allow))
			require.Equal(t, tc.name, inner.called)
		})
		t.Run(tc.name+"/deny", func(t *testing.T) {
			inner.called = ""
			require.Error(t, tc.run(deny))
			require.Empty(t, inner.called)
		})
	}
}
