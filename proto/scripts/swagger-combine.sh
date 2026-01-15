#!/bin/sh
set -e

# Makes a swagger temp file with reference pointers
swagger-combine ./client/docs/config.json -o ./client/docs/swagger-ui/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# Generate the json version
swagger-combine ./client/docs/config.json -o ./client/docs/swagger-ui/swagger.json -f json --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf ./tmp-swagger-gen
rm -rf ./tmp_deps
