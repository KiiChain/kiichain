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

// PrecompilesMetaData contains all meta data concerning the Precompiles contract.
var PrecompilesMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"denom\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"port\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"channel\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"revisionNumber\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"revisionHeight\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"timeoutTimestamp\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"memo\",\"type\":\"string\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"port\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"channel\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"denom\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint64\",\"name\":\"revisionNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"revisionHeight\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"timeoutTimestamp\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"memo\",\"type\":\"string\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"port\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"channel\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"denom\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"memo\",\"type\":\"string\"}],\"name\":\"transferWithDefaultTimeout\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// PrecompilesABI is the input ABI used to generate the binding from.
// Deprecated: Use PrecompilesMetaData.ABI instead.
var PrecompilesABI = PrecompilesMetaData.ABI

// Precompiles is an auto generated Go binding around an Ethereum contract.
type Precompiles struct {
	PrecompilesCaller     // Read-only binding to the contract
	PrecompilesTransactor // Write-only binding to the contract
	PrecompilesFilterer   // Log filterer for contract events
}

// PrecompilesCaller is an auto generated read-only Go binding around an Ethereum contract.
type PrecompilesCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PrecompilesTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PrecompilesTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PrecompilesFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PrecompilesFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PrecompilesSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PrecompilesSession struct {
	Contract     *Precompiles      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PrecompilesCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PrecompilesCallerSession struct {
	Contract *PrecompilesCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// PrecompilesTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PrecompilesTransactorSession struct {
	Contract     *PrecompilesTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// PrecompilesRaw is an auto generated low-level Go binding around an Ethereum contract.
type PrecompilesRaw struct {
	Contract *Precompiles // Generic contract binding to access the raw methods on
}

// PrecompilesCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PrecompilesCallerRaw struct {
	Contract *PrecompilesCaller // Generic read-only contract binding to access the raw methods on
}

// PrecompilesTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PrecompilesTransactorRaw struct {
	Contract *PrecompilesTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPrecompiles creates a new instance of Precompiles, bound to a specific deployed contract.
func NewPrecompiles(address common.Address, backend bind.ContractBackend) (*Precompiles, error) {
	contract, err := bindPrecompiles(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Precompiles{PrecompilesCaller: PrecompilesCaller{contract: contract}, PrecompilesTransactor: PrecompilesTransactor{contract: contract}, PrecompilesFilterer: PrecompilesFilterer{contract: contract}}, nil
}

// NewPrecompilesCaller creates a new read-only instance of Precompiles, bound to a specific deployed contract.
func NewPrecompilesCaller(address common.Address, caller bind.ContractCaller) (*PrecompilesCaller, error) {
	contract, err := bindPrecompiles(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PrecompilesCaller{contract: contract}, nil
}

// NewPrecompilesTransactor creates a new write-only instance of Precompiles, bound to a specific deployed contract.
func NewPrecompilesTransactor(address common.Address, transactor bind.ContractTransactor) (*PrecompilesTransactor, error) {
	contract, err := bindPrecompiles(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PrecompilesTransactor{contract: contract}, nil
}

// NewPrecompilesFilterer creates a new log filterer instance of Precompiles, bound to a specific deployed contract.
func NewPrecompilesFilterer(address common.Address, filterer bind.ContractFilterer) (*PrecompilesFilterer, error) {
	contract, err := bindPrecompiles(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PrecompilesFilterer{contract: contract}, nil
}

// bindPrecompiles binds a generic wrapper to an already deployed contract.
func bindPrecompiles(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PrecompilesMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Precompiles *PrecompilesRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Precompiles.Contract.PrecompilesCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Precompiles *PrecompilesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Precompiles.Contract.PrecompilesTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Precompiles *PrecompilesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Precompiles.Contract.PrecompilesTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Precompiles *PrecompilesCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Precompiles.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Precompiles *PrecompilesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Precompiles.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Precompiles *PrecompilesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Precompiles.Contract.contract.Transact(opts, method, params...)
}

// Transfer is a paid mutator transaction binding the contract method 0x98a955c7.
//
// Solidity: function transfer(string receiver, string port, string channel, string denom, uint256 amount, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo) returns(bool success)
func (_Precompiles *PrecompilesTransactor) Transfer(opts *bind.TransactOpts, receiver string, port string, channel string, denom string, amount *big.Int, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string) (*types.Transaction, error) {
	return _Precompiles.contract.Transact(opts, "transfer", receiver, port, channel, denom, amount, revisionNumber, revisionHeight, timeoutTimestamp, memo)
}

// Transfer is a paid mutator transaction binding the contract method 0x98a955c7.
//
// Solidity: function transfer(string receiver, string port, string channel, string denom, uint256 amount, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo) returns(bool success)
func (_Precompiles *PrecompilesSession) Transfer(receiver string, port string, channel string, denom string, amount *big.Int, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string) (*types.Transaction, error) {
	return _Precompiles.Contract.Transfer(&_Precompiles.TransactOpts, receiver, port, channel, denom, amount, revisionNumber, revisionHeight, timeoutTimestamp, memo)
}

// Transfer is a paid mutator transaction binding the contract method 0x98a955c7.
//
// Solidity: function transfer(string receiver, string port, string channel, string denom, uint256 amount, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo) returns(bool success)
func (_Precompiles *PrecompilesTransactorSession) Transfer(receiver string, port string, channel string, denom string, amount *big.Int, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string) (*types.Transaction, error) {
	return _Precompiles.Contract.Transfer(&_Precompiles.TransactOpts, receiver, port, channel, denom, amount, revisionNumber, revisionHeight, timeoutTimestamp, memo)
}

// TransferWithDefaultTimeout is a paid mutator transaction binding the contract method 0x97ba602b.
//
// Solidity: function transferWithDefaultTimeout(string receiver, string port, string channel, string denom, uint256 amount, string memo) returns(bool success)
func (_Precompiles *PrecompilesTransactor) TransferWithDefaultTimeout(opts *bind.TransactOpts, receiver string, port string, channel string, denom string, amount *big.Int, memo string) (*types.Transaction, error) {
	return _Precompiles.contract.Transact(opts, "transferWithDefaultTimeout", receiver, port, channel, denom, amount, memo)
}

// TransferWithDefaultTimeout is a paid mutator transaction binding the contract method 0x97ba602b.
//
// Solidity: function transferWithDefaultTimeout(string receiver, string port, string channel, string denom, uint256 amount, string memo) returns(bool success)
func (_Precompiles *PrecompilesSession) TransferWithDefaultTimeout(receiver string, port string, channel string, denom string, amount *big.Int, memo string) (*types.Transaction, error) {
	return _Precompiles.Contract.TransferWithDefaultTimeout(&_Precompiles.TransactOpts, receiver, port, channel, denom, amount, memo)
}

// TransferWithDefaultTimeout is a paid mutator transaction binding the contract method 0x97ba602b.
//
// Solidity: function transferWithDefaultTimeout(string receiver, string port, string channel, string denom, uint256 amount, string memo) returns(bool success)
func (_Precompiles *PrecompilesTransactorSession) TransferWithDefaultTimeout(receiver string, port string, channel string, denom string, amount *big.Int, memo string) (*types.Transaction, error) {
	return _Precompiles.Contract.TransferWithDefaultTimeout(&_Precompiles.TransactOpts, receiver, port, channel, denom, amount, memo)
}

// PrecompilesTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the Precompiles contract.
type PrecompilesTransferIterator struct {
	Event *PrecompilesTransfer // Event containing the contract specifics and raw log

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
func (it *PrecompilesTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PrecompilesTransfer)
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
		it.Event = new(PrecompilesTransfer)
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
func (it *PrecompilesTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PrecompilesTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PrecompilesTransfer represents a Transfer event raised by the Precompiles contract.
type PrecompilesTransfer struct {
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
func (_Precompiles *PrecompilesFilterer) FilterTransfer(opts *bind.FilterOpts, caller []common.Address, receiver []string, denom []string) (*PrecompilesTransferIterator, error) {

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

	logs, sub, err := _Precompiles.contract.FilterLogs(opts, "Transfer", callerRule, receiverRule, denomRule)
	if err != nil {
		return nil, err
	}
	return &PrecompilesTransferIterator{contract: _Precompiles.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0x1eca6bbb7a9f439d9b6c781428d705de2d9021102a3ca129d8d54149fcf030e2.
//
// Solidity: event Transfer(address indexed caller, string indexed receiver, string indexed denom, string port, string channel, uint256 amount, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo)
func (_Precompiles *PrecompilesFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *PrecompilesTransfer, caller []common.Address, receiver []string, denom []string) (event.Subscription, error) {

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

	logs, sub, err := _Precompiles.contract.WatchLogs(opts, "Transfer", callerRule, receiverRule, denomRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PrecompilesTransfer)
				if err := _Precompiles.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_Precompiles *PrecompilesFilterer) ParseTransfer(log types.Log) (*PrecompilesTransfer, error) {
	event := new(PrecompilesTransfer)
	if err := _Precompiles.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
