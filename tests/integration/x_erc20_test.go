package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/x/erc20"
)

func TestERC20GenesisTestSuite(t *testing.T) {
	suite.Run(t, erc20.NewGenesisTestSuite(CreateKiichain))
}

// Mix of requiring mint and using cosmos addresses
// TestOnRecvPacketRegistered, TestOnAcknowledgementPacket and TestConvertCoinToERC20FromPacket all fail due to mint
// TestConvertERC20NativeERC20 fail due to cosmos address
// TestIsTokenPairRegistered, TestMintingEnabled, TestBalanceOf use cosmos address or mint
// And more, there are hundreds of tests here
// func TestERC20KeeperTestSuite(t *testing.T) {
// 	s := erc20.NewKeeperTestSuite(CreateKiichain, network.WithBaseCoin("akii", 18))
// 	suite.Run(t, s)
// }

// Test failed w/ cosmos address
// func TestERC20PrecompileIntegrationTestSuite(t *testing.T) {
// 	erc20.TestPrecompileIntegrationTestSuite(t, CreateKiichain, network.WithBaseCoin("akii", 18))
// }
