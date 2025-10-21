package bank

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/bank"
	"github.com/kiichain/kiichain/v5/tests/integration"
)

func TestBankPrecompileTestSuite(t *testing.T) {
	s := bank.NewPrecompileTestSuite(integration.CreateKiichain)
	suite.Run(t, s)
}

func TestBankPrecompileIntegrationTestSuite(t *testing.T) {
	bank.TestIntegrationSuite(t, integration.CreateKiichain)
}
