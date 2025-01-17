#!/bin/bash -e

echo "Funding dApp account..."
ADMIN="admin"
DAPP_ADDRESS="kii1zlvjwlwl967gpdha3a3cn5u0hvfurnnuu52wvj"

printf "12345678\n" | kiichaind tx bank send $ADMIN $DAPP_ADDRESS 100000000000ukii -b block --fees 2000ukii --chain-id kii -y --output json

echo "dApp account funded with 100000000000ukii"
