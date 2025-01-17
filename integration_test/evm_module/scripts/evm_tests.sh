#!/bin/bash

set -e

cd contracts
npm ci
npx hardhat test --network local test/EVMCompatabilityTest.js
npx hardhat test --network local test/EVMPrecompileTest.js
# TODO: Re-enable me
# To re-enable we must investigate the address connections
# This will take a big development tally
# npx hardhat test --network local test/AssociateTest.js