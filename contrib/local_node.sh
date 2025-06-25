#!/bin/bash

CHAINID="${CHAIN_ID:-localchain_1010-1}"
MONIKER="localtestnet"
# Remember to change to other types of keyring like 'file' in-case exposing to outside world,
# otherwise your balance will be wiped quickly
# The keyring test does not require private key to steal tokens from you
KEYRING="test"
KEYALGO="eth_secp256k1"

LOGLEVEL="info"
# Set dedicated home directory for the kiichaind instance
HOMEDIR="$HOME/.kiichain"

BASEFEE=10000000
KIICHAIND=kiichaind

# Path variables
CONFIG=$HOMEDIR/config/config.toml
APP_TOML=$HOMEDIR/config/app.toml
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

# used to exit on first error (any non-zero exit code)
set -e

# Parse input flags
install=true
overwrite=""
BUILD_FOR_DEBUG=false

while [[ $# -gt 0 ]]; do
	key="$1"
	case $key in
	-y)
		echo "Flag -y passed -> Overwriting the previous chain data."
		overwrite="y"
		shift # Move past the flag
		;;
	-n)
		echo "Flag -n passed -> Not overwriting the previous chain data."
		overwrite="n"
		shift # Move past the argument
		;;
	--no-install)
		echo "Flag --no-install passed -> Skipping installation of the kiichaind binary."
		install=false
		shift # Move past the flag
		;;
	--remote-debugging)
		echo "Flag --remote-debugging passed -> Building with remote debugging options."
		BUILD_FOR_DEBUG=true
		shift # Move past the flag
		;;
	*)
		echo "Unknown flag passed: $key -> Exiting script!"
		exit 1
		;;
	esac
done

if [[ $install == true ]]; then
	if [[ $BUILD_FOR_DEBUG == true ]]; then
		# for remote debugging the optimization should be disabled and the debug info should not be stripped
		make install COSMOS_BUILD_OPTIONS=nooptimization,nostrip
	else
		make install
	fi
fi

# User prompt if neither -y nor -n was passed as a flag
# and an existing local node configuration is found.
if [[ $overwrite = "" ]]; then
	if [ -d "$HOMEDIR" ]; then
		printf "\nAn existing folder at '%s' was found. You can choose to delete this folder and start a new local node with new keys from genesis. When declined, the existing local node is started. \n" "$HOMEDIR"
		echo "Overwrite the existing configuration and start a new local node? [y/n]"
		read -r overwrite
	else
		overwrite="y"
	fi
fi

