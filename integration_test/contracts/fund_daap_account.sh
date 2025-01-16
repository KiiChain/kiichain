#!/bin/bash

kiichaindbin=$(which ~/go/bin/kiichaind | tr -d '"')
keyname=$(printf "12345678\n" | $kiichaindbin keys list --output json | jq ".[0].name" | tr -d '"')
keyaddress=$(printf "12345678\n" | $kiichaindbin keys list --output json | jq ".[0].address" | tr -d '"')
chainid=$($kiichaindbin status | jq ".NodeInfo.network" | tr -d '"')
kiihome=$(git rev-parse --show-toplevel | tr -d '"')
DAPP_ACCOUNT="kii1zlvjwlwl967gpdha3a3cn5u0hvfurnnuu52wvj"

cd $kiihome || exit
echo "Funding dApp account..."

$kiichaindbin tx bank send $keyname $DAPP_ACCOUNT 100000000000ukii -b block --fees 2000ukii --chain-id kii -y --output json

echo "dApp account funded with 100000000000ukii"
