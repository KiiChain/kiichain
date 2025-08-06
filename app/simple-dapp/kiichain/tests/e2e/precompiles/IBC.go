// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package precompiles

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// IbcPrecompileMetaData contains all meta data concerning the IbcPrecompile contract.
var IbcPrecompileMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"denom\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"port\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"channel\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"revisionNumber\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"revisionHeight\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"timeoutTimestamp\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"memo\",\"type\":\"string\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"port\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"channel\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"denom\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint64\",\"name\":\"revisionNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"revisionHeight\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"timeoutTimestamp\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"memo\",\"type\":\"string\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"port\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"channel\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"denom\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"memo\",\"type\":\"string\"}],\"name\":\"transferWithDefaultTimeout\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// IbcPrecompileABI is the input ABI used to generate the binding from.
// Deprecated: Use IbcPrecompileMetaData.ABI instead.
var IbcPrecompileABI = IbcPrecompileMetaData.ABI

// IbcPrecompile is an auto generated Go binding around an Ethereum contract.
type IbcPrecompile struct {
	IbcPrecompileCaller     // Read-only binding to the contract
	IbcPrecompileTransactor // Write-only binding to the contract
	IbcPrecompileFilterer   // Log filterer for contract events
}

// IbcPrecompileCaller is an auto generated read-only Go binding around an Ethereum contract.
type IbcPrecompileCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IbcPrecompileTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IbcPrecompileTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IbcPrecompileFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IbcPrecompileFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IbcPrecompileSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IbcPrecompileSession struct {
	Contract     *IbcPrecompile    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IbcPrecompileCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IbcPrecompileCallerSession struct {
	Contract *IbcPrecompileCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// IbcPrecompileTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IbcPrecompileTransactorSession struct {
	Contract     *IbcPrecompileTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// IbcPrecompileRaw is an auto generated low-level Go binding around an Ethereum contract.
type IbcPrecompileRaw struct {
	Contract *IbcPrecompile // Generic contract binding to access the raw methods on
}

// IbcPrecompileCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IbcPrecompileCallerRaw struct {
	Contract *IbcPrecompileCaller // Generic read-only contract binding to access the raw methods on
}

// IbcPrecompileTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IbcPrecompileTransactorRaw struct {
	Contract *IbcPrecompileTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIbcPrecompile creates a new instance of IbcPrecompile, bound to a specific deployed contract.
func NewIbcPrecompile(address common.Address, backend bind.ContractBackend) (*IbcPrecompile, error) {
	contract, err := bindIbcPrecompile(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IbcPrecompile{IbcPrecompileCaller: IbcPrecompileCaller{contract: contract}, IbcPrecompileTransactor: IbcPrecompileTransactor{contract: contract}, IbcPrecompileFilterer: IbcPrecompileFilterer{contract: contract}}, nil
}

// NewIbcPrecompileCaller creates a new read-only instance of IbcPrecompile, bound to a specific deployed contract.
func NewIbcPrecompileCaller(address common.Address, caller bind.ContractCaller) (*IbcPrecompileCaller, error) {
	contract, err := bindIbcPrecompile(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IbcPrecompileCaller{contract: contract}, nil
}

// NewIbcPrecompileTransactor creates a new write-only instance of IbcPrecompile, bound to a specific deployed contract.
func NewIbcPrecompileTransactor(address common.Address, transactor bind.ContractTransactor) (*IbcPrecompileTransactor, error) {
	contract, err := bindIbcPrecompile(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IbcPrecompileTransactor{contract: contract}, nil
}

// NewIbcPrecompileFilterer creates a new log filterer instance of IbcPrecompile, bound to a specific deployed contract.
func NewIbcPrecompileFilterer(address common.Address, filterer bind.ContractFilterer) (*IbcPrecompileFilterer, error) {
	contract, err := bindIbcPrecompile(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IbcPrecompileFilterer{contract: contract}, nil
}

// bindIbcPrecompile binds a generic wrapper to an already deployed contract.
func bindIbcPrecompile(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IbcPrecompileMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IbcPrecompile *IbcPrecompileRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IbcPrecompile.Contract.IbcPrecompileCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IbcPrecompile *IbcPrecompileRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IbcPrecompile.Contract.IbcPrecompileTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IbcPrecompile *IbcPrecompileRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IbcPrecompile.Contract.IbcPrecompileTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IbcPrecompile *IbcPrecompileCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IbcPrecompile.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IbcPrecompile *IbcPrecompileTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IbcPrecompile.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IbcPrecompile *IbcPrecompileTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IbcPrecompile.Contract.contract.Transact(opts, method, params...)
}

// Transfer is a paid mutator transaction binding the contract method 0x98a955c7.
//
// Solidity: function transfer(string receiver, string port, string channel, string denom, uint256 amount, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo) returns(bool success)
func (_IbcPrecompile *IbcPrecompileTransactor) Transfer(opts *bind.TransactOpts, receiver string, port string, channel string, denom string, amount *big.Int, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string) (*types.Transaction, error) {
	return _IbcPrecompile.contract.Transact(opts, "transfer", receiver, port, channel, denom, amount, revisionNumber, revisionHeight, timeoutTimestamp, memo)
}

// Transfer is a paid mutator transaction binding the contract method 0x98a955c7.
//
// Solidity: function transfer(string receiver, string port, string channel, string denom, uint256 amount, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo) returns(bool success)
func (_IbcPrecompile *IbcPrecompileSession) Transfer(receiver string, port string, channel string, denom string, amount *big.Int, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string) (*types.Transaction, error) {
	return _IbcPrecompile.Contract.Transfer(&_IbcPrecompile.TransactOpts, receiver, port, channel, denom, amount, revisionNumber, revisionHeight, timeoutTimestamp, memo)
}

// Transfer is a paid mutator transaction binding the contract method 0x98a955c7.
//
// Solidity: function transfer(string receiver, string port, string channel, string denom, uint256 amount, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo) returns(bool success)
func (_IbcPrecompile *IbcPrecompileTransactorSession) Transfer(receiver string, port string, channel string, denom string, amount *big.Int, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string) (*types.Transaction, error) {
	return _IbcPrecompile.Contract.Transfer(&_IbcPrecompile.TransactOpts, receiver, port, channel, denom, amount, revisionNumber, revisionHeight, timeoutTimestamp, memo)
}

// TransferWithDefaultTimeout is a paid mutator transaction binding the contract method 0x97ba602b.
//
// Solidity: function transferWithDefaultTimeout(string receiver, string port, string channel, string denom, uint256 amount, string memo) returns(bool success)
func (_IbcPrecompile *IbcPrecompileTransactor) TransferWithDefaultTimeout(opts *bind.TransactOpts, receiver string, port string, channel string, denom string, amount *big.Int, memo string) (*types.Transaction, error) {
	return _IbcPrecompile.contract.Transact(opts, "transferWithDefaultTimeout", receiver, port, channel, denom, amount, memo)
}

// TransferWithDefaultTimeout is a paid mutator transaction binding the contract method 0x97ba602b.
//
// Solidity: function transferWithDefaultTimeout(string receiver, string port, string channel, string denom, uint256 amount, string memo) returns(bool success)
func (_IbcPrecompile *IbcPrecompileSession) TransferWithDefaultTimeout(receiver string, port string, channel string, denom string, amount *big.Int, memo string) (*types.Transaction, error) {
	return _IbcPrecompile.Contract.TransferWithDefaultTimeout(&_IbcPrecompile.TransactOpts, receiver, port, channel, denom, amount, memo)
}

// TransferWithDefaultTimeout is a paid mutator transaction binding the contract method 0x97ba602b.
//
// Solidity: function transferWithDefaultTimeout(string receiver, string port, string channel, string denom, uint256 amount, string memo) returns(bool success)
func (_IbcPrecompile *IbcPrecompileTransactorSession) TransferWithDefaultTimeout(receiver string, port string, channel string, denom string, amount *big.Int, memo string) (*types.Transaction, error) {
	return _IbcPrecompile.Contract.TransferWithDefaultTimeout(&_IbcPrecompile.TransactOpts, receiver, port, channel, denom, amount, memo)
}

// IbcPrecompileTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the IbcPrecompile contract.
type IbcPrecompileTransferIterator struct {
	Event *IbcPrecompileTransfer // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IbcPrecompileTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IbcPrecompileTransfer)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IbcPrecompileTransfer)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IbcPrecompileTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IbcPrecompileTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IbcPrecompileTransfer represents a Transfer event raised by the IbcPrecompile contract.
type IbcPrecompileTransfer struct {
	Caller           common.Address
	Receiver         common.Hash
	Denom            common.Hash
	Port             string
	Channel          string
	Amount           *big.Int
	RevisionNumber   uint64
	RevisionHeight   uint64
	TimeoutTimestamp uint64
	Memo             string
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0x1eca6bbb7a9f439d9b6c781428d705de2d9021102a3ca129d8d54149fcf030e2.
//
// Solidity: event Transfer(address indexed caller, string indexed receiver, string indexed denom, string port, string channel, uint256 amount, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo)
func (_IbcPrecompile *IbcPrecompileFilterer) FilterTransfer(opts *bind.FilterOpts, caller []common.Address, receiver []string, denom []string) (*IbcPrecompileTransferIterator, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}
	var denomRule []interface{}
	for _, denomItem := range denom {
		denomRule = append(denomRule, denomItem)
	}

	logs, sub, err := _IbcPrecompile.contract.FilterLogs(opts, "Transfer", callerRule, receiverRule, denomRule)
	if err != nil {
		return nil, err
	}
	return &IbcPrecompileTransferIterator{contract: _IbcPrecompile.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0x1eca6bbb7a9f439d9b6c781428d705de2d9021102a3ca129d8d54149fcf030e2.
//
// Solidity: event Transfer(address indexed caller, string indexed receiver, string indexed denom, string port, string channel, uint256 amount, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo)
func (_IbcPrecompile *IbcPrecompileFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *IbcPrecompileTransfer, caller []common.Address, receiver []string, denom []string) (event.Subscription, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var receiverRule []interface{}
	for _, receiverItem := range receiver {
		receiverRule = append(receiverRule, receiverItem)
	}
	var denomRule []interface{}
	for _, denomItem := range denom {
		denomRule = append(denomRule, denomItem)
	}

	logs, sub, err := _IbcPrecompile.contract.WatchLogs(opts, "Transfer", callerRule, receiverRule, denomRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IbcPrecompileTransfer)
				if err := _IbcPrecompile.contract.UnpackLog(event, "Transfer", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransfer is a log parse operation binding the contract event 0x1eca6bbb7a9f439d9b6c781428d705de2d9021102a3ca129d8d54149fcf030e2.
//
// Solidity: event Transfer(address indexed caller, string indexed receiver, string indexed denom, string port, string channel, uint256 amount, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo)
func (_IbcPrecompile *IbcPrecompileFilterer) ParseTransfer(log types.Log) (*IbcPrecompileTransfer, error) {
	event := new(IbcPrecompileTransfer)
	if err := _IbcPrecompile.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
