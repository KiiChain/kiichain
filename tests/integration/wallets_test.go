package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/wallets"
)

func TestLedgerTestSuite(t *testing.T) {
	s := wallets.NewLedgerTestSuite(CreateKiichain)
	suite.Run(t, s)
}
