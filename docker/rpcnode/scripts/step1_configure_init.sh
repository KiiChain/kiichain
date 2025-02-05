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

# cosmovisor files
mkdir -p ~/.kiichain3/cosmovisor/genesis/bin
mkdir -p ~/.kiichain3/cosmovisor/upgrades
cp $GOBIN/kiichaind ~/.kiichain3/cosmovisor/genesis/bin
if [ -f ~/.kiichain3/cosmovisor/genesis/bin/kiichaind ]; then
    echo "Cosmovisor setup completed successfully."
else
    echo "Error: Cosmovisor setup failed. Binary not found in genesis/bin."
    exit 1
fi

# Override state sync configs

# SELECT AN RPC NODE FOR SYNCING eg. 18.227.13.176:26669
STATE_SYNC_RPC="REPLACE_SYNC_RPC"

# LIST PEERS FOR SYNCING eg. "42355192eb77b71edbaa7e03f38e335849993ca0@18.227.13.176:26668,f3232ca5248aeb38af1d99542316d3c784dbf6f2@3.15.3.149:26668"....
STATE_SYNC_PEER="REPLACE_SYNC_PEERS"
curl "$STATE_SYNC_RPC"/net_info |jq -r '.peers[] | .url' |sed -e 's#mconn://##' >> build/generated/PEERS
LATEST_HEIGHT=$(curl -s $STATE_SYNC_RPC/block | jq -r .block.header.height)
SYNC_BLOCK_HEIGHT=$LATEST_HEIGHT
SYNC_BLOCK_HASH=$(curl -s "$STATE_SYNC_RPC/block?height=$SYNC_BLOCK_HEIGHT" | jq -r .block_id.hash)
sed -i.bak -e "s|^enable *=.*|enable = true|" ~/.kiichain3/config/config.toml
sed -i.bak -e "s|^rpc-servers *=.*|rpc-servers = \"$STATE_SYNC_RPC,$STATE_SYNC_RPC\"|" ~/.kiichain3/config/config.toml
sed -i.bak -e "s|^trust-height *=.*|trust-height = $SYNC_BLOCK_HEIGHT|" ~/.kiichain3/config/config.toml
sed -i.bak -e "s|^trust-hash *=.*|trust-hash = \"$SYNC_BLOCK_HASH\"|" ~/.kiichain3/config/config.toml
sed -i.bak -e "s|^persistent-peers *=.*|persistent-peers = \"$STATE_SYNC_PEER\"|" ~/.kiichain3/config/config.toml
