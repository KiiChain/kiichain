#!/bin/bash

set -e

cd contracts
npm ci
npx hardhat test --network local test/CW20toERC20PointerTest.js
npx hardhat test --network local test/ERC20toCW20PointerTest.js
npx hardhat test --network local test/ERC20toNativePointerTest.js
npx hardhat test --network local test/CW721toERC721PointerTest.js
npx hardhat test --network local test/ERC721toCW721PointerTest.js
