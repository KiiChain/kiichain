package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

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
	cmd.AddCommand()
	return cmd
}
