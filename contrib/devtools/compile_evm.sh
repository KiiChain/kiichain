#!/bin/bash

# Set the variables for the contract compilation
CONTRACT_TO_COMPILE=$1
OUTPUT_DIR=tests/e2e/mock

# Compile the contract using EVM
solc --abi --bin $CONTRACT_TO_COMPILE -o $OUTPUT_DIR

# Remove all the JSON files from the output
find $OUTPUT_DIR -type f -name "*.json" -delete

# Create the go bindings using abigen
CONTRACT_NAME=$(basename $CONTRACT_TO_COMPILE .sol)
abigen --bin=$OUTPUT_DIR/$CONTRACT_NAME.bin \
	--abi=$OUTPUT_DIR/$CONTRACT_NAME.abi \
	--pkg=mock --out=$OUTPUT_DIR/$CONTRACT_NAME.go \
	--type=$CONTRACT_NAME
