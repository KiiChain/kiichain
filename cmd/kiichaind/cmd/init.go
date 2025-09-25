package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/spf13/cobra"

	tmconfig "github.com/cometbft/cometbft/config"
	tmnode "github.com/cometbft/cometbft/node"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmtypes "github.com/cometbft/cometbft/types"
	tmp2p "github.com/cometbft/cometbft/p2p"

	"github.com/kiichain/kiichain/v5/app/params"
	"github.com/kiichain/kiichain/v5/types/module"
)

// printInfo holds the node initialization info
type printInfo struct {
	Moniker  string          `json:"moniker"`
	ChainID  string          `json:"chain_id"`
	NodeID   string          `json:"node_id"`
	AppState json.RawMessage `json:"app_state"`
}

// newPrintInfo returns a formatted info struct
func newPrintInfo(moniker, chainID, nodeID string, appState json.RawMessage) printInfo {
	return printInfo{
		Moniker:  moniker,
		ChainID:  chainID,
		NodeID:   nodeID,
		AppState: appState,
	}
}

// initCmd initializes all files for a node
func initCmd(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize node configuration (validator, p2p, genesis)",
		Long: `The init command sets up configuration files for your node, including
the private validator key, P2P networking, and the genesis file.`,
		Example: `
  # Init node with a custom moniker and chain-id
  kiichaind init mynode --chain-id kiichain-1

  # Init node using a recovery seed phrase
  kiichaind init mynode --recover

  # Init and overwrite existing genesis.json
  kiichaind init mynode --overwrite
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initializes the client context
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Load Tendermint config
			config := cfg

			// Ensure root dir exists
			tmos.EnsureDir(config.RootDir, 0700)

			// Load or generate node key
			nodeKey, err := tmp2p.LoadOrGenNodeKey(filepath.Join(config.RootDir, "config", "node_key.json"))
			if err != nil {
				return fmt.Errorf("failed to load/generate node key: %w", err)
			}

			// Get node ID
			nodeID := string(nodeKey.ID())

			// Placeholder app state
			appState := json.RawMessage(`{}`)

			// Print chain info
			toPrint := newPrintInfo(config.Moniker, clientCtx.ChainID, nodeID, appState)

			// Set Tendermint configs
			params.SetTendermintConfigs(config)

			// Save config.toml
			tmconfig.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

			// Check output format
			output, _ := cmd.Flags().GetString("output")
			return displayInfo(toPrint, output)
		},
	}

	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(genutilcli.FlagDefaultBondDenom, "", "genesis file default denomination, if left blank default value is 'akii'")
	cmd.Flags().Int64(flags.FlagInitHeight, 1, "specify the initial block height at genesis")

	// Add output format flag
	cmd.Flags().String("output", "json", "Output format (json|text)")

	return cmd
}

// displayInfo formats and prints the initialization result
func displayInfo(info printInfo, format string) error {
	switch format {
	case "text":
		fmt.Printf("\n✅ Node initialized successfully!\n")
		fmt.Printf("🔹 Moniker: %s\n", info.Moniker)
		fmt.Printf("🔹 Chain ID: %s\n", info.ChainID)
		fmt.Printf("🔹 Node ID: %s\n", info.NodeID)
		return nil
	default:
		out, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(os.Stderr, "%s\n", out)
		return err
	}
}
