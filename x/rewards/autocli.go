package rewards

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kiichain/kiichain/v1/x/rewards/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current rewards parameters.",
				},
				{
					RpcMethod: "RewardReleaser",
					Use:       "releaser",
					Short:     "Query reward releaser current information",
					Example:   fmt.Sprintf("$ %s query rewards releaser ", version.AppName),
				},
				{
					RpcMethod: "RewardPool",
					Use:       "pool",
					Short:     "Query the amount of coins in the reward community pool",
					Example:   fmt.Sprintf(`$ %s query rewards pool`, version.AppName),
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: types.Msg_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "FundPool",
					Use:       "fund-pool [amount]",
					Short:     "Funds the reward community pool with the specified amount",
					Example:   fmt.Sprintf(`$ %s tx rewards fund-pool 100akii --from mykey`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "amount", Varargs: true},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod: "ExtendReward",
					Skip:      true, // skipped because authority gated
				},
			},
			EnhanceCustomCommand: true,
		},
	}
}
