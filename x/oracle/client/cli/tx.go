package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/kiichain/kiichain3/x/oracle/types"
	"github.com/spf13/cobra"
)

func GetTxCmd() *cobra.Command {
	// Register the oracle transactions subcommands
	oracleTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Oracle transmition subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Add Tx commands

	return oracleTxCmd
}
