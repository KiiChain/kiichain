# CHANGELOG

## UNRELEASED

## DEPENDENCIES
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.53.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.53.4)
- Bump [Wasmd](https://github.com/CosmWasm/wasmd) to [v0.61.2](https://github.com/CosmWasm/wasmd/releases/tag/v0.61.2)
- Bump [EVM](github.com/cosmos/evm) to [v0.4.1](https://github.com/cosmos/evm/releases/tag/v0.4.1)
- Bump [IBC-go](https://github.com/cosmos/ibc-go/) to [v10.3.0](https://github.com/cosmos/ibc-go/releases/tag/v10.3.0)
- Removed crisis module 

### Fixes

- Add missing address validation in `GetTokenfactoryDenomsByCreator` query to prevent potential crashes with malformed addresses


## v4.0.0 — 2025-08-06

### Added

- Add the fee abstraction module to the chain

## v3.0.0 — 2025-07-01

No changes were made since the release candidate.

## v3.0.0-rc1 -- 2025-06-25

### Added

- Add the oracle module to the chain
- Add the oracle wasmbinding
- Add the oracle EMV precompile
- Add E2E tests to IBC precompile
- Add E2E tests to wasmd precompile

## v2.0.0 -- 2025-06-18

### Added

- Initial chain creation
- Add EVM wasmbinding queries
- Add bech32 wasmbinding queries
- Add IBC precompile to transfer via EVM
- Add correct ibc keepers to ibc precompiles
- Add Rewards module

### Changed

- Update pipelines by adding codeql, codecov and changelog diff checker
- Refactor the tokenfactory wasmbinding into its own path
- Refactor the wasmbinding implementation to allow multiple msg and query types
