package keeper

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/kiichain/kiichain3/precompiles/bank"
	"github.com/kiichain/kiichain3/precompiles/gov"
	"github.com/kiichain/kiichain3/precompiles/staking"
	"github.com/kiichain/kiichain3/precompiles/wasmd"
)

// add any payable precompiles here
// these will suppress transfer events to/from the precompile address
var payablePrecompiles = map[string]struct{}{
	bank.BankAddress:       {},
	staking.StakingAddress: {},
	gov.GovAddress:         {},
	wasmd.WasmdAddress:     {},
}

func IsPayablePrecompile(addr *common.Address) bool {
	if addr == nil {
		return false
	}
	_, ok := payablePrecompiles[addr.Hex()]
	return ok
}
