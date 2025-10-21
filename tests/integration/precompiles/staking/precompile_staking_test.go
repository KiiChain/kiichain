package staking

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/staking"
	"github.com/kiichain/kiichain/v5/tests/integration"
)

func TestStakingPrecompileTestSuite(t *testing.T) {
	s := staking.NewPrecompileTestSuite(integration.CreateKiichain)
	suite.Run(t, s)
}

func TestStakingPrecompileIntegrationTestSuite(t *testing.T) {
	staking.TestPrecompileIntegrationTestSuite(t, integration.CreateKiichain)
}
