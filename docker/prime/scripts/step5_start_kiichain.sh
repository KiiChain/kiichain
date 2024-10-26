#!/usr/bin/env sh

NODE_ID=${ID:-0}
INVARIANT_CHECK_INTERVAL=${INVARIANT_CHECK_INTERVAL:-0}

LOG_DIR="build/generated/logs"
mkdir -p $LOG_DIR

echo "Starting the kiichaind process for node $NODE_ID with invariant check interval=$INVARIANT_CHECK_INTERVAL..."

# Background process to monitor for genesis.json and upload it
(
  GENESIS_FILE="build/generated/genesis.json"
  S3_BUCKET_URL="s3://${S3_BUCKET_NAME}/genesis.json"

  echo "Waiting for $GENESIS_FILE to be created..."
  while [ ! -f "$GENESIS_FILE" ]; do
    sleep 5
  done

  echo "Uploading $GENESIS_FILE to $S3_BUCKET_URL"
  aws s3 cp "$GENESIS_FILE" "$S3_BUCKET_URL"
  echo "Upload complete."
) &

kiichaind start --chain-id kiichain3 --inv-check-period ${INVARIANT_CHECK_INTERVAL} > "$LOG_DIR/kiichaind-$NODE_ID.log" 2>&1 &

echo "Node $NODE_ID kiichaind is started now"
echo "Done" >> build/generated/launch.complete
