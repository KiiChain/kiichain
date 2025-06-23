# CHANGELOG

## UNRELEASED

## v2.0.0 -- 2025-06-18

### Added

- Initial chain creation
- Add EVM wasmbinding queries
- Add bech32 wasmbinding queries
- Add IBC precompile to transfer via EVM
- Add bech32 wasmbinding queries
- Add correct ibc keepers to ibc precompiles
- Add Rewards module

### Changed
- Update pipelines by adding codeql, codecov and changelog diff checker
- Refactor the tokenfactory wasmbinding into its own path
- Refactor the wasmbinding implementation to allow multiple msg and query types
- Add E2E tests to IBC precompile