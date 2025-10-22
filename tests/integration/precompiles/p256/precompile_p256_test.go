package p256

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/p256"
	"github.com/cosmos/evm/testutil/integration/evm/network"
	"github.com/kiichain/kiichain/v5/tests/integration"
)

func TestP256PrecompileTestSuite(t *testing.T) {
	s := p256.NewPrecompileTestSuite(integration.CreateKiichain)
	suite.Run(t, s)
}

func TestP256PrecompileIntegrationTestSuite(t *testing.T) {
	p256.TestPrecompileIntegrationTestSuite(t, integration.CreateKiichain, network.WithBaseCoin("akii", 18))
}
