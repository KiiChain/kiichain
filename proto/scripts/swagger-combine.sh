#!/usr/bin/env bash

set -eo pipefail

# Makes a swagger temp file with reference pointers
swagger-combine ./tmp-swagger-gen/all.json -o ./client/docs/swagger-ui/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf ./tmp-swagger-gen
rm -rf ./tmp_deps