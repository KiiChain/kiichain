package keepers

import (
	"context"
	"fmt"

	circuitante "cosmossdk.io/x/circuit/ante"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ensureMsgAllowed rejects msgs that the circuit breaker has disabled.
// Precompiles call MsgServers directly and skip BaseApp's MsgServiceRouter,
// so this check must be applied at the MsgServer boundary they use.
func ensureMsgAllowed(ctx context.Context, breaker circuitante.CircuitBreaker, msg sdk.Msg) error {
	allowed, err := breaker.IsAllowed(ctx, sdk.MsgTypeURL(msg))
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("circuit breaker disables execution of this message: %s", sdk.MsgTypeURL(msg))
	}
	return nil
}

type circuitStakingMsgServer struct {
	stakingtypes.MsgServer
	breaker circuitante.CircuitBreaker
}

func newCircuitStakingMsgServer(inner stakingtypes.MsgServer, breaker circuitante.CircuitBreaker) stakingtypes.MsgServer {
	return &circuitStakingMsgServer{MsgServer: inner, breaker: breaker}
}

func (s *circuitStakingMsgServer) CreateValidator(ctx context.Context, msg *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.CreateValidator(ctx, msg)
}

func (s *circuitStakingMsgServer) EditValidator(ctx context.Context, msg *stakingtypes.MsgEditValidator) (*stakingtypes.MsgEditValidatorResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.EditValidator(ctx, msg)
}

func (s *circuitStakingMsgServer) Delegate(ctx context.Context, msg *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.Delegate(ctx, msg)
}

func (s *circuitStakingMsgServer) BeginRedelegate(ctx context.Context, msg *stakingtypes.MsgBeginRedelegate) (*stakingtypes.MsgBeginRedelegateResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.BeginRedelegate(ctx, msg)
}

func (s *circuitStakingMsgServer) Undelegate(ctx context.Context, msg *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.Undelegate(ctx, msg)
}

func (s *circuitStakingMsgServer) CancelUnbondingDelegation(ctx context.Context, msg *stakingtypes.MsgCancelUnbondingDelegation) (*stakingtypes.MsgCancelUnbondingDelegationResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.CancelUnbondingDelegation(ctx, msg)
}

func (s *circuitStakingMsgServer) UpdateParams(ctx context.Context, msg *stakingtypes.MsgUpdateParams) (*stakingtypes.MsgUpdateParamsResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.UpdateParams(ctx, msg)
}

type circuitDistrMsgServer struct {
	distrtypes.MsgServer
	breaker circuitante.CircuitBreaker
}

func newCircuitDistrMsgServer(inner distrtypes.MsgServer, breaker circuitante.CircuitBreaker) distrtypes.MsgServer {
	return &circuitDistrMsgServer{MsgServer: inner, breaker: breaker}
}

func (s *circuitDistrMsgServer) SetWithdrawAddress(ctx context.Context, msg *distrtypes.MsgSetWithdrawAddress) (*distrtypes.MsgSetWithdrawAddressResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.SetWithdrawAddress(ctx, msg)
}

func (s *circuitDistrMsgServer) WithdrawDelegatorReward(ctx context.Context, msg *distrtypes.MsgWithdrawDelegatorReward) (*distrtypes.MsgWithdrawDelegatorRewardResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.WithdrawDelegatorReward(ctx, msg)
}

func (s *circuitDistrMsgServer) WithdrawValidatorCommission(ctx context.Context, msg *distrtypes.MsgWithdrawValidatorCommission) (*distrtypes.MsgWithdrawValidatorCommissionResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.WithdrawValidatorCommission(ctx, msg)
}

func (s *circuitDistrMsgServer) FundCommunityPool(ctx context.Context, msg *distrtypes.MsgFundCommunityPool) (*distrtypes.MsgFundCommunityPoolResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.FundCommunityPool(ctx, msg)
}

func (s *circuitDistrMsgServer) UpdateParams(ctx context.Context, msg *distrtypes.MsgUpdateParams) (*distrtypes.MsgUpdateParamsResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.UpdateParams(ctx, msg)
}

func (s *circuitDistrMsgServer) CommunityPoolSpend(ctx context.Context, msg *distrtypes.MsgCommunityPoolSpend) (*distrtypes.MsgCommunityPoolSpendResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.CommunityPoolSpend(ctx, msg)
}

func (s *circuitDistrMsgServer) DepositValidatorRewardsPool(ctx context.Context, msg *distrtypes.MsgDepositValidatorRewardsPool) (*distrtypes.MsgDepositValidatorRewardsPoolResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.DepositValidatorRewardsPool(ctx, msg)
}

type circuitGovMsgServer struct {
	govv1.MsgServer
	breaker circuitante.CircuitBreaker
}

func newCircuitGovMsgServer(inner govv1.MsgServer, breaker circuitante.CircuitBreaker) govv1.MsgServer {
	return &circuitGovMsgServer{MsgServer: inner, breaker: breaker}
}

func (s *circuitGovMsgServer) SubmitProposal(ctx context.Context, msg *govv1.MsgSubmitProposal) (*govv1.MsgSubmitProposalResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.SubmitProposal(ctx, msg)
}

func (s *circuitGovMsgServer) ExecLegacyContent(ctx context.Context, msg *govv1.MsgExecLegacyContent) (*govv1.MsgExecLegacyContentResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.ExecLegacyContent(ctx, msg)
}

func (s *circuitGovMsgServer) Vote(ctx context.Context, msg *govv1.MsgVote) (*govv1.MsgVoteResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.Vote(ctx, msg)
}

func (s *circuitGovMsgServer) VoteWeighted(ctx context.Context, msg *govv1.MsgVoteWeighted) (*govv1.MsgVoteWeightedResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.VoteWeighted(ctx, msg)
}

func (s *circuitGovMsgServer) Deposit(ctx context.Context, msg *govv1.MsgDeposit) (*govv1.MsgDepositResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.Deposit(ctx, msg)
}

func (s *circuitGovMsgServer) UpdateParams(ctx context.Context, msg *govv1.MsgUpdateParams) (*govv1.MsgUpdateParamsResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.UpdateParams(ctx, msg)
}

func (s *circuitGovMsgServer) CancelProposal(ctx context.Context, msg *govv1.MsgCancelProposal) (*govv1.MsgCancelProposalResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.CancelProposal(ctx, msg)
}

type circuitSlashingMsgServer struct {
	slashingtypes.MsgServer
	breaker circuitante.CircuitBreaker
}

func newCircuitSlashingMsgServer(inner slashingtypes.MsgServer, breaker circuitante.CircuitBreaker) slashingtypes.MsgServer {
	return &circuitSlashingMsgServer{MsgServer: inner, breaker: breaker}
}

func (s *circuitSlashingMsgServer) Unjail(ctx context.Context, msg *slashingtypes.MsgUnjail) (*slashingtypes.MsgUnjailResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.Unjail(ctx, msg)
}

func (s *circuitSlashingMsgServer) UpdateParams(ctx context.Context, msg *slashingtypes.MsgUpdateParams) (*slashingtypes.MsgUpdateParamsResponse, error) {
	if err := ensureMsgAllowed(ctx, s.breaker, msg); err != nil {
		return nil, err
	}
	return s.MsgServer.UpdateParams(ctx, msg)
}
