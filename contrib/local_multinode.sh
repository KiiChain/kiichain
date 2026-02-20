#!/usr/bin/env bash
set -euo pipefail

#########################
# CONFIG - EDIT THESE
#########################

# Cosmos-SDK binary
BINARY="kiichaind"

# Chain ID and denom
CHAIN_ID="localchain-1"
STAKE_DENOM="akii"

# Genesis account balance
GENESIS_BALANCE="1000000000000000000000000000${STAKE_DENOM}"
SELF_DELEGATION="1000000000000000000000000${STAKE_DENOM}"

# Node homes
VAL_HOME="$HOME/.localnet/validator"
PUB_HOME="$HOME/.localnet/public"

# Monikers
VAL_MONIKER="local-validator"
PUB_MONIKER="local-public"

# Keyring backend (test is easiest for local dev)
KEYRING_BACKEND="test"

# Validator custom ports (public node will use defaults: 26656, 26657, etc.)
VAL_JSON_RPC_PORT=9545
VAL_P2P_PORT=27656
VAL_RPC_PORT=27657
VAL_PROXY_APP_PORT=27658
VAL_PPROF_PORT=27659

VAL_API_PORT=2317
VAL_GRPC_PORT=29090
VAL_GRPC_WEB_PORT=29091

# Validator mnemonic (24-word seed)
VALIDATOR_MNEMONIC="gesture inject test cycle original hollow east ridge hen combine junk child bacon zero hope comfort vacuum milk pitch cage oppose unhappy lunar seat"

pkill -f "kiichaind" || true

#########################
# HELPER FUNCTIONS
#########################

log() {
  echo -e "\033[1;32m[+] $*\033[0m"
}

err() {
  echo -e "\033[1;31m[!] $*\033[0m" >&2
}

clean_home() {
  local home="$1"
  if [ -d "$home" ]; then
    log "Cleaning existing home: $home"
    rm -rf "$home"
  fi
}

patch_validator_ports() {
  local home="$1"
  local cfg="$home/config/config.toml"
  local app="$home/config/app.toml"

  log "Patching validator ports in $cfg and $app"

  # --- config.toml ---
  sed -i.bak "s#^proxy_app = \".*\"#proxy_app = \"tcp://127.0.0.1:${VAL_PROXY_APP_PORT}\"#" "$cfg"
  sed -i "s#^laddr = \"tcp://127.0.0.1:26657\"#laddr = \"tcp://127.0.0.1:${VAL_RPC_PORT}\"#" "$cfg"
  sed -i "s#^laddr = \"tcp://0.0.0.0:26656\"#laddr = \"tcp://0.0.0.0:${VAL_P2P_PORT}\"#" "$cfg"
  sed -i "s#^pprof_laddr = \".*\"#pprof_laddr = \"localhost:${VAL_PPROF_PORT}\"#" "$cfg" || true

  # --- app.toml ---
  sed -i.bak "s#^address = \"tcp://0.0.0.0:1317\"#address = \"tcp://0.0.0.0:${VAL_API_PORT}\"#" "$app"
  sed -i "s#^address = \"0.0.0.0:9090\"#address = \"0.0.0.0:${VAL_GRPC_PORT}\"#" "$app"
  sed -i "s#^address = \"0.0.0.0:9091\"#address = \"0.0.0.0:${VAL_GRPC_WEB_PORT}\"#" "$app"
}

set_persistent_peers() {
  local home="$1"
  local peer_str="$2"
  local cfg="$home/config/config.toml"

  log "Setting persistent_peers in $cfg to: $peer_str"
  sed -i.bak "s/^persistent_peers = \".*\"/persistent_peers = \"${peer_str}\"/" "$cfg"
}

#########################
# SAFETY CHECKS
#########################

if [ -z "$VALIDATOR_MNEMONIC" ]; then
  err "VALIDATOR_MNEMONIC is empty. Export your 24-word seed first:"
  err "  export VALIDATOR_MNEMONIC=\"word1 word2 ... word24\""
  exit 1
fi

if ! command -v "$BINARY" >/dev/null 2>&1; then
  err "Binary '$BINARY' not found in PATH"
  exit 1
fi

#########################
# 1. INIT VALIDATOR NODE
#########################

log "Step 1: Initializing validator node"

clean_home "$VAL_HOME"
clean_home "$PUB_HOME"

# Init validator
log "Initializing validator home at $VAL_HOME"
$BINARY init "$VAL_MONIKER" --chain-id "$CHAIN_ID" --home "$VAL_HOME"

