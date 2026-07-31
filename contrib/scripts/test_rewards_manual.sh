#!/usr/bin/env bash
# Manual smoke test for inflation-based x/rewards emissions.
#
# Flow:
#   1. Build + init a single-validator localnet (short gov voting period)
#   2. Fund the rewards pool
#   3. Pass MsgUpdateParams with supply_base > 0
#   4. Wait a few blocks and assert pool decreased / total_released increased
#
# Usage:
#   ./contrib/scripts/test_rewards_manual.sh
#   KEEP_HOME=1 ./contrib/scripts/test_rewards_manual.sh   # leave node home for inspection
#
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

BIN="${BIN:-$ROOT_DIR/build/kiichaind}"
HOME_DIR="${HOME_DIR:-$HOME/.kiichaind-rewards-manual}"
CHAIN_ID="${CHAIN_ID:-localchain_1010-1}"
KEY="${KEY:-val}"
DENOM="akii"
RPC_URL="http://127.0.0.1:26657"
NODE="tcp://127.0.0.1:26657"
# Standard gov module account (bech32 "kii" prefix)
GOV_AUTHORITY="kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv"

FUND_AMOUNT="1000000000000000000000${DENOM}"       # 1000 kii
SUPPLY_BASE="1000000000000000000000000"            # 1e24
DEPOSIT_AMOUNT="10000000${DENOM}"
GAS_PRICES="3000000000${DENOM}"

log() { printf '\n==> %s\n' "$*"; }
die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

broadcast_tx() {
  local out code txhash
  out="$("$@" --broadcast-mode sync -o json)"
  code="$(echo "$out" | jq -r '.code // 0')"
  txhash="$(echo "$out" | jq -r '.txhash // empty')"
  if [[ "$code" != "0" ]]; then
    echo "$out" | jq . >&2 || echo "$out" >&2
    die "tx rejected at broadcast (code=$code)"
  fi
  [[ -n "$txhash" ]] || die "missing txhash from broadcast"
  # Wait for inclusion
  local i status
  for i in $(seq 1 30); do
    if status="$("$BIN" q tx "$txhash" --node "$NODE" -o json 2>/dev/null)"; then
      code="$(echo "$status" | jq -r '.code // 0')"
      if [[ "$code" != "0" ]]; then
        echo "$status" | jq . >&2
        die "tx $txhash failed on-chain (code=$code)"
      fi
      return 0
    fi
    sleep 1
  done
  die "timed out waiting for tx $txhash"
}

cleanup() {
  if [[ -n "${NODE_PID:-}" ]] && kill -0 "$NODE_PID" 2>/dev/null; then
    log "Stopping local node (pid $NODE_PID)"
    kill "$NODE_PID" 2>/dev/null || true
    wait "$NODE_PID" 2>/dev/null || true
  fi
  if [[ "${KEEP_HOME:-0}" != "1" ]]; then
    rm -rf "$HOME_DIR"
  else
    log "Keeping home dir at $HOME_DIR (KEEP_HOME=1)"
  fi
}
trap cleanup EXIT

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

need_cmd jq
need_cmd curl
need_cmd python3

log "Building kiichaind"
make build >/dev/null
[[ -x "$BIN" ]] || die "binary not found at $BIN"

# Free ports if a previous run left a node up
pkill -f "kiichaind.*${HOME_DIR}" >/dev/null 2>&1 || true
sleep 1
rm -rf "$HOME_DIR"

log "Initializing chain at $HOME_DIR"
"$BIN" init rewards-manual --chain-id "$CHAIN_ID" --home "$HOME_DIR" >/dev/null 2>&1
"$BIN" config set client chain-id "$CHAIN_ID" --home "$HOME_DIR"
"$BIN" config set client keyring-backend test --home "$HOME_DIR"
"$BIN" config set client node "$NODE" --home "$HOME_DIR"

"$BIN" keys add "$KEY" --home "$HOME_DIR" --keyring-backend test >/dev/null 2>&1
"$BIN" genesis add-genesis-account "$KEY" \
  "10000000000000000000000000000000000000${DENOM}" \
  --home "$HOME_DIR" --keyring-backend test >/dev/null

"$BIN" genesis gentx "$KEY" "1000000000000000000000${DENOM}" \
  --home "$HOME_DIR" --chain-id "$CHAIN_ID" --keyring-backend test >/dev/null 2>&1
"$BIN" genesis collect-gentxs --home "$HOME_DIR" >/dev/null 2>&1

