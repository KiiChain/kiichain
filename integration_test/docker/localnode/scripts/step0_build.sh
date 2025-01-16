#!/usr/bin/env sh

# Input parameters
NODE_ID=${ID:-0}
ARCH=$(uname -m)

# Build kiichaind
echo "Building kiichaind from local branch"
git config --global --add safe.directory /kiichain/kiichain3
export LEDGER_ENABLED=false
make clean
make build-linux
# make build-price-feeder-linux # TODO: Re-enable me after we have the oracle module again
mkdir -p build/generated
echo "DONE" > build/generated/build.complete
