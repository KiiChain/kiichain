package integration

import (
	"testing"

	"github.com/cosmos/evm/tests/integration/precompiles/bank"
	"github.com/stretchr/testify/suite"
)

func TestBankPrecompileTestSuite(t *testing.T) {
	s := bank.NewPrecompileTestSuite(CreateKiichain)
	suite.Run(t, s)
}

func TestBankPrecompileIntegrationTestSuite(t *testing.T) {
	bank.TestIntegrationSuite(t, CreateKiichain)
}
