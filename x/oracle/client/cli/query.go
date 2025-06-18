package cli

import (
	"context"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v2/x/oracle/types"
)

// GetQueryCmd returns the cli query commands for the module
func GetQueryCmd() *cobra.Command {
	// Register the oracle query subcommands
	oracleQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the oracle module",
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Add Query commands
	oracleQueryCmd.AddCommand(
		CmdQueryExchangeRates(),
		CmdQueryPriceSnapshotHistory(),
		CmdQueryTwaps(),
		CmdQueryActives(),
		CmdQueryParams(),
		CmdQueryFeederDelegation(),
		CmdQueryVotePenaltyCounter(),
	)

	return oracleQueryCmd
}

// CmdQueryExchangeRates is the command executed when users type "exchange-rates [denom]"
func CmdQueryExchangeRates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange-rates [denom]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "Query the current exchange rate w.r.t an asset",
		Long: strings.TrimSpace(`
Query the current exchange rate of Kii with an asset.
You can find the current list of active denoms by running 
		
$kiichaind query oracle exchange-rates
		
Or filter by denom running 

$kiichaind query oracle exchange-rates <denom>

where denom is the denom you want to filter by 
		`),

		RunE: getExchangeRate,
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdQueryPriceSnapshotHistory is the command executed when users type "price-snapshot-history" command
func CmdQueryPriceSnapshotHistory() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price-snapshot-history",
		Args:  cobra.NoArgs,
		Short: "Query the history for oracle price snapshots",
		Long: strings.TrimSpace(`
Query the history for oracle price snapshots.
		
$kiichaind query oracle price-snapshot-history`),

		RunE: getPriceSnapshotHistory,
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdQueryTwaps is the command executed when users type "twaps [lookback-seconds]" command
func CmdQueryTwaps() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "twaps [lookback-seconds]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the time weighted average (Twap) prices for denom from prices snapshot data",
		Long: strings.TrimSpace(`
Query the time weighted average prices for denoms from price snapshot data
		
$kiichaind query oracle twaps 1
		
where 1 means 1 second `),
		RunE: getTwaps,
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdQueryActives is the command executed when users type "actives" command
func CmdQueryActives() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "actives",
		Args:  cobra.NoArgs,
		Short: "Query the active assets list recognized by the oracle module",
		Long: strings.TrimSpace(`
Query the active assets list recognized by the oracle module.

$kiichaind query oracle actives
		`),
		RunE: getActives,
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdQueryParams is the command executed when users type params command
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current oracle params",
		RunE:  getParams,
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdQueryFeederDelegation is the command executed when users type feeder [validator]
func CmdQueryFeederDelegation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feeder [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the oracle feeder delegated account",
		Long: strings.TrimSpace(`
Query the account the validator's oracle voting right is delegated to
		
$kiichaind query oracle feeder kiivaloper.....`),
		RunE: getFeederDelegation,
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdQueryVotePenaltyCounter is the command executed when users type vote-penalty-counter [validator]
func CmdQueryVotePenaltyCounter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-penalty-counter [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the number of miss count and abstain count",
		Long: strings.TrimSpace(`
Query the number of vote periods missed and abstained in the current slash window

$kiichaind query oracle vote-penalty-counter kiivaloper...`),
		RunE: getVotePenaltyCounter,
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdQueryVoteTargets is the command executed when users type vote-targets
func CmdQueryVoteTargets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-targets",
		Args:  cobra.NoArgs,
		Short: "Query the current oracle vote targets",
		RunE:  getVoteTargets,
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getExchangeRate queries the exchange rates on the oracle module, returns all or
// an specific one if the user add it on the command
func getExchangeRate(cmd *cobra.Command, args []string) error {
	// get ctx
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	// create query client
	queryClient := types.NewQueryClient(clientCtx)

	// Return all exchange rates
	if len(args) == 0 {
		rates, err := queryClient.ExchangeRates(context.Background(), &types.QueryExchangeRatesRequest{})
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(rates) // print msg response
	}

	// Return specific denom
	denom := args[0]
	rate, err := queryClient.ExchangeRate(context.Background(), &types.QueryExchangeRateRequest{Denom: denom})
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(rate) // print msg response
}

// getPriceSnapshotHistory returns the price snapshot history
func getPriceSnapshotHistory(cmd *cobra.Command, args []string) error {
	// get ctx
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	// Create query client
	queryClient := types.NewQueryClient(clientCtx)

	// Get snapshot history
	res, err := queryClient.PriceSnapshotHistory(context.Background(), &types.QueryPriceSnapshotHistoryRequest{})
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res) // print msg response
}

// getTwaps returns the time weighted average price within an specific time period
func getTwaps(cmd *cobra.Command, args []string) error {
	// get ctx
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	// create query client
	queryClient := types.NewQueryClient(clientCtx)

	// get lookback time
	lookbackSeconds, err := strconv.ParseUint(args[0], 10, 64) // get uint64 from the string arg
	if err != nil {
		return err
	}

	// get twap
	res, err := queryClient.Twaps(context.Background(), &types.QueryTwapsRequest{LookbackSeconds: lookbackSeconds})
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res) // print msg response
}

// getActives returns the list of assets recognized by the oracle module
func getActives(cmd *cobra.Command, args []string) error {
	// get ctx
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	// create query client
	queryClient := types.NewQueryClient(clientCtx)

	// get active list
	res, err := queryClient.Actives(context.Background(), &types.QueryActivesRequest{})
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res) // print msg response
}

// getParams returns the current module params
func getParams(cmd *cobra.Command, args []string) error {
	// get ctx
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	// create query client
	queryClient := types.NewQueryClient(clientCtx)

	// get params
	res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res) // print msg response
}

// getFeederDelegation returns the validator's delegated account
func getFeederDelegation(cmd *cobra.Command, arg []string) error {
	// get ctx
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	// get validator address
	valAddrString := arg[0]
	validator, err := sdk.ValAddressFromBech32(valAddrString)
	if err != nil {
		return err
	}

	// create query client
	queryClient := types.NewQueryClient(clientCtx)

	// get validator's delegated account
	res, err := queryClient.FeederDelegation(context.Background(), &types.QueryFeederDelegationRequest{ValidatorAddr: validator.String()})
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res) // print msg response
}

// getVotePenaltyCounter returns the vote penalty counter by validator address
func getVotePenaltyCounter(cmd *cobra.Command, arg []string) error {
	// get ctx
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	// get validator address
	valAddrString := arg[0]
	validator, err := sdk.ValAddressFromBech32(valAddrString)
	if err != nil {
		return err
	}

	// create query client
	queryClient := types.NewQueryClient(clientCtx)

	// get validator penalty counter
	res, err := queryClient.VotePenaltyCounter(context.Background(), &types.QueryVotePenaltyCounterRequest{ValidatorAddr: validator.String()})
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res) // print msg response
}

// getVoteTargets returs the current vote targets
func getVoteTargets(cmd *cobra.Command, arg []string) error {
	// get ctx
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	// create query client
	queryClient := types.NewQueryClient(clientCtx)

	// get current vote targets
	res, err := queryClient.VoteTargets(context.Background(), &types.QueryVoteTargetsRequest{})
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res) // print msg response
}
