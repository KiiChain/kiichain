#!/usr/bin/env sh

# Set up GO PATH
echo "Configure and initialize environment"

cp build/kiichaind "$GOBIN"/

# Testing whether kiichaind works or not
kiichaind version # Uncomment the below line if there are any dependency issues
# ldd build/kiichaind

# Initialize validator node
MONIKER="kiichain-rpc-node"
kiichaind init --chain-id kiichain3 "$MONIKER"

# Copy configs
cp docker/rpcnode/config/app.toml ~/.kiichain3/config/app.toml
cp docker/rpcnode/config/config.toml ~/.kiichain3/config/config.toml
cp remote/genesis.json ~/.kiichain3/config/genesis.json

# Override state sync configs

# SELECT AN RPC NODE FOR SYNCING eg. 192.168.10.10:26657
STATE_SYNC_RPC="18.223.214.46:26657"
# LIST PEERS FOR SYNCING eg. 2f9846450b7a3dcf4af1ac0082e3279c16744df8@172.31.9.18:26656,ec98c4a28a2023f4f976828c8a8e7127bfef4e1b@172.31.4.96:26656....
STATE_SYNC_PEER="9adcd5aa5d2e85c23347c4bab2824860757ff977@172.31.28.248:26656,15e8eb069b8b0c3ff3b1b66f46b72d8a33217e44@172.31.28.248:26662,a0e69bcd4461649adaa546b5b91da0f5c4ba99f1@172.31.28.248:26659"
curl "$STATE_SYNC_RPC"/net_info |jq -r '.peers[] | .url' |sed -e 's#mconn://##' >> build/generated/PEERS
LATEST_HEIGHT=$(curl -s $STATE_SYNC_RPC/block | jq -r .block.header.height)
SYNC_BLOCK_HEIGHT=$LATEST_HEIGHT
SYNC_BLOCK_HASH=$(curl -s "$STATE_SYNC_RPC/block?height=$SYNC_BLOCK_HEIGHT" | jq -r .block_id.hash)
sed -i.bak -e "s|^enable *=.*|enable = true|" ~/.kiichain3/config/config.toml
sed -i.bak -e "s|^rpc-servers *=.*|rpc-servers = \"$STATE_SYNC_RPC,$STATE_SYNC_RPC\"|" ~/.kiichain3/config/config.toml
sed -i.bak -e "s|^trust-height *=.*|trust-height = $SYNC_BLOCK_HEIGHT|" ~/.kiichain3/config/config.toml
sed -i.bak -e "s|^trust-hash *=.*|trust-hash = \"$SYNC_BLOCK_HASH\"|" ~/.kiichain3/config/config.toml
sed -i.bak -e "s|^persistent-peers *=.*|persistent-peers = \"$STATE_SYNC_PEER\"|" ~/.kiichain3/config/config.toml
