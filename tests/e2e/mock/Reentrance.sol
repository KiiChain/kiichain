// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// Wasmd Precompile interface for queries
interface IWasmdPrecompile {
    function querySmart(
        string memory wasmContract,
        bytes memory queryData
    ) external view returns (bytes memory);

    function queryRaw(
        string memory wasmContract,
        bytes memory msg
    ) external returns (bytes memory);
}

/// @dev IWasmd contract address
address constant WASMD_PRECOMPILE_ADDRESS = 0x0000000000000000000000000000000000001001;

/// @title Reentrance gas stress test
/// @notice This contract loops precompile calls to Wasm until gas runs out.
contract Reentrance {
    /// Calls the Wasm precompile once (for regular testing)
    function reentranceQuery(
        string memory wasmBech32,
        bytes memory payload
    ) public view returns (bytes memory) {
        return
            IWasmdPrecompile(WASMD_PRECOMPILE_ADDRESS).querySmart(
                wasmBech32,
                payload
            );
    }

    /// Spin — repeatedly calls Wasm precompile to stress gas.
    /// @param wasmBech32 The Wasm contract address (bech32 string)
    /// @param payload The payload to send to querySmart()
    /// @param iterations How many times to repeat the call
    /// @return n Number of successful calls before out-of-gas
    function spin(
        string memory wasmBech32,
        bytes memory payload,
        uint256 iterations
    ) public view returns (uint256 n) {
        // Intentionally not using gas checks: will run until OOG
        for (uint256 i = 0; i < iterations; i++) {
            IWasmdPrecompile(WASMD_PRECOMPILE_ADDRESS).querySmart(
                wasmBech32,
                payload
            );
            n = i + 1;
        }
    }

    /// Minimal dummy getter (useful for payload testing)
    function ping(uint256 gasToConsume) external view returns (uint256) {
		uint256 startGas = gasleft();
		uint256 target = startGas > gasToConsume ? startGas - gasToConsume : 0;

		uint256 x = 0;
		while (gasleft() > target) {
			// burn gas
			unchecked { x+=1; }
		}
        return 42;
    }
}
