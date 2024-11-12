package aclmapping

import (
	aclkeeper "github.com/cosmos/cosmos-sdk/x/accesscontrol/keeper"
	aclbankmapping "github.com/kiichain/kiichain3/aclmapping/bank"
	aclevmmapping "github.com/kiichain/kiichain3/aclmapping/evm"
	acltokenfactorymapping "github.com/kiichain/kiichain3/aclmapping/tokenfactory"
	aclwasmmapping "github.com/kiichain/kiichain3/aclmapping/wasm"
	evmkeeper "github.com/kiichain/kiichain3/x/evm/keeper"
)

type CustomDependencyGenerator struct{}

func NewCustomDependencyGenerator() CustomDependencyGenerator {
	return CustomDependencyGenerator{}
}

func (customDepGen CustomDependencyGenerator) GetCustomDependencyGenerators(evmKeeper evmkeeper.Keeper) aclkeeper.DependencyGeneratorMap {
	dependencyGeneratorMap := make(aclkeeper.DependencyGeneratorMap)
	wasmDependencyGenerators := aclwasmmapping.NewWasmDependencyGenerator()

	dependencyGeneratorMap = dependencyGeneratorMap.Merge(aclbankmapping.GetBankDepedencyGenerator())
	dependencyGeneratorMap = dependencyGeneratorMap.Merge(acltokenfactorymapping.GetTokenFactoryDependencyGenerators())
	dependencyGeneratorMap = dependencyGeneratorMap.Merge(wasmDependencyGenerators.GetWasmDependencyGenerators())
	dependencyGeneratorMap = dependencyGeneratorMap.Merge(aclevmmapping.GetEVMDependencyGenerators(evmKeeper))

	return dependencyGeneratorMap
}
