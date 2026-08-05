#!/usr/bin/env bash
# E2E: prove blocking the gov module account fixes EVM gov-precompile deposits
# without soft-blocking native MsgDeposit.
#
# Runs twice against local kiichaind:
#   1) UNFIXED (gov deleted from blocked addrs)
#        native deposit OK, EVM deposit NOT applied on-chain
#   2) FIXED   (gov stays blocked)
#        native deposit OK, EVM deposit applied on-chain
#
# Success/failure for EVM is judged by proposal deposit totals (not cast receipts).
# The unfixed bug can return eth receipt status=0x1 while the cosmos tx never lands.
#
# Usage (from kiichain repo root):
#   ./contrib/scripts/e2e_gov_module_block_deposit.sh
#
# Env:
#   SKIP_BUILD=1           reuse BIN_FIXED / BIN_UNFIXED
#   SKIP_UNFIXED=1         only run the fixed scenario
#   SKIP_FIXED=1           only run the unfixed scenario
#   KEEP_HOME=1            leave node homes under WORKDIR
#   WORKDIR=/tmp/...       override scratch dir

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

CHAIN_ID="localgov_1010-1"
EVM_CHAIN_ID=1010
DENOM="akii"
KEY="val"
MONIKER="govblock"
GOV_PRECOMPILE="0x0000000000000000000000000000000000000805"

MIN_DEPOSIT="2000000000000000000${DENOM}"       # 2 KII
SUBMIT_DEPOSIT="1000000000000000000${DENOM}"    # 1 KII (stays in deposit period)
NATIVE_DEPOSIT="100000000000000000${DENOM}"     # 0.1 KII
EVM_DEPOSIT_WEI="100000000000000000"            # 0.1 KII

# High ports to avoid colliding with a local evmd/kiichaind on 26657/8545.
RPC_PORT="${RPC_PORT:-36657}"
GRPC_PORT="${GRPC_PORT:-39090}"
JSONRPC_PORT="${JSONRPC_PORT:-38545}"
P2P_PORT="${P2P_PORT:-36656}"
API_PORT="${API_PORT:-31317}"

WORKDIR="${WORKDIR:-$(mktemp -d /tmp/kiichain-gov-block-e2e.XXXXXX)}"
BIN_DIR="${WORKDIR}/bin"
BIN_FIXED="${BIN_FIXED:-${BIN_DIR}/kiichaind-fixed}"
BIN_UNFIXED="${BIN_UNFIXED:-${BIN_DIR}/kiichaind-unfixed}"
APP_GO="app/app.go"
NODE_PID=""

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

need jq
need cast
need go
need sed
need python3

log() { printf '\n==> %s\n' "$*"; }
die() { printf 'ERROR: %s\n' "$*" >&2; printf 'ERROR: %s\n' "$*"; exit 1; }


cleanup() {
  if [[ -n "${NODE_PID}" ]] && kill -0 "${NODE_PID}" 2>/dev/null; then
    kill "${NODE_PID}" 2>/dev/null || true
    wait "${NODE_PID}" 2>/dev/null || true
  fi
  pkill -f "kiichaind start --home ${WORKDIR}" 2>/dev/null || true
  if [[ "${KEEP_HOME:-0}" != "1" ]]; then
    rm -rf "${WORKDIR}"
  else
    echo "KEEP_HOME=1; left ${WORKDIR}"
  fi
}
trap cleanup EXIT

gov_is_unblocked_in_tree() {
  grep -q 'delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName)' "${APP_GO}"
}

ensure_govtypes_import() {
  if grep -q 'govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"' "${APP_GO}"; then
    return
  fi
  # Insert import next to govkeeper if present, else after authtypes.
  if grep -q 'govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"' "${APP_GO}"; then
    sed -i.bak '/govkeeper "github.com\/cosmos\/cosmos-sdk\/x\/gov\/keeper"/a\
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
' "${APP_GO}"
  else
    sed -i.bak '/authtypes "github.com\/cosmos\/cosmos-sdk\/x\/auth\/types"/a\
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
' "${APP_GO}"
  fi
  rm -f "${APP_GO}.bak"
}

