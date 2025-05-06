/// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../common/Types.sol";

/// @dev IWasmd contract address
address constant WASMD_PRECOMPILER_ADDRESS = 0x0000000000000000000000000000000000001001;

/// @author Kiichain Team
/// @title Wasmd Precompiles Contract
/// @dev This contract is a precompiled contract that provides a set of functions for interacting with the Wasmd protocol.
/// @custom:address 0x0000000000000000000000000000000000001001
interface IWasmd {
    /// @dev This event is emitted when a contract is instantiated on the Wasmd protocol.
    /// @param caller The address of the caller that instantiated the contract.
    /// @param codeID The code id of the contract.
    /// @param contractAddress The address of the contract that was instantiated.
    event ContractInstantiated(address indexed caller, uint64 indexed codeID, string contractAddress, bytes data);

    /// @dev This event is emitted when a contract is executed on the Wasmd protocol.
    /// @param contractAddress The address of the contract that was executed.
    /// @param caller The address of the caller that executed the contract.
    /// @param data The data returned from the contract execution.
    event ContractExecuted(string indexed contractAddress, address indexed caller, bytes data);

    /// @dev This function is used to instantiate a new contract on the Wasmd protocol.
    /// @param admin The address of the admin of the contract.
    /// @param codeID The code id of the contract.
    /// @param label The label of the contract.
    /// @param msg The init message of the contract.
    /// @param coins The funds to be sent to the contract.
    /// @return success A boolean indicating whether the instantiation was successful.
    function instantiate(
        // the admin of the contract
        address admin,
        // the code id of the contract
        uint64 codeID,
        // the label of the contract
        string memory label,
        // the init message of the contract
        bytes memory msg,
        // the funds to be sent to the contract
        Coin[] memory coins
    ) external returns (bool success);

    /// @dev This function is used to execute a contract on the Wasmd protocol.
    /// @param contractAddress The address of the contract to execute.
    /// @param msg The message to send to the contract.
    /// @param coins The funds to send to the contract.
    /// @return success A boolean indicating whether the execution
    function execute(
        // the contract address to execute
        string memory contractAddress,
        // the message to send to the contract
        bytes memory msg,
        // the funds to send to the contract
        Coin[] memory coins
    ) external returns (bool success);

    /// @dev This function is used to query a contract on the Wasmd protocol using a raw query.
    /// @param contractAddress The address of the contract to query.
    /// @param queryData The message to send to the contract.
    /// @return data The data returned from the contract query.
    function queryRaw(
        // the contract address to query
        string memory contractAddress,
        // the message to send to the contract
        bytes memory queryData
    ) external view returns (bytes memory data);

    /// @dev This function is used to query a contract on the Wasmd protocol using a smart query.
    /// @param contractAddress The address of the contract to query.
    /// @param msg The message to send to the contract.
    /// @return data The data returned from the contract query.`
    function querySmart(
        // the contract address to query
        string memory contractAddress,
        // the message to send to the contract
        bytes memory msg
    ) external view returns (bytes memory data);
}
