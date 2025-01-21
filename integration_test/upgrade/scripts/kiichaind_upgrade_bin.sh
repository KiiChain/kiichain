#!/bin/bash
set -e

# This upgrades a node by swapping binaries

NODE_ID=${ID:-0}
INVARIANT_CHECK_INTERVAL=${INVARIANT_CHECK_INTERVAL:-0}
LOG_DIR="build/generated/logs"
export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin

# kill the existing service
pkill -f "kiichaind start"
sleep 10

# Verify the process is no longer running
if pgrep -f "kiichaind start"; then
  echo "Error: kiichaind process is still running."
  exit 1
fi

# Replace the binary
cp build/kiichaind "$GOBIN"/

# start the service with a different UPGRADE_VERSION_LIST
kiichaind start --chain-id kii --inv-check-period ${INVARIANT_CHECK_INTERVAL} > "$LOG_DIR/kiichaind-$NODE_ID.log" 2>&1 &

# Sleep to catch-up consensus
sleep 30

echo "PASS"
exit 0