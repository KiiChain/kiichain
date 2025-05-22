// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/// @dev IWasmd contract address
address constant IBC_PRECOMPILE_ADDRESS = 0x0000000000000000000000000000000000001002;

IBC constant IBC_CONTRACT = IBC(
    IBC_PRECOMPILE_ADDRESS
);

/// @author Kiichain Team
/// @title IBC Precompiles Contract
/// @dev This contract is a precompiled contract that provides a set of functions for transfering with IBC.
/// @custom:address 0x0000000000000000000000000000000000001002
interface IBC {
    /// @dev This event is emitted when a transfer is done via IBC with this precompile.
    /// @param caller The caller of the transfer.
    /// @param toAddress The receiver address.
    /// @param port The IBC port.
    /// @param channel The IBC channel.
    /// @param denom The coin denom.
    /// @param amount The amount of given coin to be transferred.
    /// @param revisionNumber The revision number for the transfer.
    /// @param revisionHeight The revision height for the transfer.
    /// @param timeoutTimestamp The timeout timestamp for the transfer.
    event Transfer(
        address indexed caller,
        string indexed toAddress,
        string indexed denom,
        string port,
        string channel,
        uint256 amount,
        uint64 revisionNumber,
        uint64 revisionHeight,
        uint64 timeoutTimestamp
    );

    /// @dev This function is used to transfer a given coin via IBC.
    /// @param toAddress The receiver address.
    /// @param port The IBC port.
    /// @param channel The IBC channel.
    /// @param denom The coin denom.
    /// @param amount The amount of given coin to be transferred.
    /// @param revisionNumber The revision number for the transfer.
    /// @param revisionHeight The revision height for the transfer.
    /// @param timeoutTimestamp The timeout timestamp for the transfer.
    /// @param memo Memo message for the transfer.
    /// @return success A boolean indicating whether the transfer was successful.
    function transfer(
        string memory toAddress,
        string memory port,
        string memory channel,
        string memory denom,
        uint256 amount,
        uint64 revisionNumber,
        uint64 revisionHeight,
        uint64 timeoutTimestamp,
        string memory memo
    ) external returns (bool success);

    /// @dev This function is used to transfer a given coin via IBC.
    /// @param toAddress The receiver address.
    /// @param port The IBC port.
    /// @param channel The IBC channel.
    /// @param denom The coin denom.
    /// @param amount The amount of given coin to be transferred.
    /// @param memo Memo message for the transfer.
    /// @return success A boolean indicating whether the transfer was successful.
    function transferWithDefaultTimeout(
        string memory toAddress,
        string memory port,
        string memory channel,
        string memory denom,
        uint256 amount,
        string memory memo
    ) external returns (bool success);
}