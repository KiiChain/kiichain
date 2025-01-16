package tools

import (
	migration "github.com/kiichain/kiichain3/tools/migration/cmd"
	scanner "github.com/kiichain/kiichain3/tools/tx-scanner/cmd"
	"github.com/spf13/cobra"
)

func ToolCmd() *cobra.Command {
	toolsCmd := &cobra.Command{
		Use:   "tools",
		Short: "A set of useful tools for kii chain",
	}
	toolsCmd.AddCommand(scanner.ScanCmd())
	toolsCmd.AddCommand(migration.MigrateCmd())
	toolsCmd.AddCommand(migration.VerifyMigrationCmd())
	return toolsCmd
}
