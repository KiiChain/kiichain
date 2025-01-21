#!/bin/bash
kiidbin=$(which ~/go/bin/kiichaind | tr -d '"')
keyname=$(printf "12345678\n" | $kiidbin keys list --output json | jq ".[0].name" | tr -d '"')
chainid=$($kiidbin status | jq ".NodeInfo.network" | tr -d '"')
kiihome=$(git rev-parse --show-toplevel | tr -d '"')

echo $keyname
echo $kiidbin
echo $chainid
echo $kiihome

# Deploy all contracts
echo "Deploying kii tester contract"

cd $kiihome/loadtest/contracts
# store
echo "Storing..."

kii_tester_res=$(printf "12345678\n" | $kiidbin tx wasm store kii_tester.wasm -y --from=$keyname --chain-id=$chainid --gas=5000000 --fees=1000000ukii --broadcast-mode=block --output=json)
kii_tester_id=$(python3 parser.py code_id $kii_tester_res)

# instantiate
echo "Instantiating..."
tester_in_res=$(printf "12345678\n" | $kiidbin tx wasm instantiate $kii_tester_id '{}' -y --no-admin --from=$keyname --chain-id=$chainid --gas=5000000 --fees=1000000ukii --broadcast-mode=block  --label=dex --output=json)
tester_addr=$(python3 parser.py contract_address $tester_in_res)

# TODO fix once implemented in loadtest config
jq '.kii_tester_address = "'$tester_addr'"' $kiihome/loadtest/config.json > $kiihome/loadtest/config_temp.json && mv $kiihome/loadtest/config_temp.json $kiihome/loadtest/config.json


echo "Deployed contracts:"
echo $tester_addr

exit 0