package bech32

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/bech32"
	"github.com/kiichain/kiichain/v5/tests/integration"
)

func TestBech32PrecompileTestSuite(t *testing.T) {
	s := bech32.NewPrecompileTestSuite(integration.CreateKiichain)
	suite.Run(t, s)
}
