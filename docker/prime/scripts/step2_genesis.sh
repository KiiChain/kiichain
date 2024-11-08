#!/usr/bin/env sh

# Input parameters
NODE_ID=${ID:-0}

echo "Preparing genesis file"

# ACCOUNT_NAME="admin"
# echo "Adding account $ACCOUNT_NAME"
# printf "12345678\n12345678\ny\n" | kiichaind keys add $ACCOUNT_NAME >/dev/null 2>&1

override_genesis() {
  cat ~/.kiichain3/config/genesis.json | jq "$1" > ~/.kiichain3/config/tmp_genesis.json && mv ~/.kiichain3/config/tmp_genesis.json ~/.kiichain3/config/genesis.json;
}

override_genesis '.app_state["crisis"]["constant_fee"]["denom"]="ukii"'
override_genesis '.app_state["mint"]["params"]["mint_denom"]="ukii"'
override_genesis '.app_state["staking"]["params"]["bond_denom"]="ukii"'
override_genesis '.app_state["oracle"]["params"]["vote_period"]="1"'
override_genesis '.app_state["slashing"]["params"]["signed_blocks_window"]="10000"'
override_genesis '.app_state["slashing"]["params"]["min_signed_per_window"]="0.050000000000000000"'
override_genesis '.app_state["staking"]["params"]["max_validators"]=25'
override_genesis '.consensus_params["block"]["max_gas"]="-1"'
override_genesis '.app_state["staking"]["params"]["unbonding_time"]="1814400s"'
override_genesis '.app_state["distribution"]["params"]["community_tax"]="0.020000000000000000"'
override_genesis '.app_state["staking"]["params"]["max_voting_power_enforcement_threshold"]="1000000"'

# Set a token release schedule for the genesis file
# start_date="$(date +"%Y-%m-%d")"
# end_date="$(date -d "+3 days" +"%Y-%m-%d")"
# override_genesis ".app_state[\"mint\"][\"params\"][\"token_release_schedule\"]=[{\"start_date\": \"$start_date\", \"end_date\": \"$end_date\", \"token_release_amount\": \"999999999999\"}]"


# We already added node0's genesis account in configure_init, remove it here since we're going to re-add it in the "add genesis accounts" step
override_genesis '.app_state["auth"]["accounts"]=[]'
override_genesis '.app_state["bank"]["balances"]=[]'
override_genesis '.app_state["genutil"]["gen_txs"]=[]'
# override_genesis '.app_state["bank"]["denom_metadata"]=[{"denom_units":[{"denom":"UATOM","exponent":6,"aliases":["UATOM"]}],"base":"uatom","display":"uatom","name":"UATOM","symbol":"UATOM"}]'

# gov parameters
override_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="ukii"'
override_genesis '.app_state["gov"]["deposit_params"]["min_expedited_deposit"][0]["denom"]="ukii"'
override_genesis '.app_state["gov"]["deposit_params"]["max_deposit_period"]="172800s"'
override_genesis '.app_state["gov"]["voting_params"]["voting_period"]="432000s"'
override_genesis '.app_state["gov"]["voting_params"]["expedited_voting_period"]="86400s"'
override_genesis '.app_state["gov"]["tally_params"]["quorum"]="0.334000000000000000"'
override_genesis '.app_state["gov"]["tally_params"]["threshold"]="0.500000000000000000"'
override_genesis '.app_state["gov"]["tally_params"]["expedited_quorum"]="0.667000000000000000"'
override_genesis '.app_state["gov"]["tally_params"]["expedited_threshold"]="0.667000000000000000"'



# add genesis accounts for each node
while read account; do
  echo "Adding: $account"
  kiichaind add-genesis-account "$account" 1100000000000ukii
done <build/generated/genesis_accounts.txt

if [ "$NODE_ID" = 0 ]
then
  # New genesis accounts with balances
  accounts="private_sale:54000000000000ukii public_sale:126000000000000ukii liquidity:180000000000000ukii community_development:180000000000000ukii team:356700000000000ukii rewards:900000000000000ukii"
  # Loop through new accounts and set them up
  for account in $accounts; do
    name="${account%%:*}"
    balance="${account##*:}"

    printf "12345678\n" | kiichaind keys add "$name" >> build/generated/mnemonic.txt
    acct=$(printf "12345678\n" | kiichaind keys show "$name" -a)
    echo "$acct" >> build/generated/genesis_accounts.txt
    kiichaind add-genesis-account "$acct" "$balance"
  done
fi

mkdir -p ~/exported_keys
cp -r build/generated/gentx/* ~/.kiichain3/config/gentx
cp -r build/generated/exported_keys ~/exported_keys

# add validators to genesis
/usr/bin/add_validator_to_gensis.sh

# collect gentxs
echo "Collecting all gentx"
kiichaind collect-gentxs >/dev/null 2>&1

cat ~/.kiichain3/config/genesis.json

cp ~/.kiichain3/config/genesis.json build/generated/genesis.json
cp ~/.kiichain3/config/genesis.json remote/genesis.json

echo "Genesis file has been created successfully"
