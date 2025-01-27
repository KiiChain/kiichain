package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/kiichain/kiichain3/x/oracle/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd() *cobra.Command {
	// Register the oracle query subcommands
	oracleQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the oracle module",
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Add Query commands

	return oracleQueryCmd
}
