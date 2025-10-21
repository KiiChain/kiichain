package erc20

import (
	"testing"

	"github.com/stretchr/testify/suite"

	erc21 "github.com/cosmos/evm/tests/integration/precompiles/erc20"
	"github.com/kiichain/kiichain/v5/tests/integration"
)

func TestErc20PrecompileTestSuite(t *testing.T) {
	s := erc21.NewPrecompileTestSuite(integration.CreateKiichain)
	suite.Run(t, s)
}

func TestErc20IntegrationTestSuite(t *testing.T) {
	erc21.TestIntegrationTestSuite(t, integration.CreateKiichain)
}
