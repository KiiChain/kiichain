#!/bin/bash
set -e

# This rebuilds the binary with a new version

# Rebuild the binary with make
export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export BUILD_PATH=/kiichain/kiichain3/build
export PATH=$GOBIN:$PATH:/usr/local/go/bin:$BUILD_PATH
export LEDGER_ENABLED=false
/bin/bash -c "source $HOME/.bashrc"

git config --global --add safe.directory /kiichain/kiichain3 > /dev/null
make build-linux > /dev/null

echo "PASS"
exit 0
