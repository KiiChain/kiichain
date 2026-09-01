# Mainnet upgrade to v7.4.0

Validator instructions for the coordinated restart of `kiichain_1783-1` after the 22 August 2026 incident.

**Contents**

- [What happened](#what-happened)
- [What is next](#what-is-next)
- [Install the binary with Cosmovisor](#install-the-binary-with-cosmovisor)
  - [Confirm you are still at the halt height](#confirm-you-are-still-at-the-halt-height)
  - [Back up the node home before you touch Cosmovisor](#back-up-the-node-home-before-you-touch-cosmovisor)
  - [Download and verify the official binary](#download-and-verify-the-official-binary)
  - [Register the off-chain upgrade](#register-the-off-chain-upgrade)
  - [Verify before the start time](#verify-before-the-start-time)
  - [Start with the set](#start-with-the-set)
- [Wrong block hash during the upgrade](#wrong-block-hash-during-the-upgrade)
  - [Why it happens on this upgrade](#why-it-happens-on-this-upgrade)
  - [What to do](#what-to-do)
- [Restart the upgrade after any other error](#restart-the-upgrade-after-any-other-error)
- [What not to do](#what-not-to-do)
- [Support](#support)

This is an **off-chain upgrade**. There is no governance proposal. The `v7.4.0` binary schedules and applies the upgrade plan itself at the announced height. Every validator must start that binary together.

Confirm the values in the table below against the validator announcement before you restart. Do not guess a height.

| Item | Value |
| --- | --- |
| Chain ID | `kiichain_1783-1` |
| Upgrade name | `v7.4.0` |
| Last committed height (halt) | _to be confirmed in the announcement_ |
| Upgrade height (first new block) | _to be confirmed in the announcement_ |
| Coordinated start time (UTC) | _to be confirmed in the announcement_ |
| Release | https://github.com/KiiChain/kiichain/releases/tag/v7.4.0 |
| Linux amd64 binary | `kiichaind-v7.4.0-linux-amd64` |
| Checksums | `SHA256SUMS-v7.4.0.txt` on the same release |

The release is prepared from [kiichain#375](https://github.com/KiiChain/kiichain/pull/375).

---

<a id="what-happened"></a>

## 1. What happened

On 22 August 2026 an attacker exploited defects in the shared Cosmos EVM module, not in KiiChain-specific application code. The path depended on a vesting account and chained:

- an arithmetic underflow in the staking precompile's balance write-back after a delegation
- a missing overflow guard on the EVM value-transfer credit path

Together those bugs let the attacker mint and move native KII that the bank ledger did not authorize. Once the incident was confirmed, block production was halted so no further funds could leave the chain.

Funds that were already bridged off-chain are outside this upgrade. Funds that were still sitting in attacker-controlled addresses on Kiichain at the halt will be recovered by `v7.4.0` itself when the first new block is applied.

---

<a id="what-is-next"></a>

## 2. What is next

`v7.4.0` is the resumption binary. At the upgrade height, on mainnet only, it will:

1. Sweep remaining balances from the confirmed attacker-controlled addresses into a staging account.
2. Redistribute those recovered funds to the designated recovery wallets.
3. Permanently reject bank sends to or from those attacker addresses (Cosmos, precompile, and EVM native commits).
4. Keep vesting / permanently-locked account creation disabled, so the precondition the exploit used cannot be opened again.

The same binary is safe to run on testnet or a local rehearsal: fund recovery is a no-op off mainnet. The EVM module defects themselves are patched in this release.

**What validators must do**

1. Keep the node stopped until the coordinated start time.
2. Take (or confirm) a backup of the whole `$DAEMON_HOME` at the halt height. You need this if the first new block fails.
3. Install the official `v7.4.0` binary into Cosmovisor using the steps below.
4. Confirm Cosmovisor will start `v7.4.0` for the first new block. Do not resume the pre-incident binary.
5. Start with the rest of the set at the announced time.
6. Confirm the upgrade height applied and that block production continues.

Do not submit or wait for a software-upgrade proposal. Governance cannot run while the chain is halted, and this binary does not expect one.

---

<a id="install-the-binary-with-cosmovisor"></a>

## 3. Install the binary with Cosmovisor

Cosmovisor (sometimes written cosmosvisor) is the process manager already used by the [mainnet join script](https://github.com/KiiChain/mainnets/blob/main/kiichain/join_kiichain_cv.sh). These steps assume:

```bash
export DAEMON_NAME=kiichaind
export DAEMON_HOME="$HOME/.kiichain"   # change if your home is different
export DAEMON_ALLOW_DOWNLOAD_BINARIES=false
```

Disable auto-download for this upgrade. Use only the official release binary.

<a id="confirm-you-are-still-at-the-halt-height"></a>

### 3.1 Confirm you are still at the halt height

```bash
# If the node is still running, this should match the announced last committed height.
curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height'
```

If RPC is down, use the last height from `journalctl -u kiichain` (or your unit name). You want the height of the last committed block before the halt, not a height you produced locally after experimenting.

If your local height is **below** the halt height, sync from a peer or snapshot that is still on the pre-upgrade binary, then stop again. If your local height is **above** the halt height, stop and contact the team before you start `v7.4.0`.

<a id="back-up-the-node-home-before-you-touch-cosmovisor"></a>
<a id="back-up-data-before-you-touch-cosmovisor"></a>

### 3.2 Back up the node home before you touch Cosmovisor

Copy the whole `$DAEMON_HOME` (`data`, `wasm`, `config`, Cosmovisor layout), not only `data`. Keep that copy on the same machine or a private disk. Do not upload `config/priv_validator_key.json` or `config/node_key.json` anywhere public.

```bash
sudo systemctl stop kiichain   # use your unit name
mkdir -p "$HOME/kiichain-backup"
cp -a "$DAEMON_HOME" "$HOME/kiichain-backup/home-pre-v7.4.0"
```

If disk is tight, at minimum copy `data` (and `wasm` if it sits next to `data`). A full home copy is the safer default.

<a id="download-and-verify-the-official-binary"></a>

### 3.3 Download and verify the official binary

Public source cannot reproduce this build until the EVM hotfix is published. Install from the GitHub release, not from `make install` on `main`.

```bash
cd /tmp
curl -LO https://github.com/KiiChain/kiichain/releases/download/v7.4.0/kiichaind-v7.4.0-linux-amd64
curl -LO https://github.com/KiiChain/kiichain/releases/download/v7.4.0/SHA256SUMS-v7.4.0.txt
sha256sum -c SHA256SUMS-v7.4.0.txt --ignore-missing
chmod +x kiichaind-v7.4.0-linux-amd64
./kiichaind-v7.4.0-linux-amd64 version
# Expect: 7.4.0
```

Use the `linux-arm64` asset if that is your host architecture.

<a id="register-the-off-chain-upgrade"></a>

### 3.4 Register the off-chain upgrade

This writes the binary to Cosmovisor and records the upgrade height **without** an on-chain proposal:

```bash
# Cosmovisor must see DAEMON_NAME and DAEMON_HOME (systemd already sets these
# for the service; export them in this shell too).
cosmovisor add-upgrade v7.4.0 /tmp/kiichaind-v7.4.0-linux-amd64 \
  --upgrade-height <UPGRADE_HEIGHT> \
  --force
```

`<UPGRADE_HEIGHT>` is the height from the validator announcement (the first block after the halt).

That command:

- copies the binary to `$DAEMON_HOME/cosmovisor/upgrades/v7.4.0/bin/kiichaind`
- writes `$DAEMON_HOME/data/upgrade-info.json` so Cosmovisor switches at that height

**This chain is already halted.** The recovery handler lives only in `v7.4.0` and must run on that first new block. After `add-upgrade`, make Cosmovisor's `current` link point at `v7.4.0` so you do not accidentally start the old binary:

```bash
ln -sfn "$DAEMON_HOME/cosmovisor/upgrades/v7.4.0" "$DAEMON_HOME/cosmovisor/current"
```

If you prefer not to use `add-upgrade`, the equivalent manual layout is (this is what Kii-operated nodes will use):

```bash
mkdir -p "$DAEMON_HOME/cosmovisor/upgrades/v7.4.0/bin"
cp /tmp/kiichaind-v7.4.0-linux-amd64 "$DAEMON_HOME/cosmovisor/upgrades/v7.4.0/bin/kiichaind"
chmod 755 "$DAEMON_HOME/cosmovisor/upgrades/v7.4.0/bin/kiichaind"
ln -sfn "$DAEMON_HOME/cosmovisor/upgrades/v7.4.0" "$DAEMON_HOME/cosmovisor/current"
```

Cosmovisor v1.5.0 (the version the join script installs) can treat a pre-written `upgrade-info.json` as "switch now". That is acceptable here **only if** the binary it switches to is `v7.4.0`. It is not acceptable if it leaves you running the pre-incident binary through the upgrade height.

<a id="verify-before-the-start-time"></a>

### 3.5 Verify before the start time

```bash
readlink -f "$DAEMON_HOME/cosmovisor/current"
# .../cosmovisor/upgrades/v7.4.0

"$DAEMON_HOME/cosmovisor/current/bin/kiichaind" version
# 7.4.0

sha256sum "$DAEMON_HOME/cosmovisor/current/bin/kiichaind"
# must match SHA256SUMS-v7.4.0.txt
```

Turn off state sync for the restart. The join script often leaves `[statesync] enable = true`. You already have halt-height state and must apply the first new block locally. If state sync stays on, CometBFT can throw that state away and fetch from an RPC using a `trust_height` / `trust_hash` from before the halt (or from a peer that has already diverged). There is no live chain to trust until this set produces the upgrade block.

```toml
# $DAEMON_HOME/config/config.toml
[statesync]
enable = false
```

<a id="start-with-the-set"></a>

### 3.6 Start with the set

At the announced UTC time:

```bash
sudo systemctl start kiichain
journalctl -fu kiichain
```

You should see the emergency recovery logs on the upgrade height, then normal block production. Confirm:

```bash
curl -s http://localhost:26657/status | jq '.result.sync_info | {latest_block_height, catching_up}'
```

`latest_block_height` should move past the upgrade height and `catching_up` should be `false` for a validator that is in the active set.

---

<a id="wrong-block-hash-during-the-upgrade"></a>

## 4. Wrong block hash during the upgrade

A wrong block hash means your node computed a different app state than the block you are trying to apply. Typical log lines:

```text
wrong Block.Header.AppHash. Expected <hash>, got <hash>
wrong Block.Header.LastResultsHash
CONSENSUS FAILURE!!! err=wrong Block.Header.AppHash
```

On a Cosmovisor / CometBFT node this is the same class of error whether the log says "block hash" or `AppHash`.

<a id="why-it-happens-on-this-upgrade"></a>

### Why it happens on this upgrade

- The node started the **pre-incident** binary for the first new block, so it did not run the recovery handler.
- The node was not actually at the halt height (missing blocks, or it had already applied a partial / local upgrade block).
- A different `v7.4.0` binary was used (self-built, wrong arch, checksum ignored).
- A leftover `$DAEMON_HOME/data/upgrade-info.json` was applied against data that had already been migrated, or the opposite: migrated data with the old binary.
- State sync was left on and the node used a `trust_height` / `trust_hash` that does not exist on the restarted chain.

Do not keep restarting the service. Each retry can move `current` or rewrite `upgrade-info.json` and make a clean rollback harder.

<a id="what-to-do"></a>

### What to do

1. Stop the node.

   ```bash
   sudo systemctl stop kiichain
   ```

2. Restore **data** from the pre-upgrade backup you took in §3.2. Do not restore the whole home over the live node: that would also revert the `v7.4.0` binary you just installed. Leave the live `config/priv_validator_key.json` and `config/node_key.json` alone.

   ```bash
   rm -rf "$DAEMON_HOME/data"
   cp -a "$HOME/kiichain-backup/home-pre-v7.4.0/data" "$DAEMON_HOME/data"
   # if you keep wasm outside data:
   # rm -rf "$DAEMON_HOME/wasm"
   # cp -a "$HOME/kiichain-backup/home-pre-v7.4.0/wasm" "$DAEMON_HOME/wasm"
   ```

   If Cosmovisor wrote its own backup under `$DAEMON_HOME/cosmovisor/backup` during a failed switch, use the snapshot whose height is the halt height, not a copy taken after the failed block.

3. Remove the upgrade marker so Cosmovisor / `x/upgrade` do not think the upgrade already ran.

   ```bash
   rm -f "$DAEMON_HOME/data/upgrade-info.json"
   ```

4. Point `current` at `v7.4.0` again and re-check the binary.

   ```bash
   ln -sfn "$DAEMON_HOME/cosmovisor/upgrades/v7.4.0" "$DAEMON_HOME/cosmovisor/current"
   "$DAEMON_HOME/cosmovisor/current/bin/kiichaind" version
   sha256sum "$DAEMON_HOME/cosmovisor/current/bin/kiichaind"
   ```

5. Confirm local height is the halt height, state sync is off, then start with the set again.

`kiichaind rollback` is **not** a substitute for that backup. It only undoes the last *committed* height. If the upgrade block never committed (typical `CONSENSUS FAILURE` / `wrong AppHash` while still at the halt height), rollback would delete the last good block. If you did commit exactly one extra height, rollback can return you to the halt height, but it does not fix Cosmovisor `current` or `$DAEMON_HOME/data/upgrade-info.json` — you still have to do steps 3–4. Prefer restoring the halt-height copy.

If you have no halt-height backup, do not run `unsafe-reset-all` on a validator. Ask the team for a snapshot taken at the halt height and restore that, then start `v7.4.0`.

A wrong **state-sync** `trust_hash` (join-script style) is a different failure: the node never had halt-height state. Fix it by disabling state sync and restoring a halt-height snapshot, not by editing `trust_hash` against a half-upgraded RPC.

---

<a id="restart-the-upgrade-after-any-other-error"></a>

## 5. Restart the upgrade after any other error

Use this if Cosmovisor exits, the handler panics, the process crash-loops, or you interrupted the first block.

1. **Stop.** `sudo systemctl stop kiichain`
2. **Read the panic / Cosmovisor line.** If you see `UPGRADE_NEEDED`, `upgrade <name> is already executed`, or `applied plan`, say so in the validator channel before you delete files.
3. **Restore halt-height `data`** (same commands as §4).
4. **Delete** `$DAEMON_HOME/data/upgrade-info.json` so the plan can be scheduled again on the first new block.
5. **Do not delete** `$DAEMON_HOME/cosmovisor/upgrades/v7.4.0`. Re-run `cosmovisor add-upgrade ... --force` only if the binary in that folder is missing or has the wrong checksum.
6. **Relink** `current` → `upgrades/v7.4.0`.
7. **Start** only when your height, binary, and checksum match the rest of the set.

The upgrade handler is designed to run **once**, on the first block at the upgrade height, against halt-height state. Replaying it against data that already includes that block will fail or diverge. Always roll back to the halt height, then start `v7.4.0` again.

---

<a id="what-not-to-do"></a>

## 6. What not to do

- Do not start the pre-incident binary and wait for Cosmovisor to "flip later". The first new block would not run recovery, and the set would split on app hash.
- Do not `kiichaind comet unsafe-reset-all` or wipe `priv_validator_key.json`.
- Do not build `v7.4.0` from public GitHub source until the team says the EVM hotfix is public.
- Do not enable `DAEMON_ALLOW_DOWNLOAD_BINARIES` for this upgrade.
- Do not change the upgrade name. Cosmovisor's folder and the in-binary plan must both be `v7.4.0`.

---

<a id="support"></a>

## 7. Support

Post in the validator channel with:

- `kiichaind version` from `$DAEMON_HOME/cosmovisor/current/bin/kiichaind`
- `sha256sum` of that binary
- `readlink -f $DAEMON_HOME/cosmovisor/current`
- the last 80 lines of `journalctl -u kiichain`
- `latest_block_height` from `/status` if RPC still answers
