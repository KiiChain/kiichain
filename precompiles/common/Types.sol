// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

/// @dev Coin is a struct that represents a token with a denomination and an amount.
struct Coin {
    string denom;
    uint256 amount;
}
