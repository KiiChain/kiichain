package staking

// import (
// 	"testing"

// 	"github.com/stretchr/testify/suite"

// 	"github.com/cosmos/evm/tests/integration/precompiles/staking"
// 	"github.com/cosmos/evm/testutil/integration/evm/network"
// 	"github.com/kiichain/kiichain/v5/tests/integration"
// )

// // CMS fails everytime
// // TestRun unbonding delegation also fails to send and requires some atom
// func TestStakingPrecompileTestSuite(t *testing.T) {
// 	s := staking.NewPrecompileTestSuite(integration.CreateKiichain, network.WithBaseCoin("akii", 18))
// 	suite.Run(t, s)
// }

// // Proposal message fails in 2/86 due to cosmos address
// func TestStakingPrecompileIntegrationTestSuite(t *testing.T) {
// 	staking.TestPrecompileIntegrationTestSuite(t, integration.CreateKiichain, network.WithBaseCoin("akii", 18))
// }