write_blocked_fn() {
  local mode="$1" # fixed|unfixed
  local tmp
  tmp="$(mktemp)"
  python3 - "$APP_GO" "$mode" "$tmp" <<'PY'
import re, sys
path, mode, out = sys.argv[1:4]
src = open(path).read()
pat = re.compile(
    r"func \(app \*KiichainApp\) BlockedModuleAccountAddrs\(modAccAddrs map\[string\]bool\) map\[string\]bool \{.*?\n\}",
    re.S,
)
if mode == "unfixed":
    body = """func (app *KiichainApp) BlockedModuleAccountAddrs(modAccAddrs map[string]bool) map[string]bool {
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	return modAccAddrs
}"""
else:
    body = """func (app *KiichainApp) BlockedModuleAccountAddrs(modAccAddrs map[string]bool) map[string]bool {
	return modAccAddrs
}"""
new, n = pat.subn(body, src, count=1)
if n != 1:
    raise SystemExit("failed to rewrite BlockedModuleAccountAddrs")
open(out, "w").write(new)
PY
  mv "$tmp" "$APP_GO"
}

build_variant() {
  local mode="$1"
  local out="$2"
  local restore_copy
  restore_copy="$(mktemp)"
  cp "${APP_GO}" "${restore_copy}"

  log "Building ${mode} binary -> ${out}"
  if [[ "$mode" == "unfixed" ]]; then
    ensure_govtypes_import
    write_blocked_fn unfixed
  else
    write_blocked_fn fixed
    # Drop unused govtypes import if present and unused.
    if grep -q 'govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"' "${APP_GO}" \
      && ! grep -q 'govtypes\.' "${APP_GO}"; then
      sed -i.bak '/govtypes "github.com\/cosmos\/cosmos-sdk\/x\/gov\/types"/d' "${APP_GO}"
      rm -f "${APP_GO}.bak"
    fi
  fi

  # Force a real rebuild of app package (avoid stale object reuse).
  rm -f "${ROOT}/build/kiichaind"
  touch "${APP_GO}"
  make build
  mkdir -p "$(dirname "$out")"
  cp "${ROOT}/build/kiichaind" "$out"
  chmod +x "$out"
  shasum -a 256 "$out" | awk '{print $1}' >"${out}.sha256"
  log "${mode} sha256=$(cat "${out}.sha256")"

  mv "${restore_copy}" "${APP_GO}"
  git checkout -- "${APP_GO}" 2>/dev/null || true
}

wait_for_rpc() {
  local home="$1"
  local tries=90
  local i
  for ((i = 1; i <= tries; i++)); do
    if ! kill -0 "${NODE_PID}" 2>/dev/null; then
      echo "---- node log (tail) ----" >&2
      tail -n 120 "${home}/node.log" >&2 || true
      die "node process exited early"
    fi
    local status_json=""
    status_json="$("${BINARY}" status --home "${home}" --node "tcp://127.0.0.1:${RPC_PORT}" 2>/dev/null || true)"
    if [[ -n "${status_json}" ]]; then
      local height chain
      height="$(echo "${status_json}" | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height // empty')"
      chain="$(echo "${status_json}" | jq -r '.node_info.network // .NodeInfo.network // empty')"
      if [[ "${chain}" == "${CHAIN_ID}" && -n "${height}" && "${height}" != "0" && "${height}" -lt 1000 ]]; then
        echo "node ready at height ${height} chain_id=${chain}"
        return 0
      fi
    fi
    sleep 1
  done
  echo "---- node log (tail) ----" >&2
  tail -n 120 "${home}/node.log" >&2 || true
  die "node did not become ready on port ${RPC_PORT}"
}

wait_for_jsonrpc() {
  local tries=40
  local i
  for ((i = 1; i <= tries; i++)); do
    if cast block-number --rpc-url "http://127.0.0.1:${JSONRPC_PORT}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  die "json-rpc did not become ready on :${JSONRPC_PORT}"
}

