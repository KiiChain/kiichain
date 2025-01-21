#!/bin/bash

mkdir $HOME/kii-snapshot
mkdir $HOME/key_backup
# Move priv_validator_state out so it isn't used by anyone else
mv $HOME/.kiichain3/data/priv_validator_state.json $HOME/key_backup
# Create backups
cd $HOME/kii-snapshot
tar -czf data.tar.gz -C $HOME/.kiichain3 data/
tar -czf wasm.tar.gz -C $HOME/.kiichain3 wasm/
echo "Data and Wasm snapshots created in $HOME/kii-snapshot"