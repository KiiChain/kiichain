package interchaintest

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testutil"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	ibcconntypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	tokenfactory "github.com/kiichain/kiichain/v7/x/tokenfactory/types"
)

var (
	VotingPeriod     = "15s"
	MaxDepositPeriod = "10s"

	Denom        = "akii"
	DisplayDenom = "kii"
	Name         = "kiichain"
	ChainID      = "localchain_1010-1"
	Binary       = "kiichaind"
	Bech32       = "kii"

	NumberVals         = 1
	NumberFullNodes    = 0
	GenesisFundsAmount = sdkmath.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(600_000_000_000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))) // 600B tokens with 18 decimals

	ChainImage = ibc.NewDockerImage("kiichain", "local", "1025:1025")

	bankMetadataJSON = `[{"description":"The native staking token of the kiichain network","denomUnits":[{"denom":"akii","exponent":"0"},{"denom":"kii","exponent":"18"}],"base":"akii","display":"kii","name":"kii","symbol":"KII"}]`

	DefaultGenesis = []cosmos.GenesisKV{
		// default
		cosmos.NewGenesisKV("app_state.gov.params.voting_period", VotingPeriod),
		cosmos.NewGenesisKV("app_state.gov.params.max_deposit_period", MaxDepositPeriod),
		cosmos.NewGenesisKV("app_state.gov.params.min_deposit.0.denom", Denom),
		cosmos.NewGenesisKV("app_state.gov.params.min_deposit.0.amount", "1"),
		// evm reqs
		cosmos.NewGenesisKV("app_state.evm.params.evm_denom", Denom),
		cosmos.NewGenesisKV("app_state.bank.denom_metadata", bankMetadataJSON),
		// tokenfactory: set create cost in set denom or in gas usage.
		cosmos.NewGenesisKV("app_state.tokenfactory.params.denom_creation_fee", nil),
		cosmos.NewGenesisKV("app_state.tokenfactory.params.denom_creation_gas_consume", "1"), // cost 1 gas to create a new denom
		// consensus: allow zero fees for testing
		cosmos.NewGenesisKV("consensus.params.abci.vote_extensions_enable_height", "0"),
		// feeabstraction: configure for testing environment
		cosmos.NewGenesisKV("app_state.feeabstraction.params.native_denom", Denom),
		cosmos.NewGenesisKV("app_state.feeabstraction.params.native_oracle_denom", Denom),
		cosmos.NewGenesisKV("app_state.feeabstraction.params.enabled", false), // disable for testing
		cosmos.NewGenesisKV("app_state.feeabstraction.params.clamp_factor", "1.0"),
		cosmos.NewGenesisKV("app_state.feeabstraction.params.twap_lookback_window", "1"),
	}

	DefaultChainConfig = ibc.ChainConfig{
		Images: []ibc.DockerImage{
			ChainImage,
		},
		GasAdjustment: 1.5,
		ModifyGenesis: cosmos.ModifyGenesis(DefaultGenesis),
		ModifyGenesisAmounts: func(i int) (sdk.Coin, sdk.Coin) {
			// Set genesis amount and self-delegation for validators
			// Need to account for 18 decimal places (AttoPowerReduction = 10^18)
			genesisAmount := sdk.NewCoin(Denom, GenesisFundsAmount)
			// Self-delegation needs to be at least 275B * 10^18 to meet minimum validator power
			selfDelegation := sdk.NewCoin(Denom, sdkmath.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(300_000_000_000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))))
			return genesisAmount, selfDelegation
		},
		EncodingConfig: GetEncodingConfig(),
		Type:           "cosmos",
		Name:           Name,
		ChainID:        ChainID,
		Bin:            Binary,
		Bech32Prefix:   Bech32,
		Denom:          Denom,
		CoinType:       "118",
		GasPrices:      "0" + Denom, // Use zero gas price since fee abstraction is disabled
		TrustingPeriod: "504h",
	}

	DefaultChainSpec = interchaintest.ChainSpec{
		Name:          Name,
		ChainName:     Name,
		Version:       ChainImage.Version,
		ChainConfig:   DefaultChainConfig,
		NumValidators: &NumberVals,
		NumFullNodes:  &NumberFullNodes,
	}

	SecondDefaultChainSpec = func() interchaintest.ChainSpec {
		SecondChainSpec := DefaultChainSpec
		SecondChainSpec.ChainID += "2"
		SecondChainSpec.Name += "2"
		SecondChainSpec.ChainName += "2"
		return SecondChainSpec
	}()

	// cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr - test_node.sh
	AccMnemonic  = "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"
	Acc1Mnemonic = "wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise"

	RelayerRepo    = "ghcr.io/cosmos/relayer"
	RelayerVersion = "main"

	vals   = 1
	fNodes = 0
)

