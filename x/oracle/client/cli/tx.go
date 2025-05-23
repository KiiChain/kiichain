package cli

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the tx commands for oracle module
func GetTxCmd() *cobra.Command {
	// Register the oracle transactions subcommands
	oracleTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Oracle transmition subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Add Tx commands
	oracleTxCmd.AddCommand(
		CmdDelegateFeederPermission(),
		CmdAggregateExchangeRateVote(),
	)

	return oracleTxCmd
}

// CmdDelegateFeederPermission is the command executed when users type "$ kiichaind tx oracle set-feeder kii1...."
// on the CLI
func CmdDelegateFeederPermission() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-feeder [feeder]",
		Args:  cobra.ExactArgs(1),
		Short: "Delegate the permissions to vote for the oracle to an address",
		Long: strings.TrimSpace(`
Delegate the permission to submit exchange rate votes for the oracle to an address.
		
Delegation can keep your validator operator key offline and use a separate replaceable key online.
		
$ kiichaind tx oracle set-feeder kii1....
		
where "kii1..." is the address you want to delegate your voting rights to.`),
		RunE: setFeeder,
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdAggregateExchangeRateVote is the command executed when users type ""
// on the CLI
func CmdAggregateExchangeRateVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aggregate-vote [exchange-rates] [validator]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Submit an oracle aggregate vote with the exchange rates",
		Long: strings.TrimSpace(`
Submit an aggregate vote with the exchange rates.
		
$kiichaind tx oracle aggregate-vote 123.45ukii,678.90uatom...
		
where "ukii,uatom,ueth..." are the denominating currencies and 123.45,678.90 are the exchange rates of micro USD in micro denoms
		
If voting from a delegate account, set "validator" to the address of the validator you are voting on behalf of, i.e:
		
$ kiichaind oracle aggregate-vote 123.45ukii,678.90uatom... kiivaloper1...`),
		RunE: aggregateVote,
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// setFeeder is executed with the command "set-feeder [feeder]". It delegates
// the permission to submit exchange rate to an address
func setFeeder(cmd *cobra.Command, args []string) error {
	// get ctx
	clientCtx, err := client.GetClientTxContext(cmd)
	if err != nil {
		return err
	}

	// Get from address
	voter := clientCtx.GetFromAddress()

	// the delegator address
	valAddress := sdk.ValAddress(voter)

	// Get feeder address
	feederStr := args[0]
	feeder, err := sdk.AccAddressFromBech32(feederStr)
	if err != nil {
		return err
	}

	// Create delegate feed consent message
	msg := types.NewMsgDelegateFeedConsent(valAddress, feeder)
	err = msg.ValidateBasic()
	if err != nil {
		return err
	}

	return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
}

// aggregateVote is executed with the command "aggregate-vote [exchange-rates] [validator]"
// it sends the exchange rate voting message
func aggregateVote(cmd *cobra.Command, args []string) error {
	// get ctx
	clientCtx, err := client.GetClientTxContext(cmd)
	if err != nil {
		return err
	}

	// Get exchange rates
	exchangeRatesStr := args[0]
	_, err = types.ParseExchangeRateTuples(exchangeRatesStr)
	if err != nil {
		return err
	}

	// Get from address
	voter := clientCtx.GetFromAddress()

	// by default the voter is voting on bhalf of itself
	valAddress := sdk.ValAddress(voter)

	// overide validator if validator's address is given
	if len(args) == 2 {
		parsedVal, err := sdk.ValAddressFromBech32(args[1])
		if err != nil {
			return errors.Wrap(err, "validator address is invalid")
		}
		valAddress = parsedVal
	}

	// Create aggregate exchange rate vote message
	msg := types.NewMsgAggregateExchangeRateVote(exchangeRatesStr, voter, valAddress)
	err = msg.ValidateBasic()
	if err != nil {
		return err
	}

	return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
}
