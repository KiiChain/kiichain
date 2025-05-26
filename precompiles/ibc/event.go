package ibc

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cmn "github.com/cosmos/evm/precompiles/common"
)

const (
	// EventTypeTransfer define the event when a transfer is done via contract
	EventTypeTransfer = "Transfer"
)

// EmitEventTransfer emits the Transfer event
func (p *Precompile) EmitEventTransfer(
	ctx sdk.Context,
	stateDB vm.StateDB,
	caller common.Address,
	receiver, denom, port, channel string,
	amount *big.Int,
	height clienttypes.Height,
	timeoutTimestamp uint64,
	memo string,
) (err error) {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeTransfer]
	topics := make([]common.Hash, 4)

	// The first topic is the signature of the event
	topics[0] = event.ID

	// The second event is the contract address
	topics[1], err = cmn.MakeTopic(caller)
	if err != nil {
		return err
	}

	// The third event is the caller address
	topics[2], err = cmn.MakeTopic(receiver)
	if err != nil {
		return err
	}

	// The forth event is the denom used
	topics[3], err = cmn.MakeTopic(denom)
	if err != nil {
		return err
	}

	// Parse the data
	dataField, err := p.ABI.Events[EventTypeTransfer].Inputs.NonIndexed().Pack(port, channel, amount, height.RevisionNumber, height.RevisionHeight, timeoutTimestamp, memo)
	if err != nil {
		return err
	}

	// Write to the stateDB
	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        dataField,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}
