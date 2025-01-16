#!/bin/bash

set -e

cd contracts
npm ci
npx hardhat test --network local test/EVMCompatabilityTest.js
npx hardhat test --network local test/EVMPrecompileTest.js
npx hardhat test --network local test/AssociateTest.js