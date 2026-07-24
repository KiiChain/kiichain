# Kiichain v7.3.0 — Private Coordinated Release

> **Why binaries are attached:** Validators cannot build this release from public source. The patched EVM dependency is temporarily published only in the private `github.com/KiiChain/evm-private` mirror (`v0.6.0-fork.3`), following Cosmos Labs’ coordinated-disclosure process for the July 2026 Cosmos EVM hotfix. We are attaching prebuilt `kiichaind` binaries (and checksums) so operators can upgrade without needing private module access.
>
> **Cannot build from source (yet):** `go mod download` / `make build` against public GitHub will fail until the dependency is flipped back to a public module path.
>
> **Public build-from-source release:** A follow-up public release (public `KiiChain/evm` / upstream module path, with its own checksums) will be published **1 day after Cosmos EVM’s announced public disclosure date**.  
> - Cosmos EVM public disclosure date: **July 27, 2026**  
> - Kiichain public build-from-source release target: **July 28, 2026**  
>
> Until that public release, use **only** the binaries and `SHA256SUMS-v7.3.0.txt` attached to this GitHub Release.

## Summary

Coordinated `v7.3.0` upgrade that bumps the EVM dependency to
`KiiChain/evm-private v0.6.0-fork.3` (Cosmos Labs–coordinated July 2026 Cosmos EVM hotfix backported onto our fee-abstraction fork). This release is **state-machine-breaking** and must be applied by all validators at the same block height.

## Why private?

Cosmos Labs shared a hotfix with affected chains ahead of public disclosure. The changes are consensus-breaking (precompile gas accounting alignment, StateDB locked-balance snapshotting, and related fixes), so every validator must switch at a coordinated height. Publishing the patch in a public Go module before Cosmos’s disclosure date would leak the fix early, so the dependency is temporarily sourced from `github.com/KiiChain/evm-private`.

## Changes

### Coordinated upgrade

- chore: update `github.com/cosmos/evm` replacement to `github.com/KiiChain/evm-private v0.6.0-fork.3`.
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

## Upgrade

- Upgrade name: `v7.3.0`
- Consensus-breaking: **yes** — every validator must be running this binary by the upgrade height.
- Mechanism: governance `software-upgrade` proposal scheduling height per network; cosmovisor swaps the binary at the height.
- Store migrations: none required.

## Rollout order

1. Plata (devnet) — done
2. Oro (testnet) — governance proposal
3. Mainnet — governance proposal

## Binaries

**Validators: use the Linux prebuilt binaries only.** Do not build from public source for this release, and do not use Darwin artifacts on validator nodes.

Release linux/darwin hashes were reproduced from tag `v7.3.0` (`cedbf48`) via GoReleaser release mode (no `--snapshot`). Verify downloads against `SHA256SUMS-v7.3.0.txt` below.

| Role | OS | Arch | Artifact |
| ---- | -- | ---- | -------- |
| **Validator (recommended)** | linux | amd64 | `kiichaind-v7.3.0-linux-amd64` |
| Validator (ARM hosts) | linux | arm64 | `kiichaind-v7.3.0-linux-arm64` |
| Local Mac only (not for validators) | darwin | amd64 / arm64 | `kiichaind-v7.3.0-darwin-*` |

Verify downloads against `SHA256SUMS-v7.3.0.txt`. After install, `kiichaind version` should report `v7.3.0` (commit `cedbf486bcdf1d7744b7506e7f3880eb21bd61c3`).

### Checksums (`SHA256SUMS-v7.3.0.txt`)

```text
be9f08d4d04d4d2f6eb576c59415a7bd8965f03d8893e4a95c25b22584f3bc44  kiichaind-v7.3.0-linux-amd64
b65d99295c8ead7708ceed21a3c1f36e096a0c7160f939fe6dbc36b64f08875d  kiichaind-v7.3.0-linux-arm64
c8de862ae2a57801c03165acab8b002b283513e14d39cc791b2e27fa70af9e6e  kiichaind-v7.3.0-darwin-amd64
f544123d3a7f4db6f5580b0170c7a822ae4f625834d577761637765bf8a0b5dc  kiichaind-v7.3.0-darwin-arm64
```
