export KIICHAIND_NODE="tcp://localhost:26657"

FLAGS="--gas=2500000 --from=validator --keyring-backend=test --chain-id=local-1 --output=json --yes"

kiichaind tx wasm store ./contrib/cw_template.wasm $FLAGS
sleep 2

txhash=$(kiichaind tx wasm instantiate 1 '{"count":0}' --label=cw_template --no-admin $FLAGS | jq -r .txhash) && echo $txhash
sleep 2

addr=$(kiichaind q tx $txhash --output=json | jq -r .logs[0].events[2].attributes[0].value) && echo $addr
sleep 2

kiichaind q wasm contract $addr