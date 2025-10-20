package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/slashing"
	"github.com/cosmos/evm/testutil/integration/evm/network"
)

func TestSlashingPrecompileTestSuite(t *testing.T) {
	s := slashing.NewPrecompileTestSuite(CreateKiichain, network.WithBaseCoin("akii", 18))
	suite.Run(t, s)
}

func TestSlashingPrecompileIntegrationTestSuite(t *testing.T) {
	slashing.TestPrecompileIntegrationTestSuite(t, CreateKiichain, network.WithBaseCoin("akii", 18))
}
