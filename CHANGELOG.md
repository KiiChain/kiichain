# CHANGELOG

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