stop_node() {
  if [[ -n "${NODE_PID}" ]] && kill -0 "${NODE_PID}" 2>/dev/null; then
    kill "${NODE_PID}" 2>/dev/null || true
    # Wait for clean exit; escalate if needed.
    local i
    for ((i = 1; i <= 20; i++)); do
      if ! kill -0 "${NODE_PID}" 2>/dev/null; then
        break
      fi
      sleep 0.5
    done
    if kill -0 "${NODE_PID}" 2>/dev/null; then
      kill -9 "${NODE_PID}" 2>/dev/null || true
      wait "${NODE_PID}" 2>/dev/null || true
    else
      wait "${NODE_PID}" 2>/dev/null || true
    fi
  fi
  NODE_PID=""
  # Also reap any stray process bound to our home path / ports.
  pkill -f "kiichaind start --home ${WORKDIR}" 2>/dev/null || true
  local port
  for port in "${RPC_PORT}" "${JSONRPC_PORT}" "${GRPC_PORT}" "${P2P_PORT}"; do
    local pids
    pids="$(lsof -tiTCP:"${port}" -sTCP:LISTEN 2>/dev/null || true)"
    if [[ -n "${pids}" ]]; then
      # shellcheck disable=SC2086
      kill ${pids} 2>/dev/null || true
      sleep 0.5
      pids="$(lsof -tiTCP:"${port}" -sTCP:LISTEN 2>/dev/null || true)"
      if [[ -n "${pids}" ]]; then
        # shellcheck disable=SC2086
        kill -9 ${pids} 2>/dev/null || true
      fi
    fi
  done
  sleep 1
}


init_home() {
  local home="$1"
  rm -rf "${home}"
  mkdir -p "${home}"

  "${BINARY}" init "${MONIKER}" --chain-id "${CHAIN_ID}" --home "${home}" >/dev/null 2>&1
  "${BINARY}" config set client chain-id "${CHAIN_ID}" --home "${home}" >/dev/null
  "${BINARY}" config set client keyring-backend test --home "${home}" >/dev/null
  "${BINARY}" config set client node "tcp://127.0.0.1:${RPC_PORT}" --home "${home}" >/dev/null
  "${BINARY}" keys add "${KEY}" --home "${home}" --keyring-backend test >/dev/null 2>&1

  "${BINARY}" genesis add-genesis-account "${KEY}" \
    "1000000000000000000000000${DENOM}" \
    --home "${home}" --keyring-backend test >/dev/null

  "${BINARY}" genesis gentx "${KEY}" "1000000000000000000000${DENOM}" \
    --home "${home}" --chain-id "${CHAIN_ID}" --keyring-backend test >/dev/null 2>&1

  "${BINARY}" genesis collect-gentxs --home "${home}" >/dev/null 2>&1

  local genesis="${home}/config/genesis.json"
  local tmp
  tmp="$(mktemp)"
  jq --arg denom "$DENOM" '
    .app_state.staking.params.bond_denom = $denom
    | .app_state.crisis.constant_fee.denom = $denom
    | .app_state.gov.params.min_deposit = [{"denom":$denom,"amount":"2000000000000000000"}]
    | .app_state.gov.params.expedited_min_deposit = [{"denom":$denom,"amount":"2000000000000000000"}]
    | .app_state.gov.params.min_initial_deposit_ratio = "0.000000000000000000"
    | .app_state.gov.params.min_deposit_ratio = "0.000000000000000000"
    | .app_state.gov.params.max_deposit_period = "600s"
    | .app_state.gov.params.voting_period = "600s"
    | .app_state.gov.params.expedited_voting_period = "300s"
    | .app_state.evm.params.evm_denom = $denom
    | .app_state.evm.params.active_static_precompiles = [
        "0x0000000000000000000000000000000000000100",
        "0x0000000000000000000000000000000000000400",
        "0x0000000000000000000000000000000000000800",
        "0x0000000000000000000000000000000000000801",
        "0x0000000000000000000000000000000000000802",
        "0x0000000000000000000000000000000000000803",
        "0x0000000000000000000000000000000000000804",
        "0x0000000000000000000000000000000000000805",
        "0x0000000000000000000000000000000000000806",
        "0x0000000000000000000000000000000000001003"
      ]
    | .app_state.feemarket.params.no_base_fee = true
    | .app_state.feemarket.params.base_fee = "0"
    | .app_state.bank.denom_metadata = [{
        "description": "The native staking token of the kiichain network",
        "denom_units": [
          {"denom": "akii", "exponent": 0, "aliases": []},
          {"denom": "kii", "exponent": 18, "aliases": []}
        ],
        "base": "akii",
        "display": "kii",
        "name": "kii",
        "symbol": "KII",
        "uri": "",
        "uri_hash": ""
      }]
  ' "${genesis}" >"${tmp}"
  mv "${tmp}" "${genesis}"

  # Prefer sed edits that work on both Linux and macOS.
  local app_toml="${home}/config/app.toml"
  local client_toml="${home}/config/client.toml"
  local config_toml="${home}/config/config.toml"

  python3 - "${app_toml}" "${client_toml}" "${config_toml}" "${RPC_PORT}" "${P2P_PORT}" <<'PY'
from pathlib import Path
import sys
app_toml, client_toml, config_toml, rpc_port, p2p_port = sys.argv[1:6]

def set_key(text, key, value):
    import re
    pat = re.compile(rf'(?m)^(\s*{re.escape(key)}\s*=\s*).*$')
    if pat.search(text):
        return pat.sub(rf'\1{value}', text, count=1)
    return text + f"\n{key} = {value}\n"

app = Path(app_toml).read_text()
app = set_key(app, "minimum-gas-prices", '"0akii"')
import re
app = re.sub(r'(?ms)(\[api\]\n(?:.*?\n)*?)(enable\s*=\s*)true', r'\1\2false', app, count=1)
app = re.sub(r'(?ms)(\[grpc-web\]\n(?:.*?\n)*?)(enable\s*=\s*)true', r'\1\2false', app, count=1)
Path(app_toml).write_text(app)

client = Path(client_toml).read_text()
client = set_key(client, "node", f'"tcp://127.0.0.1:{rpc_port}"')
Path(client_toml).write_text(client)

cfg = Path(config_toml).read_text()
# RPC and P2P listen addresses appear as laddr under [rpc]/[p2p]; rewrite both occurrences carefully.
cfg = cfg.replace('laddr = "tcp://127.0.0.1:26657"', f'laddr = "tcp://127.0.0.1:{rpc_port}"')
cfg = cfg.replace('laddr = "tcp://0.0.0.0:26656"', f'laddr = "tcp://127.0.0.1:{p2p_port}"')
Path(config_toml).write_text(cfg)
PY
}

