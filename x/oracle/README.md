# Oracle module

The Oracle module brings up price feeds and exchange rates to Kiichain.
- It allows validators to submit price data, which can be used by other modules or smart contracts (EVM and Wasm).
- It supports multiple price feeds, allowing for a diverse set of data sources
- It provides a mechanism for price data to be aggregated and made available on-chain

## Price feeder

Our official implementation for the price feeder is available at [Kiichain Price Feeder](https://github.com/KiiChain/price-feeder).
- The price must be ran by a validator or a delegated address
- The price feeder is responsible for submitting the price data to the Oracle module
- More information can be found at the project readme

## Core functionality

The Oracle module works as follows:
- Validators submit price data for arbitrary assets, which are defined in the module params
- The module aggregates the price data submitted by validators and calculates a final exchange rate for each asset
- The final exchange rate is stored on-chain and can be queried by other modules or smart contracts

The Exchange Vote is done as following:
1. A new block is created
2. Validators submit their votes for the price of each asset in the whitelist though the `MsgExchangeRateVote` message
  - This message is feeless as long as its the first vote for the validator in the current voting period
3. The module aggregates the votes and calculates the final exchange rate for each asset
4. If no vote is submitted by a validator in the current voting period, the module will slash the validator's stake according to the `slash_fraction` parameter
5. The final exchange rate is stored on-chain and can be queried by other modules or smart contracts

## State

These are the most important state types used by the Oracle module:

### Params

Params stores the module wide configuration. Its defined as:

```proto
// Params defines the parameters for the module
message Params {
    option (gogoproto.equal)            = true;
    option (gogoproto.goproto_stringer) = false; // Allow custom logs 
    
    // The number of blocks per voting 
    uint64 vote_period =1 [(gogoproto.moretags) = "yaml:\"vote_period\""];

    // Minimum percentage of validators required to approve a price. For instance, if vote_threshold = "0.5" at least 50% of validators must submit votes 
    // "cosmossdk.io/math.LegacyDec" = Cosmos SDK decimal data type
    string vote_threshold = 2 [
        (gogoproto.moretags) = "yaml:\"vote_threshold\"",
        (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec", 
        (gogoproto.nullable) = false 
    ];        

    // Acceptable deviation from the media price (higher and lower)
    // "cosmossdk.io/math.LegacyDec" = Cosmos SDK decimal data type
    string reward_band = 3 [
        (gogoproto.moretags) = "yaml:\"reward_band\"",
        (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec", 
        (gogoproto.nullable) = false
    ];

    // List of allowed assets 
    // DenomList is a custom data type, defined on x/oracle/types/denom.go
    repeated Denom whitelist = 4 [
        (gogoproto.moretags) = "yaml:\"whitelist\"",
        (gogoproto.castrepeated) = "DenomList", 
        (gogoproto.nullable) = false
    ];

    // How much stake is slashed if a validator fails to submit votes 
    // "cosmossdk.io/math.LegacyDec" = Cosmos SDK decimal data type
    string slash_fraction = 5 [
        (gogoproto.moretags) = "yaml:\"slash_fraction\"",
        (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec", 
        (gogoproto.nullable) = false
    ];

    // Define the window (in blocks) to vote, if not receive penalties due to bad performance
    uint64 slash_window = 6 [(gogoproto.moretags) = "yaml:\"slash_window\""];

    // Minimum percentage of voting on windows to avoid slashing. For instance, if min_valid_per_window = 0.8, then a validator must submit votes in 80% of windows to avoid slashing 
    // "cosmossdk.io/math.LegacyDec" = Cosmos SDK decimal data type
    string min_valid_per_window = 7 [
        (gogoproto.moretags) = "yaml:\"min_valid_per_window\"",
        (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec", 
        (gogoproto.nullable) = false
    ];

    // How far back (in blocks) the module can compute historical price metrics 
    uint64 lookback_duration = 9 [(gogoproto.moretags) = "yaml:\"lookback_duration\""];
}
```

### Exchange Rates

Exchange rates are the single entry for a price data on the chain. Its stored as a Key-Value pair in the store, where the key is the asset denom and the value is the price data.

The price data is defined as:

```proto
// Data type that stores the final calculated exchange rate after all votes were 
// aggregated to that single exchange, record the last block height and timestamp when rate was updated 
message OracleExchangeRate {
    option (gogoproto.equal)            = false;
    option (gogoproto.goproto_getters)  = false;
    option (gogoproto.goproto_stringer) = false;

    string exchange_rate = 1 [
        (gogoproto.moretags)   = "yaml:\"exchange_rate\"",
        (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
        (gogoproto.nullable)   = false
    ];

    string last_update = 2 [
        (gogoproto.moretags)   = "yaml:\"last_update\"",
        (gogoproto.customtype) = "cosmossdk.io/math.Int",
        (gogoproto.nullable)   = false
    ];

    int64 last_update_timestamp = 3 [(gogoproto.moretags)   = "yaml:\"last_update_timestamp\""];
}
```

### FeederDelegation

Feeder delegations is the correlation between a validator and a feeder address.
It allows validators to use a different address to submit votes.

The FeederDelegation is defined as:

```proto
// FeederDelegation is the structure on the genesis regarding the delegation process 
message FeederDelegation {
  // feeder_address is the address delegated  
  string feeder_address    = 1;

  // validator_address is the validator's address who delegate its voting action 
  string validator_address = 2;
}
```

## Messages

The Oracle module expose the following messages:

### AggregateExchangeRateVote

The `MsgAggregateExchangeRateVote` message is used by validators to submit their votes for the price of each asset in the whitelist. It contains the following fields:

```proto
// MsgAggregateExchangeRateVote represent the message to submit
// an aggregate exchange rate vote
message MsgAggregateExchangeRateVote{
  option (gogoproto.equal)           = false;
  option (gogoproto.goproto_getters) = false;

  option (cosmos.msg.v1.signer) = "feeder";
  option (amino.name) = "oracle/aggregate-exchange-rate-vote";

  string exchange_rates = 1 [(gogoproto.moretags) = "yaml:\"exchange_rates\""];
  string feeder = 2 [(gogoproto.moretags) = "yaml:\"feeder\""];
  string validator = 3 [(gogoproto.moretags) = "yaml:\"validator\""];
}
```

### DelegateFeedConsent

The `MsgDelegateFeedConsent` message is used to delegate the right to submit votes to a different address. The message contains the following fields:

```proto
// MsgDelegateFeedConsent represents a message to delegate oracle voting 
// rights to another address
message MsgDelegateFeedConsent{
  option (gogoproto.equal)           = false;
  option (gogoproto.goproto_getters) = false;

  option (cosmos.msg.v1.signer) = "validator_owner";
  option (amino.name) = "oracle/delegate-feed-consent";

  string validator_owner = 1 [(gogoproto.moretags) = "yaml:\"validator_owner\""];
  string delegate = 2 [(gogoproto.moretags) = "yaml:\"delegate\""];
}
```

### UpdateParams

The `MsgUpdateParams` message is used to update the module parameters. Only the governance module can call the message. It contains the following fields:

```proto
// MsgUpdateParams is the Msg/UpdateParams request type
message MsgUpdateParams {
  option (cosmos.msg.v1.signer) = "authority";

  // authority is the address that controls the module (defaults to x/gov)
  string authority    = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  option (amino.name) = "kiichain/x/oracle/MsgUpdateParams";

  // params defines the x/oracle parameters to update
  //
  // NOTE: All parameters must be supplied.
  Params params = 2 [(gogoproto.nullable) = false, (amino.dont_omitempty) = true];
}
```

## Begin block

On each ABCI call, the Oracle module performs the following actions:
1. Check if we are under a new slash window
2. Check the slash counters for validators and slash them if they didn't submit enough votes in the previous voting period
3. Remove the excess feeds

## End block

At the end of each block, the Oracle module performs the following actions:
1. Check if we are under a new voting period
2. Iterate the votes
3. Calculate the final exchange rate for each asset in the whitelist
4. Store the final exchange rate on-chain

## Ante handler

The Oracle module ignores fees from validators on their first vote in the current voting period.
The following is done:
1. Check if the message is a `MsgAggregateExchangeRateVote`
2. Check the validator/feeder relationship
3. If the validator is voting for the first time in the current voting period, ignore the fees

# Acknowledgments

Special thanks to the SEI team. Your contributions to the Cosmos SDK ecosystem are greatly appreciated. The original implementation of the Oracle module can be found in the [SEI repository](https://github.com/sei-protocol/sei-chain/tree/main/x/oracle)