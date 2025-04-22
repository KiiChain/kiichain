#!/usr/bin/env bash

set -eo pipefail

mkdir -p ./tmp-swagger-gen
proto_dirs=$(find ./proto ./tmp_deps -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    buf generate --template proto/buf.gen.swagger.yaml $query_file
  fi
done

# Remove files we will not use
rm -rf ./tmp-swagger-gen/cosmos/mint
rm -rf ./tmp-swagger-gen/cosmos/nft
rm -rf ./tmp-swagger-gen/cosmos/autocli
rm -rf ./tmp-swagger-gen/cosmos/circuit
rm -rf ./tmp-swagger-gen/cosmos/group
rm -rf ./tmp-swagger-gen/cosmos/orm
rm -rf ./tmp-swagger-gen/cosmos/params
rm -rf ./tmp-swagger-gen/cosmos/query
rm -rf ./tmp-swagger-gen/ibc/lightclients/wasm
rm -rf ./tmp-swagger-gen/testpb

# Convert all *.swagger.json files into a single folder _all
files=$(find ./tmp-swagger-gen -name '*.swagger.json' -print0 | xargs -0)
mkdir -p ./tmp-swagger-gen/_all
counter=0
for f in $files; do
  echo "[+] $f"

  # check gaia first before cosmos
  if [[ "$f" =~ "kiichain" ]]; then
    cp $f ./tmp-swagger-gen/_all/kiichain-$counter.json
  elif [[ "$f" =~ "cosmos" ]]; then
    cp $f ./tmp-swagger-gen/_all/cosmos-$counter.json
  elif [[ "$f" =~ "cosmwasm" ]]; then
    cp $f ./tmp-swagger-gen/_all/cosmwasm-$counter.json
  elif [[ "$f" =~ "ratelimit" ]]; then
    cp $f ./tmp-swagger-gen/_all/ratelimit-$counter.json
  elif [[ "$f" =~ "ibc" ]]; then
    cp $f ./tmp-swagger-gen/_all/ibc-$counter.json
  else
    cp $f ./tmp-swagger-gen/_all/other-$counter.json
  fi

  counter=$(expr $counter + 1)
done
