#!/bin/bash

jq '.validators = []' ~/.kiichain3/config/genesis.json > ~/.kiichain3/config/tmp_genesis.json
cd build/generated/gentx
IDX=0
for FILE in *
do
    jq '.validators['$IDX'] |= .+ {}' ~/.kiichain3/config/tmp_genesis.json > ~/.kiichain3/config/tmp_genesis_step_1.json && rm ~/.kiichain3/config/tmp_genesis.json
    KEY=$(jq '.body.messages[0].pubkey.key' $FILE -c)
    DELEGATION=$(jq -r '.body.messages[0].value.amount' $FILE)
    POWER=$(($DELEGATION / 1000000))
    jq '.validators['$IDX'] += {"power":"'$POWER'"}' ~/.kiichain3/config/tmp_genesis_step_1.json > ~/.kiichain3/config/tmp_genesis_step_2.json && rm ~/.kiichain3/config/tmp_genesis_step_1.json
    jq '.validators['$IDX'] += {"pub_key":{"type":"tendermint/PubKeyEd25519","value":'$KEY'}}' ~/.kiichain3/config/tmp_genesis_step_2.json > ~/.kiichain3/config/tmp_genesis_step_3.json && rm ~/.kiichain3/config/tmp_genesis_step_2.json
    mv ~/.kiichain3/config/tmp_genesis_step_3.json ~/.kiichain3/config/tmp_genesis.json
    IDX=$(($IDX+1))
done

mv ~/.kiichain3/config/tmp_genesis.json ~/.kiichain3/config/genesis.json

echo "Validators added to genesis"
