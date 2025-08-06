package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v4/x/rewards/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Transaction commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewFundPoolCmd(),
		NewUpdateParamsCmd(),
		NewChangeScheduleCmd(),
	)

	return cmd
}

// NewFundPoolCmd implements the fund-pool tx command.
func NewFundPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fund-pool [amount]",
		Short: "Fund the rewards community pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}

			msg := types.NewMsgFundPool(clientCtx.GetFromAddress(), amount)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewUpdateParamsCmd implements the update-params tx command.
func NewUpdateParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params [params-json]",
		Short: "Update module parameters (gov proposal)",
		Long: `Update module parameters through a governance proposal. Example:
$ %s tx gov submit-proposal update-rewards-params <path/to/params.json> --from mykey
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var params types.Params
			if err := clientCtx.Codec.UnmarshalJSON([]byte(args[0]), &params); err != nil {
				return fmt.Errorf("failed to parse params: %w", err)
			}

			msg := types.NewMsgUpdateParams(clientCtx.GetFromAddress().String(), params)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewChangeScheduleCmd implements the change-schedule tx command.
func NewChangeScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-schedule [schedule-json]",
		Short: "Change schedule information (gov proposal)",
		Long: `Change schedule information through a governance proposal. Example:
$ %s tx gov submit-proposal change-schedule <path/to/schedule.json> --from mykey
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var schedule types.ReleaseSchedule
			if err := clientCtx.Codec.UnmarshalJSON([]byte(args[0]), &schedule); err != nil {
				return fmt.Errorf("failed to parse params: %w", err)
			}

			msg := types.NewMsgChangeSchedule(clientCtx.GetFromAddress().String(), schedule)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
