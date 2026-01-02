package slashing

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/slashing"

	"github.com/kiichain/kiichain/v7/tests/integration"
)

func TestSlashingPrecompileTestSuite(t *testing.T) {
	s := slashing.NewPrecompileTestSuite(integration.CreateKiichain)
	suite.Run(t, s)
}
