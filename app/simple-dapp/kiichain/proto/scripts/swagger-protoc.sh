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
rm -rf ./tmp-swagger-gen/cosmos/app
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

# Makes a swagger temp file with reference pointers
swagger-combine ./client/docs/config.json -o ./client/docs/swagger-ui/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true
# Generate the json version
swagger-combine ./client/docs/config.json -o ./client/docs/swagger-ui/swagger.json -f json --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf ./tmp-swagger-gen
rm -rf ./tmp_deps