package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	cfg "github.com/cometbft/cometbft/config"

	"github.com/cosmos/go-bip39"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/kiichain/kiichain/v5/app/params"
)

// printInfo represents node initialization output info
type printInfo struct {
	Moniker    string          `json:"moniker" yaml:"moniker"`
	ChainID    string          `json:"chain_id" yaml:"chain_id"`
	NodeID     string          `json:"node_id" yaml:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir" yaml:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message" yaml:"app_message"`
}

// newPrintInfo creates a new printInfo struct with provided parameters
func newPrintInfo(moniker, chainID, nodeID, genTxsDir string, appMessage json.RawMessage) printInfo {
	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}

// displayInfo formats and prints info to stdout
func displayInfo(info printInfo) error {
	out, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	_, err = os.Stdout.Write(out)
	return err
}

// initCmd initializes node validator, p2p, genesis, and app configuration files
func initCmd(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long:  `Initialize validator and node configuration files.`,
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			// Determine ChainID
			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			switch {
			case chainID != "":
			case clientCtx.ChainID != "":
				chainID = clientCtx.ChainID
			default:
				rand.Seed(time.Now().UnixNano())
				chainID = fmt.Sprintf("test_1234-%d", rand.Intn(1_000_000))
			}

			// Handle mnemonic recovery
			var mnemonic string
			recoverFlag, _ := cmd.Flags().GetBool(genutilcli.FlagRecover)
			if recoverFlag {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				value, err := input.GetString("Enter your bip39 mnemonic", inBuf)
				if err != nil {
					return err
				}
				mnemonic = value
				if !bip39.IsMnemonicValid(mnemonic) {
					return errors.New("invalid mnemonic")
				}
			}

			// Get initial height
			initHeight, _ := cmd.Flags().GetInt64(flags.FlagInitHeight)
			if initHeight < 1 {
				initHeight = 1
			}

			// Initialize validator files
			nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
			if err != nil {
				return err
			}

			config.Moniker = args[0]
			genFile := config.GenesisFile()
			overwrite, _ := cmd.Flags().GetBool(genutilcli.FlagOverwrite)
			defaultDenom, _ := cmd.Flags().GetString(genutilcli.FlagDefaultBondDenom)

			// Check if genesis file exists
			_, err = os.Stat(genFile)
			if !overwrite && !os.IsNotExist(err) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}

			// Override default denom if provided
			if defaultDenom != "" {
				sdk.DefaultBondDenom = defaultDenom
			}

			appGenState := mbm.DefaultGenesis(cdc)
			appState, err := json.MarshalIndent(appGenState, "", " ")
			if err != nil {
				return errorsmod.Wrap(err, "failed to marshal default genesis state")
			}

			appGenesis := &types.AppGenesis{}
			if _, err := os.Stat(genFile); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				appGenesis, err = types.AppGenesisFromFile(genFile)
				if err != nil {
					return errorsmod.Wrap(err, "failed to read genesis doc from file")
				}
			}

			appGenesis.AppName = version.AppName
			appGenesis.AppVersion = version.Version
			appGenesis.ChainID = chainID
			appGenesis.AppState = appState
			appGenesis.InitialHeight = initHeight
			appGenesis.Consensus = &types.ConsensusGenesis{Validators: nil}

			if err = genutil.ExportGenesisFile(appGenesis, genFile); err != nil {
				return errorsmod.Wrap(err, "failed to export genesis file")
			}

			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "", appState)

			params.SetTendermintConfigs(config)
			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

			return displayInfo(toPrint)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(genutilcli.FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().Bool(genutilcli.FlagRecover, false, "recover existing key using seed phrase")
	cmd.Flags().String(flags.FlagChainID, "", "genesis chain-id (random if left blank)")
	cmd.Flags().String(genutilcli.FlagDefaultBondDenom, "", "default denomination (default 'akii')")
	cmd.Flags().Int64(flags.FlagInitHeight, 1, "initial block height at genesis")

	return cmd
}