GENESIS="$HOME_DIR/config/genesis.json"
APP_TOML="$HOME_DIR/config/app.toml"

# Short gov window + matching EVM denom / metadata (same as start-localnet-ci)
tmp="$(mktemp)"
jq '
  .app_state.gov.params.voting_period = "12s"
  | .app_state.gov.params.expedited_voting_period = "8s"
  | .app_state.gov.params.min_deposit = [{"denom":"akii","amount":"10000000"}]
  | .app_state.gov.params.expedited_min_deposit = [{"denom":"akii","amount":"10000000"}]
  | .app_state.gov.params.quorum = "0.000000000000000001"
  | .app_state.gov.params.threshold = "0.000000000000000001"
  | .app_state.evm.params.evm_denom = "akii"
  | .app_state.bank.denom_metadata = [{
      "description":"The native staking token of the kiichain network",
      "denom_units":[{"denom":"akii","exponent":0},{"denom":"kii","exponent":18}],
      "base":"akii","display":"kii","name":"kii","symbol":"KII"
    }]
' "$GENESIS" > "$tmp" && mv "$tmp" "$GENESIS"

if [[ "$(uname)" == "Darwin" ]]; then
  sed -i.bak 's/minimum-gas-prices = ""/minimum-gas-prices = "0akii"/' "$APP_TOML"
else
  sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0akii"/' "$APP_TOML"
fi
# Keep feemarket from blocking with a high floor relative to our gas prices
tmp="$(mktemp)"
jq '
  .app_state.feemarket.params.no_base_fee = true
  | .app_state.feemarket.params.min_gas_price = "0.000000000000000000"
  | .app_state.feemarket.params.base_fee = "0.000000000000000000"
' "$GENESIS" > "$tmp" && mv "$tmp" "$GENESIS"

log "Starting node"
"$BIN" start --home "$HOME_DIR" >"$HOME_DIR/node.log" 2>&1 &
NODE_PID=$!

# Wait for RPC
for i in $(seq 1 60); do
  if curl -sf "$RPC_URL/status" >/dev/null 2>&1; then
    break
  fi
  sleep 1
  if [[ $i -eq 60 ]]; then
    tail -n 80 "$HOME_DIR/node.log" || true
    die "node did not become ready"
  fi
done
# Wait until height > 1
for i in $(seq 1 30); do
  height="$(curl -sf "$RPC_URL/status" | jq -r '.result.sync_info.latest_block_height')"
  if [[ "$height" =~ ^[0-9]+$ ]] && (( height > 1 )); then
    log "Node ready at height $height"
    break
  fi
  sleep 1
  if [[ $i -eq 30 ]]; then
    die "node never reached height > 1"
  fi
done

TX_FLAGS=(
  --home "$HOME_DIR"
  --keyring-backend test
  --chain-id "$CHAIN_ID"
  --node "$NODE"
  --gas 5000000
  --gas-prices "$GAS_PRICES"
  --yes
)

log "Query initial rewards params / pool"
"$BIN" q rewards params --node "$NODE" -o json | jq .
INITIAL_POOL="$("$BIN" q rewards reward-pool --node "$NODE" -o json)"
echo "$INITIAL_POOL" | jq .
INITIAL_SUPPLY_BASE="$(
  "$BIN" q rewards params --node "$NODE" -o json | jq -r '.params.supply_base // .supply_base // "0"'
)"
[[ "$INITIAL_SUPPLY_BASE" == "0" || "$INITIAL_SUPPLY_BASE" == "0.000000000000000000" ]] \
  || die "expected default supply_base=0, got $INITIAL_SUPPLY_BASE"

log "Fund rewards pool with $FUND_AMOUNT"
broadcast_tx "$BIN" tx rewards fund-pool "$FUND_AMOUNT" --from "$KEY" "${TX_FLAGS[@]}"

