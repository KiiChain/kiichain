# Upgrade to v9.0.0

Operator instructions for the `v9.0.0` software upgrade. This release moves Kiichain onto cosmos/evm v0.7.3, cosmos-sdk v0.54, CometBFT v0.39, ibc-go v11, and wasmd v0.70.

This is a governance upgrade. The chain halts at the height in the `MsgSoftwareUpgrade` plan named `v9.0.0`, and every node switches to the `v9.0.0` binary at that height.

| Item | Value |
| --- | --- |
| Upgrade name | `v9.0.0` |
| Upgrade height | _from the governance proposal_ |
| Go version (source builds) | 1.25.9 |
| Required `config.toml` change | `[mempool] type = "app"` |

## 1. Install the binary

```bash
export DAEMON_NAME=kiichaind
export DAEMON_HOME="$HOME/.kiichain"   # change if your home is different

mkdir -p "$DAEMON_HOME/cosmovisor/upgrades/v9.0.0/bin"
cp kiichaind-v9.0.0-linux-amd64 "$DAEMON_HOME/cosmovisor/upgrades/v9.0.0/bin/kiichaind"
chmod 755 "$DAEMON_HOME/cosmovisor/upgrades/v9.0.0/bin/kiichaind"
"$DAEMON_HOME/cosmovisor/upgrades/v9.0.0/bin/kiichaind" version
```

## 2. Switch the CometBFT mempool type at the halt

`v9.0.0` uses the Krakatoa EVM mempool, which needs the CometBFT application-side mempool. The binary refuses to start while `config.toml` still has the `"flood"` default:

```text
EVM mempool enabled, but comet-bft has invalid config.toml:mempool.type (want 'app', got 'flood')
```

The pre-upgrade binary (CometBFT v0.38) does not accept `"app"`, so make this change **after the node halts at the upgrade height and before `v9.0.0` starts**.

Set `DAEMON_RESTART_AFTER_UPGRADE=false` in the Cosmovisor service before the upgrade height so Cosmovisor switches the binary and exits instead of starting it right away. Once the node logs `UPGRADE "v9.0.0" NEEDED` and stops:

```bash
sed -i 's/^type = "flood"/type = "app"/' "$DAEMON_HOME/config/config.toml"
grep -A2 '^\[mempool\]' "$DAEMON_HOME/config/config.toml"
# [mempool]
# ...
# type = "app"

sudo systemctl start kiichain   # use your unit name
```

If Cosmovisor or a systemd `Restart=` policy already started `v9.0.0` and the node exits with the error above, stop the service, apply the same `sed`, and start it again. The check runs before any block is applied, so no rollback is needed.

`app.toml` needs no edits. Missing `[evm.mempool]` keys fall back to the defaults.

New nodes created with `kiichaind init` from `v9.0.0` already have `type = "app"`.

## 3. Verify

```bash
readlink -f "$DAEMON_HOME/cosmovisor/current"
# .../cosmovisor/upgrades/v9.0.0

curl -s http://localhost:26657/status | jq '.result.sync_info | {latest_block_height, catching_up}'
```

The node should log `Upgrade v9.0.0 complete` at the upgrade height and keep producing blocks.