start_node() {
  local home="$1"
  stop_node
  log "Starting node (${home})"
  "${BINARY}" start \
    --home "${home}" \
    --minimum-gas-prices "0${DENOM}" \
    --json-rpc.enable true \
    --json-rpc.address "127.0.0.1:${JSONRPC_PORT}" \
    --json-rpc.ws-address "127.0.0.1:$((JSONRPC_PORT + 1))" \
    --json-rpc.api eth,txpool,personal,net,debug,web3 \
    --grpc.enable true \
    --grpc.address "127.0.0.1:${GRPC_PORT}" \
    --rpc.laddr "tcp://127.0.0.1:${RPC_PORT}" \
    --p2p.laddr "tcp://127.0.0.1:${P2P_PORT}" \
    --rpc.pprof_laddr "localhost:$((RPC_PORT + 1000))" \
    >"${home}/node.log" 2>&1 &
  NODE_PID=$!
  wait_for_rpc "${home}"
  wait_for_jsonrpc
}

export_eth_key() {
  local home="$1"
  "${BINARY}" keys export "${KEY}" --home "${home}" --keyring-backend test --unarmored-hex --unsafe -y 2>/dev/null \
    | tr -d '\n\r ' | sed 's/^0x//'
}

bech32_to_hex() {
  # Convert 20-byte account address from bech32 via kiichaind debug if available,
  # else derive from cast wallet.
  local addr="$1"
  if "${BINARY}" debug addr "${addr}" >/dev/null 2>&1; then
    "${BINARY}" debug addr "${addr}" 2>/dev/null | awk '/Address \(hex\):/ {print $3; exit}'
    return
  fi
  # Fallback: use eth address from private key (preferred for EVM).
  echo ""
}

