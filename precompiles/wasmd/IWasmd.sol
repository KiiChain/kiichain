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
    /// @param contractAddress The address of the newly instantiated contract.
    /// @param caller The address of the caller that instantiated the contract.
    /// @param codeID The code id of the contract.
    event ContractInstantiated(string indexed contractAddress, address indexed caller, uint64 codeID);

    /// @dev This event is emitted when a contract is executed on the Wasmd protocol.
    /// @param contractAddress The address of the contract that was executed.
    /// @param caller The address of the caller that executed the contract.
    /// @param msg The message that was sent to the contract.
    event ContractExecuted(string indexed contractAddress, address indexed caller, bytes msg);

    /// @dev This function is used to instantiate a new contract on the Wasmd protocol.
    /// @param admin The admin of the contract.
    /// @param codeID The code id of the contract.
    /// @param label The label of the contract.
    /// @param msg The init message of the contract.
    /// @param coins The funds to be sent to the contract.
    /// @return contractAddress The address of the newly instantiated contract.
    /// @return data The data returned from the contract instantiation.
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
    ) external returns (string memory contractAddress, bytes memory data);

    /// @dev This function is used to execute a contract on the Wasmd protocol.
    /// @param contractAddress The address of the contract to execute.
    /// @param msg The message to send to the contract.
    /// @param coins The funds to be sent to the contract.
    /// @return data The data returned from the contract execution.
    function execute(
        // the contract address to execute
        string memory contractAddress,
        // the message to send to the contract
        bytes memory msg,
        // the funds to send to the contract
        Coin[] memory coins
    ) external returns (bytes memory data);

    /// @dev This function is used to query a contract on the Wasmd protocol using a raw query.
    /// @param contractAddress The address of the contract to query.
    /// @param queryData The message to send to the contract.
    /// @return data The data returned from the contract query.
    function queryRaw(
        // the contract address to query
        string memory contractAddress,
        // the message to send to the contract
        bytes memory queryData
    ) external returns (bytes memory data);

    /// @dev This function is used to query a contract on the Wasmd protocol using a smart query.
    /// @param contractAddress The address of the contract to query.
    /// @param msg The message to send to the contract.
    /// @return data The data returned from the contract query.`
    function querySmart(
        // the contract address to query
        string memory contractAddress,
        // the message to send to the contract
        bytes memory msg
    ) external returns (bytes memory data);
}