func GetEncodingConfig() *moduletestutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()
	// TODO: add encoding types here for the modules you want to use
	wasm.RegisterInterfaces(cfg.InterfaceRegistry)
	tokenfactory.RegisterInterfaces(cfg.InterfaceRegistry)
	return &cfg
}

// Other Helpers
func ExecuteQuery(ctx context.Context, chain *cosmos.CosmosChain, cmd []string, i interface{}, extraFlags ...string) {
	flags := []string{
		"--node", chain.GetRPCAddress(),
		"--output=json",
	}
	flags = append(flags, extraFlags...)

	ExecuteExec(ctx, chain, cmd, i, flags...)
}

func ExecuteExec(ctx context.Context, chain *cosmos.CosmosChain, cmd []string, i interface{}, extraFlags ...string) {
	command := []string{chain.Config().Bin}
	command = append(command, cmd...)
	command = append(command, extraFlags...)
	fmt.Println(command)

	stdout, _, err := chain.Exec(ctx, command, nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(stdout))
	if err := json.Unmarshal(stdout, &i); err != nil {
		fmt.Println(err)
	}
}

// Executes a command from CommandBuilder
func ExecuteTransaction(ctx context.Context, chain *cosmos.CosmosChain, cmd []string) (sdk.TxResponse, error) {
	var err error
	var stdout []byte

	stdout, _, err = chain.Exec(ctx, cmd, nil)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		return sdk.TxResponse{}, err
	}

	var res sdk.TxResponse
	if err := json.Unmarshal(stdout, &res); err != nil {
		return res, err
	}

	return res, err
}

func TxCommandBuilder(ctx context.Context, chain *cosmos.CosmosChain, cmd []string, fromUser string, extraFlags ...string) []string {
	return TxCommandBuilderNode(ctx, chain.GetNode(), cmd, fromUser, extraFlags...)
}

func TxCommandBuilderNode(ctx context.Context, node *cosmos.ChainNode, cmd []string, fromUser string, extraFlags ...string) []string {
	command := []string{node.Chain.Config().Bin}
	command = append(command, cmd...)
	command = append(command, "--node", node.Chain.GetRPCAddress())
	command = append(command, "--home", node.HomeDir())
	command = append(command, "--chain-id", node.Chain.Config().ChainID)
	command = append(command, "--from", fromUser)
	command = append(command, "--keyring-backend", keyring.BackendTest)
	command = append(command, "--output=json")
	command = append(command, "--yes")

	gasFlag := false
	for _, flag := range extraFlags {
		if flag == "--gas" {
			gasFlag = true
		}
	}

	if !gasFlag {
		command = append(command, "--gas", "500000")
	}

	command = append(command, extraFlags...)
	return command
}

func getTransferChannel(channels []ibc.ChannelOutput) (string, error) {
	for _, channel := range channels {
		if channel.PortID == "transfer" && channel.State == ibcconntypes.OPEN.String() {
			return channel.ChannelID, nil
		}
	}

	return "", fmt.Errorf("no open transfer channel found: %+v", channels)
}

func SetupContract(t interface{}, ctx context.Context, chain *cosmos.CosmosChain, keyname string, fileLoc string, message string, extraFlags ...string) (codeId, contract string) {
	codeId, err := chain.StoreContract(ctx, keyname, fileLoc)
	if err != nil {
		panic(err)
	}

	needsNoAdminFlag := true
	for _, flag := range extraFlags {
		if flag == "--admin" {
			needsNoAdminFlag = false
		}
	}

	contractAddr, err := chain.InstantiateContract(ctx, keyname, codeId, message, needsNoAdminFlag, extraFlags...)
	if err != nil {
		panic(err)
	}

	return codeId, contractAddr
}

func SmartQueryString(t interface{}, ctx context.Context, chain *cosmos.CosmosChain, contractAddr, queryMsg string, res interface{}) error {
	var jsonMap map[string]interface{}
	if err := json.Unmarshal([]byte(queryMsg), &jsonMap); err != nil {
		panic(err)
	}
	err := chain.QueryContract(ctx, contractAddr, jsonMap, &res)
	return err
}
