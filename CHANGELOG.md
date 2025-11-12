# CHANGELOG

## Unreleased

## Added

- Cosmos EVM integrations tests added to the repo
- Added logs and telemetry for reentrance detection in wasmd precompile
- Add further validations to Wasm Oracle query bindings
- Moves Wasmd reentrance lock to the core of the wasmd contract to avoid reentrance attacks on queries and instantiations

## Removed

- Removed IBC precompile since ICS20 precompile also handles IBC transfers.
- Removed fallback native price param on the fee abstraction module. Instead of using hardcoded price, the module disables itself when price is lacking.

## DEPENDENCIES

-   Bump [EVM](github.com/cosmos/evm) to [v0.4.2](https://github.com/cosmos/evm/releases/tag/v0.4.2)

## Fixed

- Fix IBC unsafe log as reported on issue [#143](https://github.com/KiiChain/kiichain/issues/143)
- Fix wasmd precompile bad input handling for coins [#147](https://github.com/KiiChain/kiichain/issues/147)
- Fix Oracle module ante decorator to allow first vote and install oracle ante handlers
- Fix IBC validation of negative numbers happens a bit early [#144](https://github.com/KiiChain/kiichain/issues/144)
- Fix incorrect error passing on tokenfactory wasmbinding
- Fix account balance information being stale on fee abstraction's cosmos' ante

### Documentation

- Add further docs related to the reentrance key creation

## v5.1.0 — 2025-10-15

### Fixes

- Fix wasmd precompile against reentrance attacks

## v5.0.0 — 2025-09-26
- Swapped default micro oracle coins to their normal denom
- Correct some doc strings mentioning mint module

## DEPENDENCIES

-   Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.53.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.53.4)
-   Bump [Wasmd](https://github.com/CosmWasm/wasmd) to [v0.61.2](https://github.com/CosmWasm/wasmd/releases/tag/v0.61.2)
-   Bump [EVM](github.com/cosmos/evm) to [v0.4.1](https://github.com/cosmos/evm/releases/tag/v0.4.1)
-   Bump [IBC-go](https://github.com/cosmos/ibc-go/) to [v10.3.0](https://github.com/cosmos/ibc-go/releases/tag/v10.3.0)
-   Removed crisis module

### Fixes

-   Add missing address validation in `GetTokenfactoryDenomsByCreator` query to prevent potential crashes with malformed addresses
-   Override EVM chain ID if default
-   Change evm chain coin info mapping to always default

## v4.0.0 — 2025-08-06

### Added

-   Add the fee abstraction module to the chain

## v3.0.0 — 2025-07-01

No changes were made since the release candidate.

## v3.0.0-rc1 -- 2025-06-25

### Added

-   Add the oracle module to the chain
-   Add the oracle wasmbinding
-   Add the oracle EMV precompile
-   Add E2E tests to IBC precompile
-   Add E2E tests to wasmd precompile

## v2.0.0 -- 2025-06-18

### Added

-   Initial chain creation
-   Add EVM wasmbinding queries
-   Add bech32 wasmbinding queries
-   Add IBC precompile to transfer via EVM
-   Add correct ibc keepers to ibc precompiles
-   Add Rewards module

### Changed

-   Update pipelines by adding codeql, codecov and changelog diff checker
-   Refactor the tokenfactory wasmbinding into its own path
-   Refactor the wasmbinding implementation to allow multiple msg and query types
