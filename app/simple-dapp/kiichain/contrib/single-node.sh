#!/bin/sh

set -o errexit -o nounset

HOME_DIR="${1:-$HOME}"
CHAINID="test-kiichain"
USER_COINS="100000000000stake"
STAKE="100000000stake"
MONIKER="kiichain-test-node"
KIICHAIND="kiichaind"


echo "Using home dir: $HOME_DIR"
rm -rf $HOME_DIR/.kiichain
$KIICHAIND init --chain-id $CHAINID $MONIKER --home "$HOME_DIR/.kiichain"

echo "Setting up genesis file"
jq ".app_state.gov.params.voting_period = \"20s\" | .app_state.gov.params.expedited_voting_period = \"10s\" | .app_state.staking.params.unbonding_time = \"86400s\"" \
   "${HOME_DIR}/.kiichain/config/genesis.json" > \
   "${HOME_DIR}/edited_genesis.json" && mv "${HOME_DIR}/edited_genesis.json" "${HOME_DIR}/.kiichain/config/genesis.json"

$KIICHAIND keys add validator --keyring-backend="test"
$KIICHAIND keys add user --keyring-backend="test"
$KIICHAIND genesis add-genesis-account $("${KIICHAIND}" keys show validator -a --keyring-backend="test") $USER_COINS
$KIICHAIND genesis add-genesis-account $("${KIICHAIND}" keys show user -a --keyring-backend="test") $USER_COINS
$KIICHAIND genesis gentx validator $STAKE --keyring-backend="test" --chain-id $CHAINID
$KIICHAIND genesis collect-gentxs

# Set proper defaults and change ports
echo "Setting up node configs"
# sed -i '' 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:26657"#g' ~/.kiichain/config/config.toml
sleep 1
sed -i -r 's/index_all_keys = false/index_all_keys = true/g' ~/.kiichain/config/config.toml
sed -i -r 's/minimum-gas-prices = ""/minimum-gas-prices = "0stake"/g' ~/.kiichain/config/app.toml

# Start the kiichain
$KIICHAIND start --api.enable=true
