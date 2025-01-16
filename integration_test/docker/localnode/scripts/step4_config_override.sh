#!/usr/bin/env sh

NODE_ID=${ID:-0}

APP_CONFIG_FILE="build/generated/node_$NODE_ID/app.toml"
TENDERMINT_CONFIG_FILE="build/generated/node_$NODE_ID/config.toml"
cp build/generated/genesis.json ~/.kiichain3/config/genesis.json
cp "$APP_CONFIG_FILE" ~/.kiichain3/config/app.toml
cp "$TENDERMINT_CONFIG_FILE" ~/.kiichain3/config/config.toml

# Override up persistent peers
NODE_IP=$(hostname -i | awk '{print $1}')
PEERS=$(cat build/generated/persistent_peers.txt |grep -v "$NODE_IP" | paste -sd "," -)
sed -i'' -e 's/persistent-peers = ""/persistent-peers = "'$PEERS'"/g' ~/.kiichain3/config/config.toml

# Override snapshot directory
sed -i.bak -e "s|^snapshot-directory *=.*|snapshot-directory = \"./build/generated/node_$NODE_ID/snapshots\"|" ~/.kiichain3/config/app.toml
