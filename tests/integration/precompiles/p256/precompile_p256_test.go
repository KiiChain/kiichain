package p256

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/p256"

	"github.com/kiichain/kiichain/v6/tests/integration"
)

func TestP256PrecompileTestSuite(t *testing.T) {
	s := p256.NewPrecompileTestSuite(integration.CreateKiichain)
	suite.Run(t, s)
}
