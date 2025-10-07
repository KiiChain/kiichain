# EVM v0.1 to v0.3 upgrade
PR with changes can be found here:
- [Initial changes](https://github.com/KiiChain/kiichain/pull/110)
- [Upgrade](https://github.com/KiiChain/kiichain/pull/120)
- [Bulk of version changes](https://github.com/KiiChain/kiichain/pull/115)
- [Cmd d chain id flag override](https://github.com/KiiChain/kiichain/pull/121)

It's important to highlight we went with a different approach for the chain Id changes, mimicking mantrachain. This required some back and forth fixes we were not expecting. 
We are also using our own fork of cosmos/evm  v0.4.1 with changes to fee collection to allow feeless transactions.

## Requirements
- IBC v10
### Import changes
- Update vm core usage:
  - "github.com/cosmos/evm/x/vm/core/vm" -> "github.com/ethereum/go-ethereum/core/vm"
- Replace go ethereum with fork in go.mod:
```go
	// use Cosmos geth fork
	github.com/ethereum/go-ethereum => github.com/cosmos/go-ethereum v1.15.11-cosmos-0
```
### Config changes
- IDs are uint64 instead of strings
- We followed Mantrachain fetching of the chain ID at init. Most of the changes are found on [this file](https://github.com/KiiChain/kiichain/blob/feat/upgrade-v5/app/config.go)

### App changes
Usage of the int chain ID as a parameter for the configs
```go
	// Use the EVM encoding config
	encodingConfig := evmencoding.MakeConfig(KiichainID)
...
	// initialize the Cosmos EVM application configuration
	if err := evmAppOptions(KiichainID); err != nil {
		panic(err)
	}

```

### Cmd root.go changes
In both E2E test setup and the root, we must configure the evm config to utilize the chain ID correctly.
```go
func initAppConfig(evmChainID uint64) (string, interface{}) {
	// Can optionally overwrite the SDK's default server config.
	srvCfg := serverconfig.DefaultConfig()
	srvCfg.StateSync.SnapshotInterval = 1000
	srvCfg.StateSync.SnapshotKeepRecent = 10

	// Setup evm chain ID
	evmCfg := evmserverconfig.DefaultEVMConfig()
	evmCfg.EVMChainID = evmChainID

	customAppConfig := CustomAppConfig{
		Config:  *srvCfg,
		EVM:     *evmCfg,
		JSONRPC: *evmserverconfig.DefaultJSONRPCConfig(),
		TLS:     *evmserverconfig.DefaultTLSConfig(),
		Wasm:    wasmtypes.DefaultNodeConfig(),
	}

	// Default template
	defaultAppTemplate := serverconfig.DefaultConfigTemplate + wasmtypes.DefaultConfigTemplate()

	// EVM template
	defaultAppTemplate += evmserverconfig.DefaultEVMConfigTemplate

	return defaultAppTemplate, customAppConfig
}
```

### On Precompiles:
- Switch transfer keeper `transferkeeper "github.com/cosmos/evm/x/ibc/transfer/keeper"`
- Remove authzKeper and ApprovalExpiration from cmn.Precompile inheritance on NewPrecompiles
  - This will require a change on integration tests
  - And on precompile keepers
- Remove snapshot return from `p.RunSetup(evm, contract, readOnly, p.IsTransaction)` and its journal entry
```go
	if err := p.AddJournalEntries(stateDB, snapshot); err != nil {
		return nil, err
	}
```
- Add extra argument to cost calculation in `Run` functions:
```go
	if !contract.UseGas(cost, nil, tracing.GasChangeCallPrecompiledContract) {
		return nil, vm.ErrOutOfGas
	}
```
- On `tx_test.go` files, replace `testutil.NewPrecompileContract` precompile usage with address:
```go
  s.precompile.Address()
```

### Upgrade
For the upgrade, we needed to use an old proto file to migrate old params to a new version.
The old proto file can be found [here](https://github.com/KiiChain/kiichain/blob/main/app/upgrades/v5_0/evm.pb.go)

We also used the default value for `evmParams.AccessControl` since we did not use that field and the types mismatched, making it harder to migrate.

```go
// MigrateEVMParams imports relevant old v0.1 params and sets them on new EVM param type
func MigrateEVMParams(
	ctx sdk.Context,
	keepers *keepers.AppKeepers,
) error {
	storekeys := keepers.GetKVStoreKey()
	store := runtime.NewKVStoreService(storekeys[evmtypes.StoreKey]).OpenKVStore(ctx)

	// Fetch byte data of old params
	oldData, err := store.Get(evmtypes.KeyPrefixParams)
	if err != nil {
		return err
	}

	// Read old params
	var oldParams Params
	if oldData != nil {
		if err := oldParams.Unmarshal(oldData); err != nil {
			return err
		}
	}

	// set the evm/vm params
	evmParams := evmtypes.DefaultParams()
	evmParams.EvmDenom = evmtypes.GetEVMCoinDenom()
	evmParams.ActiveStaticPrecompiles = oldParams.ActiveStaticPrecompiles
	evmParams.EVMChannels = oldParams.EVMChannels
	evmParams.AllowUnprotectedTxs = oldParams.AllowUnprotectedTxs

	return keepers.EVMKeeper.SetParams(ctx, evmParams)
}
```

The ERC20 also got a new param

```go
// Add missing ERC20 param
params := keepers.Erc20Keeper.GetParams(ctx)
params.PermissionlessRegistration = false
err = keepers.Erc20Keeper.SetParams(ctx, params)
if err != nil {
	return err
}
```

# EVM v0.3.1 -> v0.4.1 missing notes
- CallEVM now takes CapGas, which can be used as `nil`
- testkeyring changed to `testkeyring "github.com/cosmos/evm/testutil/keyring"`