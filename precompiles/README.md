# KiiChain Precompiles

KiiChain ships with EVM precompiled contracts that expose Cosmos SDK module functionality to Solidity smart contracts. These precompiles are deployed at fixed addresses and can be called like any other contract.

## Available Precompiles

| Precompile | Address | Description |
|------------|---------|-------------|
| [Oracle](./oracle/) | `0x0000000000000000000000000000000000001003` | Query exchange rates, TWAP prices from the oracle module |
| [Staking](./staking/) | `0x0000000000000000000000000000000000000800` | Delegate, undelegate, and query delegation amounts |

## Architecture

Each precompile follows the same pattern:

1. **Solidity Interface** (`I<Name>.sol`) — defines the EVM-callable functions
2. **ABI** (`abi.json`) — the compiled ABI used to decode calls at runtime
3. **Go Implementation** (`<name>.go`) — implements `vm.PrecompiledContract`, routing method IDs to handler functions
4. **Query/Tx Handlers** (`query.go` / `tx.go`) — execute the underlying Cosmos SDK keeper calls
5. **Type Parsers** (`types.go`) — parse and validate ABI-decoded arguments

## Utilities

- **`common/`** — shared Solidity types (`Types.sol`) used across precompiles
- **`scripts/`** — tooling to convert Solidity compiler output to Hardhat-compatible artifacts

## Security

- All precompile inputs are validated (e.g., denom length capped at 128 bytes)
- Query results are bounded to prevent unbounded memory allocation
- Gas metering follows the `KVGasConfig` / `TransientGasConfig` from the Cosmos SDK store

## Further Reading

- [KiiChain Documentation](https://docs.kiiglobal.io/)
- [Cosmos EVM Precompiles](https://github.com/cosmos/evm/tree/main/precompiles)
