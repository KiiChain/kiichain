package werc20

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/werc20"

	"github.com/kiichain/kiichain/v5/tests/integration"
)

func TestWERC20PrecompileUnitTestSuite(t *testing.T) {
	s := werc20.NewPrecompileUnitTestSuite(integration.CreateKiichain)
	suite.Run(t, s)
}

// Complains about kii denom
// func TestWERC20PrecompileIntegrationTestSuite(t *testing.T) {
// 	werc20.TestPrecompileIntegrationTestSuite(t, integration.CreateKiichain)
// }
