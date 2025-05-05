package wasmd

import (
	"bytes"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// EventContractInstantiated define the event when a contract is instantiated
	EventTypeContractInstantiated = "ContractInstantiated"
	// EventContractExecuted define the event when a contract is executed
	EventTypeContractExecuted = "ContractExecuted"
)

// EmitEventContractInstantiated emits the ContractInstantiated event
func (p *Precompile) EmitEventContractInstantiated(ctx sdk.Context, stateDB vm.StateDB, contractAddress string, caller common.Address, codeID uint64) (err error) {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeContractInstantiated]
	topics := make([]common.Hash, 3)

	// The first topic is the signature of the event
	topics[0] = event.ID

	// The second event is the contract address
	topics[1], err = cmn.MakeTopic(contractAddress)
	if err != nil {
		return err
	}

	// The third event is the caller address
	topics[2], err = cmn.MakeTopic(caller)
	if err != nil {
		return err
	}

	// Prepare the event data
	var b bytes.Buffer
	b.Write(cmn.PackNum(reflect.ValueOf(codeID)))

	// Write to the stateDB
	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        b.Bytes(),
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitEventContractExecuted emits the ContractExecuted event
func (p *Precompile) EmitEventContractExecuted(ctx sdk.Context, stateDB vm.StateDB, contractAddress string, caller common.Address, data []byte) (err error) {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeContractExecuted]
	topics := make([]common.Hash, 3)

	// The first topic is the signature of the event
	topics[0] = event.ID

	// The second event is the contract address
	topics[1], err = cmn.MakeTopic(contractAddress)
	if err != nil {
		return err
	}

	// The third event is the caller address
	topics[2], err = cmn.MakeTopic(caller)
	if err != nil {
		return err
	}

	// Parse the data
	dataField, err := p.ABI.Events[EventTypeContractExecuted].Inputs.NonIndexed().Pack(data)
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
