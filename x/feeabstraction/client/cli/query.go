package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// GetQueryCmd returns the cli query commands
func GetQueryCmd() *cobra.Command {
	// Create the core cobra command
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Add all the commands and return the CMD
	cmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdQueryFeeTokens(),
	)
	return cmd
}

// GetCmdQueryParams implements the params query command.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current fee abstraction parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize the client
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Create a new query client
			queryClient := types.NewQueryClient(clientCtx)

			// Call the Params query
			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			// Print the response
			return clientCtx.PrintProto(&res.Params)
		},
	}
	// Add query flags to the command
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryFeeTokens implements the fee tokens query command.
func GetCmdQueryFeeTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee-tokens",
		Short: "Query the current fee tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize the client
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Create a new query client
			queryClient := types.NewQueryClient(clientCtx)

			// Call the FeeTokens query
			res, err := queryClient.FeeTokens(cmd.Context(), &types.QueryFeeTokensRequest{})
			if err != nil {
				return err
			}

			// Print the response
			return clientCtx.PrintProto(res.FeeTokens)
		},
	}
	// Add query flags to the command
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
