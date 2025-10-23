package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/x/vm"
	"github.com/cosmos/evm/testutil/integration/evm/network"
)

func TestNestedEVMExtensionCallSuite(t *testing.T) {
	s := vm.NewNestedEVMExtensionCallSuite(CreateKiichain, network.WithBaseCoin("akii", 18))
	suite.Run(t, s)
}

func TestGenesisTestSuite(t *testing.T) {
	s := vm.NewGenesisTestSuite(CreateKiichain, network.WithBaseCoin("akii", 18))
	suite.Run(t, s)
}

func TestVmAnteTestSuite(t *testing.T) {
	s := vm.NewEvmAnteTestSuite(CreateKiichain)
	suite.Run(t, s)
}

func TestIterateContracts(t *testing.T) {
	vm.TestIterateContracts(t, CreateKiichain, network.WithBaseCoin("akii", 18))
}
