# Kiichain v7.3.0 — Private Coordinated Release

> **Why binaries are attached:** This release ships prebuilt `kiichaind` binaries because the patched EVM dependency lives in the private `github.com/KiiChain/evm-private` mirror. **You cannot build this release from public source** until after Cosmos EVM’s public disclosure window.
>
> **Public build-from-source release:** We will publish a follow-up public release (same binary hashes / equivalent public module path) **1 day after Cosmos EVM’s announced public disclosure date**. Until then, use only the binaries and checksums attached to this GitHub Release.

## Summary

Coordinated `v7.3.0` upgrade that bumps the EVM dependency to
`KiiChain/evm-private v0.6.0-fork.3` (Cosmos Labs–coordinated July 2026 Cosmos EVM hotfix backported onto our fee-abstraction fork). This release is **state-machine-breaking** and must be applied by all validators at the same block height.

## Why private?

Cosmos Labs shared a hotfix with affected chains ahead of public disclosure. The changes are consensus-breaking (precompile gas accounting alignment, StateDB locked-balance snapshotting, and related fixes), so every validator must switch at a coordinated height. Publishing the patch in a public Go module before Cosmos’s disclosure date would leak the fix early, so the dependency is temporarily sourced from `github.com/KiiChain/evm-private`.

## Changes

- chore: update `github.com/cosmos/evm` replacement to `github.com/KiiChain/evm-private v0.6.0-fork.3`.
- feat: add `v7.3.0` coordinated upgrade handler (module migrations only; no store migrations).
- chore: strip completed `v7.2.0` upgrade handler.

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

Build with `make goreleaser-build-local` using `--platform linux/amd64` for reproducible hashes across machines.

| OS     | Arch  | Artifact                          |
| ------ | ----- | --------------------------------- |
| linux  | amd64 | `kiichaind-v7.3.0-linux-amd64`    |
| linux  | arm64 | `kiichaind-v7.3.0-linux-arm64`    |
| darwin | arm64 | `kiichaind-v7.3.0-darwin-arm64`   |
| darwin | amd64 | `kiichaind-v7.3.0-darwin-amd64`   |

Verify downloads against `SHA256SUMS-v7.3.0.txt`.

### Reproduced checksums (linux/amd64 GoReleaser container)

```text
6dd640a9b9943856e5890e9705fb9f8267b80a36d0fb3ad6dd5e4cc64803c62a  kiichaind-darwin-amd64
b91dedaea1e822618ef05f2ac77e9c9baa6a2bbe26b2bc952972a92f8cff8bbf  kiichaind-darwin-arm64
b99a7dd0097fe377c7248a12184d3c0ff1efb1c7886db7e6b366fc1701225b01  kiichaind-linux-amd64
46e553bee97c8bf68b3d5b14885497ad660735e36d4dc0b4c3788a96c662e688  kiichaind-linux-arm64
```
