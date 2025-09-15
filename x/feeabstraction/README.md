# Fee Abstraction Module

The Fee Abstraction module allows the payment of transaction fees in other arbitrary tokens different from the native token of the chain. This allows users to pay for fees in a more flexible way, using stablecoins, for example.

The functionality of the module allows:

- Pricing of transaction fees in arbitrary tokens.
- Payment through native token (unwrapped) or through ERC20 tokens (wrapped).
- Pricing of tokens is defined by the oracle module, which provides the exchange rates for the tokens.
  - This allows the use of stablecoins as fee tokens.
- Applies safety mechanisms to prevent price manipulation:
  - Uses TWAP (Time-Weighted Average Price) instead of spot prices.
  - Applies a clamp factor to limit extreme price deviations.
  - Disables tokens with missing or zero prices.

## Core functionality

In its code, the module changes the fee-paying ante handlers to allow payment of fees with tokens different from the native gas token.

- This is done through both EVM and cosmos SDK ante handlers.
- The pricing of tokens is done with the following logic:
  - The price of the token is defined by the oracle module in USD
  - The module then stores the price of the token per gas token
  - This allows for easy conversion when paying fees
  - And allows for easy querying

### Price calculation

The price calculation happens at the beginning of each block:

- We query the oracle module for the Twap (Time Weighted Average Price) of all available tokens in USD
- Then we iterate over all the possible fee tokens and calculate the price of the token in gas tokens
  - The formula is:

```
price_in_gas_token = price_in_usd / gas_price_in_usd
```

- During this calculation, we don't consider decimals, and all prices are in the same unit
- The final pricing is stored in the module state

We have a few safety mechanisms to ensure the prices are valid:

- If the oracle module can't provide a price, the fee token is disabled
- If prices go to zero, the fee token is disabled
- The Twap of the token is used to avoid sudden price changes
- Price changes are clamped to avoid extreme values

### Fee payment

The module by itself does not handle the fee payment:

- It returns available assets to each ante handler to process the deduct fee

The following is the process to check for possible fee tokens:

1. Ante handler receives and calculates the fee for the native gas token
2. The fee goes to the fee abstraction module and is validated
   - The module ignores zero tokens, tokens different from the native fee token, and fees with multiple tokens
   - Due to the chain EVM nature, fees with multiple tokens are not supported
3. The user's balance is checked for the fee token
   - If the user has enough balance, the fee is returned as it is
4. If the user does not have enough balance, the module checks for available fee tokens
   - The available fee tokens are those that have a valid price and are not disabled
5. The module then calculates the fee in the available fee tokens
   - The fee is calculated using the price stored in the module state
6. If not available through the unwrapped native token, the module checks for wrapped ERC20 tokens
   - If the user has enough balance in the wrapped token, the token is unwrapped
   - This makes the token available for the fee payment
7. The fee is returned from the module to the ante handler
8. The ante handler then deducts the fee from the user's balance

```mermaid
flowchart TD
    A[Ante handler receives fee in native gas token] --> B[Fee sent to Fee Abstraction module for validation]
    B --> C{Is the fee valid?}
    C -->|Invalid fee| Z[Reject fee]
    C -->|Valid fee| D[Check user balance for native fee token]
    D --> E{Enough balance?}
    E -->|Yes| F[Return new fee to ante handler]
    E -->|No| G[Check available fee tokens with valid price]
    G --> H[Calculate fee using stored price]
    H --> I{Available via unwrapped native token?}
    I -->|Yes| F
    I -->|No| J[Check wrapped ERC20 tokens]
    J --> K{Enough wrapped balance?}
    K -->|Yes| L[Unwrap ERC20 token → Make available for fee payment]
    L --> F
    K -->|No| Z
    F --> M[Ante handler deducts fee from user balance]
```

## State

The most important state types used by the Fee Abstraction module are:

### Params

Params stores the module wide configuration. It is defined as:

```proto
// Params defines the parameters for the fee abstraction module
message Params {
  // Native denom
  string native_denom = 1;
  // Oracle module identifier
  string native_oracle_denom = 2;
  // Enabled indicates if the fee abstraction module is enabled
  bool enabled = 3;
  // ClampFactor is the factor to clamp the price deviation
  string clamp_factor = 4 [
    (gogoproto.moretags) = "yaml:\"clamp_factor\"",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];
  // TwapLookbackWindow is the lookback window for calculating TWAPs
  uint64 twap_lookback_window = 5;
  // FallbackNativePrice is the fallback price for the native token if the
  // oracle price is not available (in USD)
  string fallback_native_price = 6 [
    (gogoproto.moretags) = "yaml:\"fallback_native_price\"",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];
}
```

### FeeToken

FeeToken stores the information about a fee token. It is defined as:

```proto
// FeeTokenMetadata defines the metadata for a fee token
message FeeTokenMetadata {
  // Denom is the token denom
  string denom = 1;
  // Identifier on the oracle module
  string oracle_denom = 2;
  // Decimals is the number of decimals for the token
  uint32 decimals = 3;
  // Price is the price of the token in the native denom
  // This price is paired against the native denom
  // So, this equals to the token/native denom
  string price = 4 [
    (gogoproto.moretags) = "yaml:\"price\"",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];
  // Enabled indicates if the token is enabled for fee abstraction
  bool enabled = 6;
}

// Defines a collection of fee token metadata
message FeeTokenMetadataCollection {
  // Items is a repeated field of FeeTokenMetadata
  repeated FeeTokenMetadata items = 1 [ (gogoproto.nullable) = false ];
}
```