# Setup local node if overwrite is set to Yes, otherwise skip setup
if [[ $overwrite == "y" || $overwrite == "Y" ]]; then
	# Remove the previous folder
	rm -rf "$HOMEDIR"

	# Set client config
	$KIICHAIND config set client chain-id "$CHAINID" --home "$HOMEDIR"
	$KIICHAIND config set client keyring-backend "$KEYRING" --home "$HOMEDIR"

	# myKey address 0x7cb61d4117ae31a12e393a1cfa3bac666481d02e | os10jmp6sgh4cc6zt3e8gw05wavvejgr5pwjnpcky
	VAL_KEY="mykey"
	VAL_MNEMONIC="gesture inject test cycle original hollow east ridge hen combine junk child bacon zero hope comfort vacuum milk pitch cage oppose unhappy lunar seat"

	# dev0 address 0xc6fe5d33615a1c52c08018c47e8bc53646a0e101 | os1cml96vmptgw99syqrrz8az79xer2pcgp84pdun
	USER1_KEY="dev0"
	USER1_MNEMONIC="copper push brief egg scan entry inform record adjust fossil boss egg comic alien upon aspect dry avoid interest fury window hint race symptom"

	# Import keys from mnemonics
	echo "$VAL_MNEMONIC" | $KIICHAIND keys add "$VAL_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$HOMEDIR"
	echo "$USER1_MNEMONIC" | $KIICHAIND keys add "$USER1_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$HOMEDIR"

	# Set moniker and chain-id for the example chain (Moniker can be anything, chain-id must be an integer)
	$KIICHAIND init $MONIKER -o --chain-id "$CHAINID" --home "$HOMEDIR"

	# Change parameter token denominations to desired value
	jq '.app_state["staking"]["params"]["bond_denom"]="akii"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="akii"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="akii"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state["evm"]["params"]["evm_denom"]="akii"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state["mint"]["params"]["mint_denom"]="akii"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state["tokenfactory"]["params"]["denom_creation_fee"][0]["denom"]="akii"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Enable precompiles in EVM params
	jq '.app_state["evm"]["params"]["active_static_precompiles"]=["0x0000000000000000000000000000000000000100","0x0000000000000000000000000000000000000400","0x0000000000000000000000000000000000000800","0x0000000000000000000000000000000000000801","0x0000000000000000000000000000000000000802","0x0000000000000000000000000000000000000803","0x0000000000000000000000000000000000000804","0x0000000000000000000000000000000000000805","0x0000000000000000000000000000000000001001", "0x0000000000000000000000000000000000001002","0x0000000000000000000000000000000000001003"]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Enable native denomination as a token pair for STRv2
	jq '.app_state.erc20.params.native_precompiles=["0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state.erc20.token_pairs=[{contract_owner:1,erc20_address:"0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE",denom:"akii",enabled:true}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Set gas limit in genesis
	jq '.consensus_params["block"]["max_gas"]="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# enable prometheus metrics and all APIs for dev node
	if [[ "$OSTYPE" == "darwin"* ]]; then
		sed -i '' 's/prometheus = false/prometheus = true/' "$CONFIG"
		sed -i '' 's/prometheus-retention-time = 0/prometheus-retention-time  = 1000000000000/g' "$APP_TOML"
		sed -i '' 's/enabled = false/enabled = true/g' "$APP_TOML"
		sed -i '' 's/enable = false/enable = true/g' "$APP_TOML"
	else
		sed -i 's/prometheus = false/prometheus = true/' "$CONFIG"
		sed -i 's/prometheus-retention-time  = "0"/prometheus-retention-time  = "1000000000000"/g' "$APP_TOML"
		sed -i 's/enabled = false/enabled = true/g' "$APP_TOML"
		sed -i 's/enable = false/enable = true/g' "$APP_TOML"
	fi

	# Change proposal periods to pass within a reasonable time for local testing
	sed -i.bak 's/"max_deposit_period": "172800s"/"max_deposit_period": "30s"/g' "$GENESIS"
	sed -i.bak 's/"voting_period": "172800s"/"voting_period": "30s"/g' "$GENESIS"
	sed -i.bak 's/"expedited_voting_period": "86400s"/"expedited_voting_period": "15s"/g' "$GENESIS"

	# set custom pruning settings
	sed -i.bak 's/pruning = "default"/pruning = "custom"/g' "$APP_TOML"
	sed -i.bak 's/pruning-keep-recent = "0"/pruning-keep-recent = "2"/g' "$APP_TOML"
	sed -i.bak 's/pruning-interval = "0"/pruning-interval = "10"/g' "$APP_TOML"

	# Allocate genesis accounts (cosmos formatted addresses)
	$KIICHAIND genesis add-genesis-account "$VAL_KEY" 1000000000000000000000000000akii --keyring-backend "$KEYRING" --home "$HOMEDIR"
	$KIICHAIND genesis add-genesis-account "$USER1_KEY" 1000000000000000000000000000akii --keyring-backend "$KEYRING" --home "$HOMEDIR"

	# Sign genesis transaction
	$KIICHAIND genesis gentx "$VAL_KEY" 100000000000000000000akii --gas-prices ${BASEFEE}akii --keyring-backend "$KEYRING" --chain-id "$CHAINID" --home "$HOMEDIR"

	# Collect genesis tx
	$KIICHAIND genesis collect-gentxs --home "$HOMEDIR"

	# Run this to ensure everything worked and that the genesis file is setup correctly
	$KIICHAIND genesis validate-genesis --home "$HOMEDIR"

	if [[ $1 == "pending" ]]; then
		echo "pending mode is on, please wait for the first block committed."
	fi
fi

# Start the node
$KIICHAIND start "$TRACE" \
	--log_level $LOGLEVEL \
	--minimum-gas-prices=0.0001akii \
	--home "$HOMEDIR" \
	--json-rpc.api eth,txpool,personal,net,debug,web3 \
	--chain-id "$CHAINID"
