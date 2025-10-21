package integration

import (
	"testing"

	"github.com/cosmos/evm/tests/integration/x/vm"

	"github.com/stretchr/testify/suite"
)

func TestKeeperTestSuite(t *testing.T) {
	s := vm.NewKeeperTestSuite(CreateKiichain)
	s.EnableFeemarket = false
	s.EnableLondonHF = true
	suite.Run(t, s)
}

func TestNestedEVMExtensionCallSuite(t *testing.T) {
	s := vm.NewNestedEVMExtensionCallSuite(CreateKiichain)
	suite.Run(t, s)
}

func TestGenesisTestSuite(t *testing.T) {
	s := vm.NewGenesisTestSuite(CreateKiichain)
	suite.Run(t, s)
}

func TestVmAnteTestSuite(t *testing.T) {
	s := vm.NewEvmAnteTestSuite(CreateKiichain)
	suite.Run(t, s)
}

func TestIterateContracts(t *testing.T) {
	vm.TestIterateContracts(t, CreateKiichain)
}