FUNDED_POOL="$("$BIN" q rewards reward-pool --node "$NODE" -o json)"
echo "$FUNDED_POOL" | jq .
FUNDED_AMOUNT="$(echo "$FUNDED_POOL" | jq -r '
  (.reward_pool.community_pool // .community_pool // [])
  | map(select(.denom=="akii")) | .[0].amount // "0"
')"
[[ "$FUNDED_AMOUNT" != "0" && "$FUNDED_AMOUNT" != "null" ]] || die "pool still empty after fund-pool"
log "Pool funded amount=$FUNDED_AMOUNT"

PROPOSAL_FILE="$HOME_DIR/proposal_update_rewards_params.json"
cat >"$PROPOSAL_FILE" <<EOF
{
  "messages": [
    {
      "@type": "/kiichain.rewards.v1beta1.MsgUpdateParams",
      "authority": "$GOV_AUTHORITY",
      "params": {
        "token_denom": "akii",
        "goal_bonded": "0.670000000000000000",
        "inflation_min": "0.000000000000000000",
        "inflation_max": "0.200000000000000000",
        "supply_base": "$SUPPLY_BASE"
      }
    }
  ],
  "metadata": "ipfs://CID",
  "deposit": "$DEPOSIT_AMOUNT",
  "title": "Enable Rewards Emissions",
  "summary": "set supply_base to enable inflation-based emissions"
}
EOF

log "Submit + vote MsgUpdateParams (supply_base=$SUPPLY_BASE)"
broadcast_tx "$BIN" tx gov submit-proposal "$PROPOSAL_FILE" --from "$KEY" "${TX_FLAGS[@]}"
PROPOSAL_ID="$("$BIN" q gov proposals --node "$NODE" -o json | jq -r '.proposals | max_by(.id | tonumber) | .id')"
[[ -n "$PROPOSAL_ID" && "$PROPOSAL_ID" != "null" ]] || die "could not find proposal id"
log "Proposal id=$PROPOSAL_ID"
broadcast_tx "$BIN" tx gov vote "$PROPOSAL_ID" yes --from "$KEY" "${TX_FLAGS[@]}"

log "Waiting for proposal to pass"
PASSED=0
for i in $(seq 1 40); do
  STATUS="$("$BIN" q gov proposal "$PROPOSAL_ID" --node "$NODE" -o json | jq -r '.proposal.status // .status')"
  log "Proposal status=$STATUS (wait $i)"
  if [[ "$STATUS" == "PROPOSAL_STATUS_PASSED" || "$STATUS" == "3" ]]; then
    PASSED=1
    break
  fi
  if [[ "$STATUS" == "PROPOSAL_STATUS_REJECTED" || "$STATUS" == "PROPOSAL_STATUS_FAILED" || "$STATUS" == "4" || "$STATUS" == "5" ]]; then
    "$BIN" q gov proposal "$PROPOSAL_ID" --node "$NODE" -o json | jq .
    die "proposal did not pass"
  fi
  sleep 2
done
[[ "$PASSED" -eq 1 ]] || die "timed out waiting for proposal"

UPDATED_SUPPLY_BASE="$(
  "$BIN" q rewards params --node "$NODE" -o json | jq -r '.params.supply_base // .supply_base'
)"
log "Updated supply_base=$UPDATED_SUPPLY_BASE"
[[ "$UPDATED_SUPPLY_BASE" == "$SUPPLY_BASE" ]] || die "supply_base not updated"

# First begin-blocker after enable only stamps last_release_time; wait for releases
log "Waiting for emissions across several blocks"
sleep 12

FINAL_POOL="$("$BIN" q rewards reward-pool --node "$NODE" -o json)"
echo "$FINAL_POOL" | jq .
FINAL_AMOUNT="$(echo "$FINAL_POOL" | jq -r '
  (.reward_pool.community_pool // .community_pool // [])
  | map(select(.denom=="akii")) | .[0].amount // "0"
')"
TOTAL_RELEASED="$(echo "$FINAL_POOL" | jq -r '
  (.reward_pool.total_released.amount // .total_released.amount // "0")
')"
LAST_RELEASE="$(echo "$FINAL_POOL" | jq -r '
  (.reward_pool.last_release_time // .last_release_time // "")
')"

log "Funded amount = $FUNDED_AMOUNT"
log "Final amount  = $FINAL_AMOUNT"
log "Total released= $TOTAL_RELEASED"
log "Last release  = $LAST_RELEASE"

# Compare as integers (strip decimal portion from DecCoin strings)
python3 - "$FUNDED_AMOUNT" "$FINAL_AMOUNT" "$TOTAL_RELEASED" <<'PY'
import sys

def to_int(s: str) -> int:
    s = s.strip()
    if "." in s:
        s = s.split(".", 1)[0]
    return int(s)

funded = to_int(sys.argv[1])
final = to_int(sys.argv[2])
released = to_int(sys.argv[3])
if not (final < funded):
    raise SystemExit(f"pool did not decrease: funded={funded} final={final}")
if not (released > 0):
    raise SystemExit(f"total_released not positive: {released}")
print(f"OK: pool decreased by {funded - final}, total_released={released}")
PY

log "Manual rewards smoke test PASSED"
