package ante

import (
	"testing"

	"github.com/cosmos/evm/tests/integration/ante"
	"github.com/kiichain/kiichain/v5/tests/integration"
)

func TestAnte_Integration(t *testing.T) {
	ante.TestIntegrationAnteHandler(t, integration.CreateKiichain)
}