submit_text_proposal() {
  local home="$1"
  local out
  out="$("${BINARY}" tx gov submit-legacy-proposal \
    --title "gov-block-e2e" \
    --description "e2e deposit period proposal" \
    --deposit "${SUBMIT_DEPOSIT}" \
    --type Text \
    --from "${KEY}" \
    --home "${home}" \
    --keyring-backend test \
    --chain-id "${CHAIN_ID}" \
    --node "tcp://127.0.0.1:${RPC_PORT}" \
    --fees "2000000000000000${DENOM}" \
    --gas auto --gas-adjustment 1.5 \
    --yes \
    --output json)"
  local code
  code="$(echo "${out}" | jq -r '.code // 0')"
  [[ "${code}" == "0" ]] || {
    echo "${out}" >&2
    die "submit-proposal failed"
  }
  # Wait for the tx to land, then resolve proposal id.
  local i id=""
  for ((i = 1; i <= 20; i++)); do
    id="$("${BINARY}" query gov proposals --home "${home}" --node "tcp://127.0.0.1:${RPC_PORT}" --output json 2>/dev/null \
      | jq -r '
          (.proposals // [])
          | map(.id // .proposal_id)
          | map(select(. != null and . != ""))
          | last // empty
        ')"
    if [[ -n "${id}" && "${id}" != "null" ]]; then
      break
    fi
    sleep 1
  done
  [[ -n "${id}" && "${id}" != "null" ]] || {
    echo "${out}" >&2
    "${BINARY}" query gov proposals --home "${home}" --node "tcp://127.0.0.1:${RPC_PORT}" --output json >&2 || true
    die "could not determine proposal id"
  }
  echo "${id}"
}

native_deposit() {
  local home="$1"
  local proposal_id="$2"
  "${BINARY}" tx gov deposit "${proposal_id}" "${NATIVE_DEPOSIT}" \
    --from "${KEY}" \
    --home "${home}" \
    --keyring-backend test \
    --chain-id "${CHAIN_ID}" \
    --node "tcp://127.0.0.1:${RPC_PORT}" \
    --fees "2000000000000000${DENOM}" \
    --gas auto --gas-adjustment 1.5 \
    --yes \
    --output json | jq -e '.code == 0' >/dev/null
}

proposal_deposit_total() {
  local home="$1"
  local proposal_id="$2"
  "${BINARY}" query gov deposits "${proposal_id}" \
    --home "${home}" \
    --node "tcp://127.0.0.1:${RPC_PORT}" \
    --output json 2>/dev/null \
    | jq -r --arg d "${DENOM}" '
        [.deposits // [] | .[] | .amount // [] | .[] | select(.denom == $d) | .amount | tonumber]
        | add // 0
      '
}

# Fire EVM gov deposit. Exit 0 only if on-chain proposal deposits increased by
# EVM_DEPOSIT_WEI. Cast receipts alone are insufficient: the unfixed bug can
# return eth receipt status=0x1 while the cosmos tx never lands / deposit is skipped.
evm_deposit_applied() {
  # Avoid `set -e` + `return 1` (bash 3.2 exits the whole script).
  set +e
  local home="$1"
  local pk="$2"
  local eth_addr="$3"
  local proposal_id="$4"
  local before after
  before="$(proposal_deposit_total "${home}" "${proposal_id}")"

  local rpc="http://127.0.0.1:${JSONRPC_PORT}"
  local nonce
  nonce="$(cast nonce "${eth_addr}" --rpc-url "${rpc}")"

  local out cast_rc
  out="$(cast send "${GOV_PRECOMPILE}" \
    "deposit(address,uint64,(string,uint256)[])" \
    "${eth_addr}" \
    "${proposal_id}" \
    "[(${DENOM},${EVM_DEPOSIT_WEI})]" \
    --rpc-url "${rpc}" \
    --private-key "${pk}" \
    --chain "${EVM_CHAIN_ID}" \
    --nonce "${nonce}" \
    --gas-limit 500000 \
    --gas-price 1000000000 \
    --json 2>"${WORKDIR}/cast.err")"
  cast_rc=$?
  printf '%s\n' "${out}" >"${WORKDIR}/cast.out"
  cat "${WORKDIR}/cast.err" >>"${WORKDIR}/cast.out" 2>/dev/null || true

  # Wait a couple blocks for indexing / commit side-effects.
  sleep 3
  after="$(proposal_deposit_total "${home}" "${proposal_id}")"
  log "deposit totals before=${before} after=${after} cast_rc=${cast_rc}"

  python3 - "${before}" "${after}" "${EVM_DEPOSIT_WEI}" <<'PY'
import sys
before, after, expect = map(int, sys.argv[1:4])
raise SystemExit(0 if after - before == expect else 1)
PY
}

run_scenario() {
  local label="$1"
  local binary="$2"
  local expect_evm="$3" # success|fail

  BINARY="${binary}"
  local home="${WORKDIR}/home-${label}"

  # Isolate ports per scenario so a slow shutdown cannot leak traffic
  # from the previous binary onto the next one (same chain-id).
  if [[ "${label}" == "unfixed" ]]; then
    RPC_PORT=36657
    GRPC_PORT=39090
    JSONRPC_PORT=38545
    P2P_PORT=36656
  else
    RPC_PORT=36667
    GRPC_PORT=39100
    JSONRPC_PORT=38555
    P2P_PORT=36666
  fi

  log "=== SCENARIO: ${label} (expect EVM deposit ${expect_evm}) ports rpc=${RPC_PORT} eth=${JSONRPC_PORT} ==="
  stop_node
  init_home "${home}"
  start_node "${home}"

  log "binary=$(basename "${BINARY}") sha=$(shasum -a 256 "${BINARY}" | awk '{print $1}')"

  local pk eth_addr proposal_id
  pk="$(export_eth_key "${home}")"
  [[ -n "${pk}" ]] || die "failed to export eth key"
  eth_addr="$(cast wallet address --private-key "${pk}")"
  log "depositor eth=${eth_addr}"

  proposal_id="$(submit_text_proposal "${home}")"
  log "proposal_id=${proposal_id}"

  log "native MsgDeposit (must succeed — no soft-block)"
  local before_native after_native
  before_native="$(proposal_deposit_total "${home}" "${proposal_id}")"
  native_deposit "${home}" "${proposal_id}" \
    || die "[${label}] native deposit failed — soft-block regression"
  sleep 2
  after_native="$(proposal_deposit_total "${home}" "${proposal_id}")"
  if ! python3 - "${before_native}" "${after_native}" "${NATIVE_DEPOSIT%${DENOM}}" <<'PY'
import sys
before, after, expect = map(int, sys.argv[1:4])
raise SystemExit(0 if after - before == expect else 1)
PY
  then
    die "[${label}] native deposit did not increase proposal deposits (${before_native} -> ${after_native})"
  fi
  log "[${label}] native deposit applied (${before_native} -> ${after_native})"

  log "EVM gov precompile deposit"
  # Cosmos and EVM share the account nonce for eth_secp256k1 keys; wait for
  # prior cosmos txs to commit before reading the next nonce.
  sleep 3
  local evm_rc=0
  set +e
  evm_deposit_applied "${home}" "${pk}" "${eth_addr}" "${proposal_id}"
  evm_rc=$?
  set -e
  log "[${label}] evm_deposit_applied rc=${evm_rc}"

  if [[ "${expect_evm}" == "success" ]]; then
    [[ ${evm_rc} -eq 0 ]] || die "[${label}] expected EVM deposit to increase proposal deposits"
    log "[${label}] EVM deposit applied on-chain (fix OK)"
  else
    [[ ${evm_rc} -ne 0 ]] || die "[${label}] expected EVM deposit NOT applied (unfixed bug), but deposits increased"
    log "[${label}] EVM deposit not applied as expected (bug reproduced; cast receipt alone is unreliable)"
  fi

  stop_node
}


mkdir -p "${BIN_DIR}" "${WORKDIR}"
log "WORKDIR=${WORKDIR}"

if [[ "${SKIP_BUILD:-0}" != "1" ]]; then
  if [[ "${SKIP_FIXED:-0}" != "1" ]]; then
    build_variant fixed "${BIN_FIXED}"
  fi
  if [[ "${SKIP_UNFIXED:-0}" != "1" ]]; then
    build_variant unfixed "${BIN_UNFIXED}"
  fi
  if [[ "${SKIP_FIXED:-0}" != "1" && "${SKIP_UNFIXED:-0}" != "1" ]]; then
    [[ "$(cat "${BIN_FIXED}.sha256")" != "$(cat "${BIN_UNFIXED}.sha256")" ]] \
      || die "fixed and unfixed binaries are identical — rebuild/patch failed"
  fi
else
  [[ -x "${BIN_FIXED}" || "${SKIP_FIXED:-0}" == "1" ]] || die "SKIP_BUILD=1 but missing ${BIN_FIXED}"
  [[ -x "${BIN_UNFIXED}" || "${SKIP_UNFIXED:-0}" == "1" ]] || die "SKIP_BUILD=1 but missing ${BIN_UNFIXED}"
fi

if [[ "${SKIP_UNFIXED:-0}" != "1" ]]; then
  run_scenario "unfixed" "${BIN_UNFIXED}" "fail"
fi
if [[ "${SKIP_FIXED:-0}" != "1" ]]; then
  run_scenario "fixed" "${BIN_FIXED}" "success"
fi

log "ALL CHECKS PASSED"
echo "Summary:"
echo "  - native MsgDeposit works with gov blocked (no soft-block)"
echo "  - EVM gov deposit does not apply on-chain when gov is unblocked (bug)"
echo "  - EVM gov deposit applies on-chain when gov is blocked (fix)"
