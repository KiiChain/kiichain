package ibc_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testkeyring "github.com/cosmos/evm/testutil/integration/os/keyring"
	"github.com/cosmos/evm/x/vm/statedb"

	kiichainApp "github.com/kiichain/kiichain/v4/app"
	ibcprecompile "github.com/kiichain/kiichain/v4/precompiles/ibc"
	"github.com/kiichain/kiichain/v4/x/tokenfactory/types"
)

type IBCPrecompileTestSuite struct {
	suite.Suite
	coordinator *ibctesting.Coordinator
	keyring     testkeyring.Keyring

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
	path   *ibctesting.Path

	// Precompile
	Precompile *ibcprecompile.Precompile
}

func TestIBCPrecompileTestSuite(t *testing.T) {
	ibcPrecompileSuite := &IBCPrecompileTestSuite{}
	suite.Run(t, ibcPrecompileSuite)
}

func (s *IBCPrecompileTestSuite) SetupTest() {
	// Set the DefaultBondDenom as default
	sdk.DefaultBondDenom = "stake"
	ibctesting.DefaultTestingAppInit = KiichainAppIniterTempDir

	// Start a new keyring
	keyring := testkeyring.New(2)
	s.keyring = keyring

	// Setup coordinator and chains
	s.coordinator = ibctesting.NewCoordinator(s.T(), 2)
	s.chainA = s.coordinator.GetChain(ibctesting.GetChainID(1))
	s.chainB = s.coordinator.GetChain(ibctesting.GetChainID(2))

	// Check if chain is KiichainApp
	chain, ok := s.coordinator.Chains[ibctesting.GetChainID(1)]
	s.Require().True(ok, "chain not found")
	_, ok = chain.App.(*kiichainApp.KiichainApp)
	s.Require().True(ok, "expected App to be KiichainApp")

	// Create path / channel betweed A and B
	path := ibctesting.NewPath(s.chainA, s.chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = "ics20-1"
	path.EndpointB.ChannelConfig.Version = "ics20-1"

	s.path = path
	s.coordinator.Setup(s.path)

	// Setup ibc precompile on chain A
	pc, err := ibcprecompile.NewPrecompile(
		chain.App.(*kiichainApp.KiichainApp).TransferKeeper, chain.App.GetIBCKeeper().ClientKeeper,
		chain.App.GetIBCKeeper().ConnectionKeeper, chain.App.GetIBCKeeper().ChannelKeeper,
		chain.App.(*kiichainApp.KiichainApp).AuthzKeeper)
	s.Require().NoError(err)
	s.Precompile = pc

	// Fund user 0 on chain A
	s.fundAddress(s.keyring.GetKey(0).AccAddr, chain)
}

// GetStateDB returns the state database for the precompile from a given chain
func (s *IBCPrecompileTestSuite) GetStateDB(chain *ibctesting.TestChain) *statedb.StateDB {
	ctx := chain.GetContext()
	// Get the header hash
	headerHash := ctx.HeaderHash()

	// Return the statedb
	return statedb.New(
		ctx,
		GetApp(chain).EVMKeeper,
		statedb.NewEmptyTxConfig(common.BytesToHash(headerHash)),
	)
}

// fundAddress adds default funds to a given address
func (s *IBCPrecompileTestSuite) fundAddress(address sdk.AccAddress, chain *ibctesting.TestChain) {
	// Define coin amount and name
	coins := sdk.NewCoins(
		ibctesting.TestCoin, // IBC test coin (1000000stake)
	)
	// Mint
	err := GetApp(chain).BankKeeper.MintCoins(
		chain.GetContext(),
		types.ModuleName,
		coins,
	)
	s.Require().NoError(err)

	// Send
	err = GetApp(chain).BankKeeper.SendCoinsFromModuleToAccount(
		chain.GetContext(),
		types.ModuleName,
		address,
		coins,
	)
	s.Require().NoError(err)
}
