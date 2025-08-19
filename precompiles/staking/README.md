# Staking Precompile

The staking precompile enables interaction with KiiChain's staking module via EVM-compatible smart contracts, allowing developers to perform staking operations (e.g., delegate, undelegate) directly from Solidity contracts.

## Overview
- **Purpose**: Facilitates staking operations for validators and delegators in an EVM context.
- **Location**: Deployed at a precompiled contract address (e.g., `0x0000000000000000000000000000000000000800`).
- **Key Functions**:
  - `delegate(address validator, uint256 amount)`: Delegates tokens to a validator.
  - `undelegate(address validator, uint256 amount)`: Initiates undelegation of tokens.
  - `getDelegatedAmount(address delegator, address validator)`: Returns the delegated amount (view function).

## Usage
To use in a Solidity contract:
```solidity
pragma solidity ^0.8.0;
interface IStaking {
    function delegate(address validator, uint256 amount) external;
	function undelegate(address validator, uint256 amount) external;
    function getDelegatedAmount(address delegator, address validator) external view returns (uint256);
}
contract MyContract {
    IStaking staking = IStaking(0x0000000000000000000000000000000000000800);
    function delegateTokens(address validator, uint256 amount) external {
        staking.delegate(validator, amount);
    }
	function undelegateTokens(address validator, uint256 amount) external {
		staking.undelegate(validator, amount);
	}
	function myDelegationTo(address validator) external view returns (uint256) {
		return staking.getDelegatedAmount(msg.sender, validator);
	 }
}
```
### Notes
- Ensure the contract address matches KiiChain’s staking precompile address (check chain documentation).
- Requires sufficient token balance for delegation.
- See the KiiChain documentation for chain-specific details: [docs.kiiglobal.io](https://docs.kiiglobal.io/)


### This README:
- Describes the staking precompile’s purpose and key functions.
- Provides a Solidity example for usability.
- References kiichain-docs for further details, aligning with the issue’s requirements.
