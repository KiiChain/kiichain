# Oracle Precompile

The oracle precompile enables EVM smart contracts to query KiiChain's oracle module for real-time exchange rates and TWAP (Time-Weighted Average Price) data.

## Overview

- **Address**: `0x0000000000000000000000000000000000001003`
- **Interface**: [`IOracle.sol`](./IOracle.sol)
- **Module**: `x/oracle`

## Functions

### `getExchangeRate(string denom)`

Returns the current exchange rate for a single denomination.

| Parameter | Type | Description |
|-----------|------|-------------|
| `denom` | `string` | The denomination to query (max 128 bytes) |

**Returns:**

| Name | Type | Description |
|------|------|-------------|
| `rate` | `string` | The current exchange rate |
| `lastUpdate` | `string` | Block height of the last update |
| `lastUpdateTimestamp` | `int64` | Unix timestamp of the last update |

### `getExchangeRates()`

Returns exchange rates for all active denominations.

**Returns:**

| Name | Type | Description |
|------|------|-------------|
| `denoms` | `string[]` | Array of denomination names |
| `rates` | `string[]` | Array of exchange rates |
| `lastUpdate` | `string[]` | Array of last update block heights |
| `lastUpdateTimestamps` | `uint256[]` | Array of last update timestamps |

> **Note:** Results are capped at 1000 entries to prevent unbounded memory allocation.

### `getTwaps(uint256 lookbackSeconds)`

Returns TWAP values for all denominations over the specified lookback window.

| Parameter | Type | Description |
|-----------|------|-------------|
| `lookbackSeconds` | `uint256` | Number of seconds to look back for the TWAP calculation |

**Returns:**

| Name | Type | Description |
|------|------|-------------|
| `denoms` | `string[]` | Array of denomination names |
| `twaps` | `string[]` | Array of TWAP values |

## Usage

```solidity
pragma solidity ^0.8.17;

import "./IOracle.sol";

contract MyDeFiApp {
    IOracle constant ORACLE = IOracle(0x0000000000000000000000000000000000001003);

    /// @notice Get the current price of a token
    function getPrice(string memory denom) external view returns (string memory) {
        (string memory rate, , ) = ORACLE.getExchangeRate(denom);
        return rate;
    }

    /// @notice Get all available prices
    function getAllPrices()
        external
        view
        returns (string[] memory denoms, string[] memory rates)
    {
        (denoms, rates, , ) = ORACLE.getExchangeRates();
    }

    /// @notice Get TWAP over the last hour
    function getHourlyTwap()
        external
        view
        returns (string[] memory denoms, string[] memory twaps)
    {
        (denoms, twaps) = ORACLE.getTwaps(3600);
    }
}
```

## Security Considerations

- **Denom length validation**: Denom strings are capped at 128 bytes (`MaxDenomLength`). Requests with longer denoms are rejected.
- **Result limits**: List queries (`getExchangeRates`) return at most 1000 results (`MaxQueryResults`).
- **Gas metering**: All calls are gas-metered via the standard `KVGasConfig`. Malformed inputs (< 4 bytes) return 0 required gas.

## File Structure

| File | Purpose |
|------|---------|
| `IOracle.sol` | Solidity interface definition |
| `abi.json` | Compiled ABI for runtime decoding |
| `oracle.go` | Precompile entry point, method routing |
| `query.go` | Query handler implementations |
| `types.go` | Argument parsing and validation |
| `types_test.go` | Unit tests for argument parsing |
| `query_test.go` | Unit tests for query handlers |
| `integration_test.go` | Integration tests |

## Further Reading

- [KiiChain Documentation](https://docs.kiiglobal.io/)
- [Oracle Module (`x/oracle`)](../../x/oracle/)
- [Precompiles Overview](../README.md)
