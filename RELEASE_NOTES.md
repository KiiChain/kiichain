# Kiichain v7.3.0

## Summary

Coordinated `v7.3.0` upgrade that bumps the EVM dependency to
[`KiiChain/evm v0.6.1-fork.1`](https://github.com/KiiChain/evm/releases/tag/v0.6.1-fork.1), carrying the July 2026 Cosmos EVM hotfix ([cosmos/evm v0.6.1](https://github.com/cosmos/evm/releases/tag/v0.6.1)) on top of our fee-abstraction fork. This release is **state-machine-breaking** and must be applied by all validators at the same block height.

The hotfix was developed under Cosmos Labs' coordinated-disclosure process and is now public upstream, so this release builds from public source — `make build` and `make install` work without private module access.

## Changes

### Coordinated upgrade

- chore: update `github.com/cosmos/evm` replacement to `github.com/KiiChain/evm v0.6.1-fork.1`.
- feat: add `v7.3.0` coordinated upgrade handler (module migrations only; no store migrations).
- chore: strip completed `v7.2.0` upgrade handler.

### Kiichain ([KiiChain/kiichain](https://github.com/KiiChain/kiichain))

- [#341](https://github.com/KiiChain/kiichain/pull/341) fix: tokenfactory metadata size limit
- [#340](https://github.com/KiiChain/kiichain/pull/340) fix: distribution precompile 32-byte withdraw address inflates native supply
- [#344](https://github.com/KiiChain/kiichain/pull/344) fix: MsgVoteWeighted and Nested Authz Bypass Vote Stake Requirements
- [#343](https://github.com/KiiChain/kiichain/pull/343) fix: Feegrant Denomination Bypass via Post-Allowance Fee Conversion
- [#342](https://github.com/KiiChain/kiichain/pull/342) fix: add authz guarded router
- [#345](https://github.com/KiiChain/kiichain/pull/345) fix: cosmwasm evm query path repeatable undercharged evm exec
- [#346](https://github.com/KiiChain/kiichain/pull/346) fix: prevent rewards BeginBlocker chain halt on bank transfer failure
- [#347](https://github.com/KiiChain/kiichain/pull/347) fix: Reward Pool Exhaustion via Forced Minimum 1-Unit-Per-Block Release
- [#350](https://github.com/KiiChain/kiichain/pull/350) fix: EIP-7702 Delegated EOAs Permanently Locked Out
- [#352](https://github.com/KiiChain/kiichain/pull/352) fix: Oracle Slashing Bypass via voteTargets Map Mutation
- [#354](https://github.com/KiiChain/kiichain/pull/354) fix: Unweighted Oracle Standard Deviation Inflates Reward Band
- [#353](https://github.com/KiiChain/kiichain/pull/353) fix: Expedited Governance Proposal Whitelist Bypass via authz Wrapping
- [#365](https://github.com/KiiChain/kiichain/pull/365) chore: bump evm fork

### EVM ([KiiChain/evm](https://github.com/KiiChain/evm))

- [#16](https://github.com/KiiChain/evm/pull/16) fix: distribution precompile 32-byte withdraw address inflates native supply
- [#17](https://github.com/KiiChain/evm/pull/17) fix: evm fees refund
- [#18](https://github.com/KiiChain/evm/pull/18) fix: cosmwasm evm query path repeatable undercharged evm exec
- [#19](https://github.com/KiiChain/evm/pull/19) fix: port July 2026 Cosmos EVM hotfix (upstream `v0.6.1`)

## Upgrade

- Upgrade name: `v7.3.0`
- Consensus-breaking: **yes** — every validator must be running this binary by the upgrade height.
- Mechanism: governance `software-upgrade` proposal scheduling height per network; cosmovisor swaps the binary at the height.
- Store migrations: none required.

## Rollout order

1. Plata (devnet) — done
2. Oro (testnet) — done, applied at height 33,659,735
3. Mainnet — governance proposal

## Binaries

Validators can either build from source or use the prebuilt linux binaries attached to this release.

| Role | OS | Arch | Artifact |
| ---- | -- | ---- | -------- |
| **Validator (recommended)** | linux | amd64 | `kiichaind-linux-amd64` |
| Validator (ARM hosts) | linux | arm64 | `kiichaind-linux-arm64` |
| Local Mac only (not for validators) | darwin | amd64 / arm64 | `kiichaind-darwin-*` |

Verify downloads against the `SHA256SUMS` file attached to this release, and confirm `kiichaind version` reports the released tag before staging the binary in cosmovisor under the upgrade name `v7.3.0`.
