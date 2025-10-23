package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/x/feemarket"
)

func TestFeeMarketKeeperTestSuite(t *testing.T) {
	s := feemarket.NewTestKeeperTestSuite(CreateKiichain)
	suite.Run(t, s)
}
