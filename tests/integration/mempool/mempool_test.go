package mempool

import (
	"testing"

	"github.com/kiichain/kiichain/v5/tests/integration"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/mempool"
	"github.com/cosmos/evm/testutil/integration/evm/network"
)

func TestMempoolIntegrationTestSuite(t *testing.T) {
	suite.Run(t, mempool.NewMempoolIntegrationTestSuite(integration.CreateKiichain, network.WithBaseCoin("akii", 18)))
}