# Create validator key from mnemonic
log "Recovering validator key (val) from VALIDATOR_MNEMONIC"
# The CLI will prompt for the mnemonic; we pipe it in non-interactively.
printf '%s\n' "$VALIDATOR_MNEMONIC" | \
  $BINARY keys add val \
    --keyring-backend "$KEYRING_BACKEND" \
    --home "$VAL_HOME" \
    --recover

VAL_ADDR=$($BINARY keys show val -a --keyring-backend "$KEYRING_BACKEND" --home "$VAL_HOME")
log "Validator address: $VAL_ADDR"

# Add genesis account (big balance for local dev)
log "Adding genesis account"
$BINARY genesis add-genesis-account "$VAL_ADDR" "${GENESIS_BALANCE}" --home "$VAL_HOME"

# Create gentx
log "Creating gentx"
$BINARY genesis gentx val "${SELF_DELEGATION}" \
  --chain-id "$CHAIN_ID" \
  --keyring-backend "$KEYRING_BACKEND" \
  --home "$VAL_HOME"

# Collect gentxs
log "Collecting gentxs"
$BINARY genesis collect-gentxs --home "$VAL_HOME"

# Patch the genesis to hold the correct bank information
log "Patching genesis with correct bank denom metadata"
GENESIS="$VAL_HOME/config/genesis.json"
TMP_GENESIS="$VAL_HOME/config/genesis_tmp.json"
jq '.app_state["evm"]["params"]["evm_denom"]="akii"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["bank"]["denom_metadata"]=[{"description":"The native staking token of the kiichain network","denomUnits":[{"denom":"akii","exponent":"0"},{"denom":"kii","exponent":"18"}],"base":"akii","display":"kii","name":"kii","symbol":"KII"}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Patch validator ports so they DON'T conflict with default ports
patch_validator_ports "$VAL_HOME"

# Get validator node ID (for persistent_peers)
VAL_NODE_ID=$($BINARY comet show-node-id --home "$VAL_HOME")
VAL_P2P_ADDR="${VAL_NODE_ID}@0.0.0.0:${VAL_P2P_PORT}"

log "Validator node-id: $VAL_NODE_ID"
log "Validator P2P address: $VAL_P2P_ADDR"

#########################
# 2. INIT PUBLIC NODE
#########################

log "Step 2: Initializing public node"

log "Initializing public home at $PUB_HOME"
$BINARY init "$PUB_MONIKER" --chain-id "$CHAIN_ID" --home "$PUB_HOME"

# Use the SAME genesis as validator
log "Copying genesis.json from validator to public node"
cp "$VAL_HOME/config/genesis.json" "$PUB_HOME/config/genesis.json"

# Point public node to validator as persistent peer
set_persistent_peers "$PUB_HOME" "$VAL_P2P_ADDR"

#########################
# 3. START BOTH NODES
#########################

log "Step 3: Starting nodes"

log "Starting validator node..."
$BINARY start --minimum-gas-prices 0akii --home "$VAL_HOME" --json-rpc.enable true --json-rpc.address "127.0.0.1:${VAL_JSON_RPC_PORT}" --log_no_color >"$VAL_HOME/validator.log" 2>&1 &
VAL_PID=$!
log "Validator PID: $VAL_PID (logs: $VAL_HOME/validator.log)"

log "Starting public node..."
$BINARY start --minimum-gas-prices 0akii --mempool.max-txs 0 --home "$PUB_HOME" --json-rpc.enable true --log_no_color >"$PUB_HOME/public.log" 2>&1 &
PUB_PID=$!
log "Public PID: $PUB_PID (logs: $PUB_HOME/public.log)"

log "==============================="
log "Local 2-node network started!"
log
log "Validator:"
echo "  Home: $VAL_HOME"
echo "  RPC : http://127.0.0.1:${VAL_RPC_PORT}"
echo "  P2P : ${VAL_P2P_ADDR}"
log
log "Public node (DEFAULT PORTS):"
echo "  Home: $PUB_HOME"
echo "  RPC : http://127.0.0.1:26657"
echo "  P2P : tcp://0.0.0.0:26656"
log "==============================="

echo
log "Tip: check validator log: tail -f $VAL_HOME/validator.log"
log "Tip: check public log   : tail -f $PUB_HOME/public.log"

# echo
# log "Killing the public node in 5 seconds..."
# sleep 5
# log "Killing public node (PID: $PUB_PID)"
# kill "$PUB_PID"
# log "Public node killed."
