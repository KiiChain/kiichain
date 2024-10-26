#!/usr/bin/env sh

LOG_DIR="build/generated/logs"
mkdir -p $LOG_DIR

# Starting kiichain chain
echo "RPC Node is starting now, check logs under $LOG_DIR"

kiichaind start --chain-id kiichain3 > "$LOG_DIR/rpc-node.log" 2>&1 &
echo "Done" >> build/generated/rpc-launch.complete