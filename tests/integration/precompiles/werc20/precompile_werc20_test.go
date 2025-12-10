package werc20

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/werc20"

	"github.com/kiichain/kiichain/v6/tests/integration"
)

func TestWERC20PrecompileUnitTestSuite(t *testing.T) {
	s := werc20.NewPrecompileUnitTestSuite(integration.CreateKiichain)
	suite.Run(t, s)
}
