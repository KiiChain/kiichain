#!/usr/bin/env bash

set -eo pipefail

# Update any missing dependencies
go mod tidy

# Create a temporary directory for dependencies
mkdir -p tmp_deps

# Define the dependencies for the proto files generation
deps="github.com/cosmos/cosmos-sdk \
github.com/cosmos/ibc-go/v8 \
github.com/CosmWasm/wasmd \
github.com/cosmos/ibc-apps/modules/rate-limiting/v8 \
github.com/cosmos/evm"

# Copy the dependencies to the temporary directory
for dep in $deps; do
  path=$(go list -f '{{ .Dir }}' -m $dep);
  cp -r --no-preserve=mode $path tmp_deps;
done
