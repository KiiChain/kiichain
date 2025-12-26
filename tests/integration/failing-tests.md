# The following tests are failing:

## VM Test
On the `TestKeeperTestSuite` several tests are failing because they need the mint module or the precise bank. 

The test `vm.NewNestedEVMExtensionCallSuite(CreateKiichain, network.WithBaseCoin("akii", 18))` was removed due to a validator requiring balance it does not have. I could not find a straightforward way to add balance to it. It only contains a test for flash bound exploit.

This test also needs a test flag `-tags=test`, as well as `network.WithBaseCoin("akii", 18)` being passed.

## ERC20 Tests
`TestERC20KeeperTestSuite` has hundreds of tests.
- All Packet related fail due to mint module not being there
  - `TestOnRecvPacketRegistered`, `TestOnAcknowledgementPacket` and `TestConvertCoinToERC20FromPacket`
- `TestMintingEnabled` needs mint module
- `TestIsTokenPairRegistered`and `TestBalanceOf` use a cosmos address

`TestERC20PrecompileIntegrationTestSuite` Fails due to using a cosmos address.

## Wallet tests
`TestLedgerTestSuite` has hardcoded cosmos prefix on tests.

## IBC
`Case_no-op_-_disabled_erc20_by_params` and `Case_error_-_disabled_erc20_by_params` use a cosmos address.

Test requires `network.WithBaseCoin("akii", 18)`

## EIP712
Requires unsealed config.

## Mempool
About 70% pass, but 30% do cosmos transactions using fee as Aatom.

## Gov
On `TestGovPrecompileTestSuite`:
- `TestGovPrecompileTestSuite/TestGetDeposit` uses a cosmos address
- `TestGetTallyResult/valid_query` uses a cosmos address

`TestGovPrecompileIntegrationTestSuite` has aatom hardcoded in it.

## Slashing
`TestSlashingPrecompileIntegrationTestSuite` fails with a 'condition not met' error.

## Staking
On `TestStakingPrecompileTestSuite`:
- All CMS tests fails due to kii denom being bound
- `TestRun unbonding delegation` requires aatom

`TestStakingPrecompileIntegrationTestSuite` has 2/86 tests using a cosmos address.

## Werc20 
`TestWERC20PrecompileIntegrationTestSuite` tests complain about kii denom being bound

## p256
`TestP256PrecompileIntegrationTestSuite` failing due to gas being slightly under what it wants