## Messages

The module defines the following messages:

### MsgUpdateParams

MsgUpdateParams is used to update the module parameters.
Only the governance account can update the parameters, and it requires a valid signature.
It is defined as:

```proto
// MsgUpdateParams is the Msg/UpdateParams request type.
message MsgUpdateParams {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name) = "feeabstraction/update-params";

  // authority is the address of the governance account.
  string authority = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];

  // params defines the x/feeabstraction parameters to update.
  Params params = 2 [ (gogoproto.nullable) = false ];
}
```

### MsgUpdateFeeTokens

The `MsgUpdateFeeTokens` message is used to update the fee tokens metadata.
Only the governance account can update the fee tokens, and it requires a valid signature.
On each update, all fee tokens must be provided, even if they are not changed:

- Ordering is important, since the balance will be calculated based on the order of the fee tokens.

It is defined as:

```proto
// MsgUpdateFeeTokens is the Msg/UpdateFeeTokens request type.
message MsgUpdateFeeTokens {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name) = "feeabstraction/update-fee-tokens";

  // authority is the address of the governance account.
  string authority = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];

  // fee_tokens defines the fee tokens to update.
  FeeTokenMetadataCollection fee_tokens = 2 [ (gogoproto.nullable) = false ];
}
```

## Queries

The module provides the following queries:

### QueryParams

The `QueryParams` query is used to retrieve the current module parameters. It does not require any input and returns the current parameters.

```proto
// QueryParamsResponse is the response type for the Query/Params RPC method.
message QueryParamsResponse {
  // params defines the parameters of the module.
  Params params = 1 [ (gogoproto.nullable) = false ];
}
```

### QueryFeeTokens

The `QueryFeeTokens` query is used to retrieve the current fee tokens metadata. It does not require any input and returns the current fee tokens.

```proto
// QueryFeeTokensResponse is the response type for the Query/FeeTokens RPC
// method
message QueryFeeTokensResponse {
  // fee_tokens defines the fee tokens registered in the module.
  FeeTokenMetadataCollection fee_tokens = 1;
}
```

## Begin block

On each ABCI call, the Fee Abstraction module performs the following actions:

1. Check the current Twap for each fee token against the oracle module.
2. Update the price of each fee token based on the Twap.
3. Clamp the price of each fee token to avoid extreme values.
4. Disable fee tokens that have no valid price or have a price of zero.
5. Update the module state with the new prices and enabled status of the fee tokens.
6. If the module is disabled, it will not perform any of the above actions and will not allow fee abstraction.

```mermaid
flowchart TD
    A[ABCI Call] --> B{Is Fee Abstraction module enabled?}
    B -->|No| G[Skip actions and disallow fee abstraction]
    B -->|Yes| C[Check current TWAP for each fee token via oracle module]
    C --> D[Update price of each fee token based on TWAP]
    D --> E[Clamp price of each fee token to avoid extremes]
    E --> F[Disable tokens with no valid price or zero price]
    F --> H[Update module state with new prices and enabled status]
```

## Ante Handlers

The Fee Abstraction module provides custom ante handlers to handle fee payments in the specified fee tokens.

### fee.go (Cosmos Ante Handler)

This is the Cosmos SDK ante handler:

- Has the same implementation as the [original fee ante handler](https://github.com/cosmos/cosmos-sdk/blob/main/x/auth/ante/fee.go).
- The main difference is that the fees goes though the Fee Abstraction module before fee deduction.

### mono_decorator.go (EVM Ante Handler)

This is the EVM ante handler:

- Since the EVM module has a coupled implementation of the ante handler, the whole ante had to be migrated
- The implementation is mostly the same as the [original EVM ante handler](https://github.com/cosmos/evm/blob/main/ante/evm/mono_decorator.go)

The following was required to be changed:

- Balance checks were broken down into different steps:
  - Check if the user has enough balance to pay for fees
  - Check if the user has enough balance to pay for TX value
- Account creation was moved up to allow accounts to exist before the fee deduction
- At the end of the ante handler, the fee is registered on the context
  - This allows fee refunds to be processed correctly

## Limitation

A limitation happens with **fresh wallets** (wallets that have never executed a transaction):

- **Scenario 1: Fresh wallet with ERC20 (wrapped) fee tokens**

  - If the first transaction attempted is a **Cosmos SDK transaction** using only ERC20 fee tokens:

    - The transaction will **fail with an “account doesn’t exist” error**.
    - This is because the Cosmos SDK requires an on-chain account to exist prior to executing any Cosmos transaction, and ERC20 balances alone do not create the account record.

- **Scenario 2: Fresh wallet with native (unwrapped) fee tokens**

  - If the first transaction attempted is a **Cosmos SDK transaction** using native fee tokens:

    - The transaction will **succeed**, as the presence of the native asset allows account creation during the transaction flow, and the fee abstraction module can process the fees normally.

- **Scenario 3: Fresh wallet with either ERC20 or native fee tokens, performing EVM transactions**

  - If the first transaction attempted is an **EVM transaction**:

    - The transaction will **succeed**, regardless of whether the wallet holds ERC20 or native tokens, as the EVM module internally handles account creation differently.

### Implication

Users must perform **one of the following actions** before attempting a Cosmos SDK transaction using ERC20 fee tokens:

- Fund the wallet with **any amount of the native (unwrapped) token** and execute a small Cosmos transaction to create the account.
- Or perform an **EVM transaction first** which also creates the account record.
