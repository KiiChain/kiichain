/// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

/// @dev IOracle contract address
address constant ORACLE_PRECOMPILER_ADDRESS = 0x0000000000000000000000000000000000001003;

/// @author Kiichain Team
/// @title Oracle Precompiles Contract
/// @dev This contract is a precompiled contract that provides a set of functions for interacting with the Oracle
/// @custom:address 0x0000000000000000000000000000000000001003
interface IOracle {
    /// @dev Get the exchange rate for a specific denomination
    /// @param denom The denomination for which to get the exchange rate
    /// @return rate The exchange rate for the specified denomination
    /// @return lastUpdate The block number when the exchange rate was last updated
    /// @return lastUpdateTimestamp The timestamp when the exchange rate was last updated
    function getExchangeRate(
        string memory denom
    )
        external
        view
        returns (
            string memory rate,
            string memory lastUpdate,
            int64 lastUpdateTimestamp
        );

    /// @dev Get the exchange rates for all denominations
    /// @return denoms An array of all denominations
    /// @return rates An array of exchange rates corresponding to the denominations
    /// @return lastUpdate An array of block numbers when each exchange rate was last updated
    /// @return lastUpdateTimestamps An array of timestamps when each exchange rate was last updated
    function getExchangeRates()
        external
        view
        returns (
            string[] memory denoms,
            string[] memory rates,
            string[] memory lastUpdate,
            uint256[] memory lastUpdateTimestamps
        );

    /// @dev Get the TWAP (Time-Weighted Average Price) for a specific lookback period
    /// @param lookbackSeconds The number of seconds to look back for the TWAP calculation
    /// @return denoms An array of denominations for which the TWAP is calculated
    /// @return twaps An array of TWAP values corresponding to the denominations
    function getTwaps(
        uint256 lookbackSeconds
    ) external view returns (string[] memory denoms, string[] memory twaps);
}
