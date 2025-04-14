#!/bin/bash

set -ex

# initialize Hermes relayer configuration
mkdir -p /root/.hermes/
touch /root/.hermes/config.toml

echo $KIICHAIN_B_E2E_RLY_MNEMONIC > /root/.hermes/KIICHAIN_B_E2E_RLY_MNEMONIC.txt
echo $KIICHAIN_A_E2E_RLY_MNEMONIC > /root/.hermes/KIICHAIN_A_E2E_RLY_MNEMONIC.txt

# setup Hermes relayer configuration with non-zero gas_price
tee /root/.hermes/config.toml <<EOF
[global]
log_level = 'info'

[mode]

[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = false

[mode.channels]
enabled = true

[mode.packets]
enabled = true
clear_interval = 100
clear_on_start = true
tx_confirmation = true

[rest]
enabled = true
host = '0.0.0.0'
port = 3031

[telemetry]
enabled = true
host = '127.0.0.1'
port = 3001

[[chains]]
id = '$KIICHAIN_A_E2E_CHAIN_ID'
rpc_addr = 'http://$KIICHAIN_A_E2E_VAL_HOST:26657'
grpc_addr = 'http://$KIICHAIN_A_E2E_VAL_HOST:9090'
event_source = { mode = 'push', url = 'ws://$KIICHAIN_A_E2E_VAL_HOST:26657/websocket' , batch_delay = '50ms' }
rpc_timeout = '10s'
account_prefix = 'kii'
address_type = { derivation = 'ethermint', proto_type = { pk_type = '/cosmos.evm.crypto.v1.ethsecp256k1.PubKey' } }
key_name = 'rly01-kiichain-a'
store_prefix = 'ibc'
max_gas = 6000000
gas_price = { price = 1000000000, denom = 'akii' }
gas_multiplier = 1.5
clock_drift = '1m' # to accommodate docker containers
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
dynamic_gas_price = { enabled = true, multiplier = 1.3, max = 1000000000 }


[[chains]]
id = '$KIICHAIN_B_E2E_CHAIN_ID'
rpc_addr = 'http://$KIICHAIN_B_E2E_VAL_HOST:26657'
grpc_addr = 'http://$KIICHAIN_B_E2E_VAL_HOST:9090'
event_source = { mode = 'push', url = 'ws://$KIICHAIN_B_E2E_VAL_HOST:26657/websocket' , batch_delay = '50ms' }
rpc_timeout = '10s'
account_prefix = 'kii'
address_type = { derivation = 'ethermint', proto_type = { pk_type = '/cosmos.evm.crypto.v1.ethsecp256k1.PubKey' } }
key_name = 'rly01-kiichain-b'
store_prefix = 'ibc'
max_gas =  6000000
gas_price = { price = 1000000000, denom = 'akii' }
gas_multiplier = 1.5
clock_drift = '1m' # to accommodate docker containers
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
dynamic_gas_price = { enabled = true, multiplier = 1.3, max = 1000000000 }
EOF

# setup Hermes relayer configuration with zero gas_price
tee /root/.hermes/config-zero.toml <<EOF
[global]
log_level = 'info'

[mode]

[mode.clients]
enabled = false
refresh = true
misbehaviour = true

[mode.connections]
enabled = false

[mode.channels]
enabled = false

[mode.packets]
enabled = true
clear_interval = 100
clear_on_start = true
tx_confirmation = true

[rest]
enabled = true
host = '0.0.0.0'
port = 3031

[telemetry]
enabled = true
host = '127.0.0.1'
port = 3002

[[chains]]
id = '$KIICHAIN_A_E2E_CHAIN_ID'
rpc_addr = 'http://$KIICHAIN_A_E2E_VAL_HOST:26657'
grpc_addr = 'http://$KIICHAIN_A_E2E_VAL_HOST:9090'
event_source = { mode = 'push', url = 'ws://$KIICHAIN_A_E2E_VAL_HOST:26657/websocket' , batch_delay = '50ms' }
rpc_timeout = '10s'
account_prefix = 'kii'
address_type = { derivation = 'ethermint', proto_type = { pk_type = '/cosmos.evm.crypto.v1.ethsecp256k1.PubKey' } }
key_name = 'rly01-kiichain-a'
store_prefix = 'ibc'
max_gas = 6000000
gas_price = { price = 0, denom = 'akii' }
gas_multiplier = 1.5
clock_drift = '1m' # to accommodate docker containers
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
dynamic_gas_price = { enabled = true, multiplier = 1.3, max = 0.05 }


[[chains]]
id = '$KIICHAIN_B_E2E_CHAIN_ID'
rpc_addr = 'http://$KIICHAIN_B_E2E_VAL_HOST:26657'
grpc_addr = 'http://$KIICHAIN_B_E2E_VAL_HOST:9090'
event_source = { mode = 'push', url = 'ws://$KIICHAIN_B_E2E_VAL_HOST:26657/websocket' , batch_delay = '50ms' }
rpc_timeout = '10s'
account_prefix = 'kii'
address_type = { derivation = 'ethermint', proto_type = { pk_type = '/cosmos.evm.crypto.v1.ethsecp256k1.PubKey' } }
key_name = 'rly01-kiichain-b'
store_prefix = 'ibc'
max_gas =  6000000
gas_price = { price = 0, denom = 'akii' }
gas_multiplier = 1.5
clock_drift = '1m' # to accommodate docker containers
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
dynamic_gas_price = { enabled = true, multiplier = 1.3, max = 0.05 }
EOF

# import keys
hermes keys add --hd-path "m/44'/60'/0'/0/0" --key-name rly01-kiichain-b  --chain $KIICHAIN_B_E2E_CHAIN_ID --mnemonic-file /root/.hermes/KIICHAIN_B_E2E_RLY_MNEMONIC.txt 
sleep 5
hermes keys add --hd-path "m/44'/60'/0'/0/0" --key-name rly01-kiichain-a  --chain $KIICHAIN_A_E2E_CHAIN_ID --mnemonic-file /root/.hermes/KIICHAIN_A_E2E_RLY_MNEMONIC.txt

# Run a health check
hermes health-check