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

// Coin is an auto generated low-level Go binding around an user-defined struct.
type Coin struct {
	Denom  string
	Amount *big.Int
}

// WasmdPrecompileMetaData contains all meta data concerning the WasmdPrecompile contract.
var WasmdPrecompileMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"contractAddress\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"ContractExecuted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"codeID\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"contractAddress\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"ContractInstantiated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"contractAddress\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"msg\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"denom\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structCoin[]\",\"name\":\"coins\",\"type\":\"tuple[]\"}],\"name\":\"execute\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"admin\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"codeID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"label\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"msg\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"denom\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structCoin[]\",\"name\":\"coins\",\"type\":\"tuple[]\"}],\"name\":\"instantiate\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"contractAddress\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"queryData\",\"type\":\"bytes\"}],\"name\":\"queryRaw\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"contractAddress\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"msg\",\"type\":\"bytes\"}],\"name\":\"querySmart\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// WasmdPrecompileABI is the input ABI used to generate the binding from.
// Deprecated: Use WasmdPrecompileMetaData.ABI instead.
var WasmdPrecompileABI = WasmdPrecompileMetaData.ABI

// WasmdPrecompile is an auto generated Go binding around an Ethereum contract.
type WasmdPrecompile struct {
	WasmdPrecompileCaller     // Read-only binding to the contract
	WasmdPrecompileTransactor // Write-only binding to the contract
	WasmdPrecompileFilterer   // Log filterer for contract events
}

// WasmdPrecompileCaller is an auto generated read-only Go binding around an Ethereum contract.
type WasmdPrecompileCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WasmdPrecompileTransactor is an auto generated write-only Go binding around an Ethereum contract.
type WasmdPrecompileTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WasmdPrecompileFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type WasmdPrecompileFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WasmdPrecompileSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type WasmdPrecompileSession struct {
	Contract     *WasmdPrecompile  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// WasmdPrecompileCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type WasmdPrecompileCallerSession struct {
	Contract *WasmdPrecompileCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// WasmdPrecompileTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type WasmdPrecompileTransactorSession struct {
	Contract     *WasmdPrecompileTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// WasmdPrecompileRaw is an auto generated low-level Go binding around an Ethereum contract.
type WasmdPrecompileRaw struct {
	Contract *WasmdPrecompile // Generic contract binding to access the raw methods on
}

// WasmdPrecompileCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type WasmdPrecompileCallerRaw struct {
	Contract *WasmdPrecompileCaller // Generic read-only contract binding to access the raw methods on
}

// WasmdPrecompileTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type WasmdPrecompileTransactorRaw struct {
	Contract *WasmdPrecompileTransactor // Generic write-only contract binding to access the raw methods on
}

// NewWasmdPrecompile creates a new instance of WasmdPrecompile, bound to a specific deployed contract.
func NewWasmdPrecompile(address common.Address, backend bind.ContractBackend) (*WasmdPrecompile, error) {
	contract, err := bindWasmdPrecompile(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &WasmdPrecompile{WasmdPrecompileCaller: WasmdPrecompileCaller{contract: contract}, WasmdPrecompileTransactor: WasmdPrecompileTransactor{contract: contract}, WasmdPrecompileFilterer: WasmdPrecompileFilterer{contract: contract}}, nil
}

// NewWasmdPrecompileCaller creates a new read-only instance of WasmdPrecompile, bound to a specific deployed contract.
func NewWasmdPrecompileCaller(address common.Address, caller bind.ContractCaller) (*WasmdPrecompileCaller, error) {
	contract, err := bindWasmdPrecompile(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &WasmdPrecompileCaller{contract: contract}, nil
}

// NewWasmdPrecompileTransactor creates a new write-only instance of WasmdPrecompile, bound to a specific deployed contract.
func NewWasmdPrecompileTransactor(address common.Address, transactor bind.ContractTransactor) (*WasmdPrecompileTransactor, error) {
	contract, err := bindWasmdPrecompile(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &WasmdPrecompileTransactor{contract: contract}, nil
}

// NewWasmdPrecompileFilterer creates a new log filterer instance of WasmdPrecompile, bound to a specific deployed contract.
func NewWasmdPrecompileFilterer(address common.Address, filterer bind.ContractFilterer) (*WasmdPrecompileFilterer, error) {
	contract, err := bindWasmdPrecompile(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &WasmdPrecompileFilterer{contract: contract}, nil
}

// bindWasmdPrecompile binds a generic wrapper to an already deployed contract.
func bindWasmdPrecompile(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := WasmdPrecompileMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_WasmdPrecompile *WasmdPrecompileRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _WasmdPrecompile.Contract.WasmdPrecompileCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_WasmdPrecompile *WasmdPrecompileRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WasmdPrecompile.Contract.WasmdPrecompileTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_WasmdPrecompile *WasmdPrecompileRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _WasmdPrecompile.Contract.WasmdPrecompileTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_WasmdPrecompile *WasmdPrecompileCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _WasmdPrecompile.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_WasmdPrecompile *WasmdPrecompileTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WasmdPrecompile.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_WasmdPrecompile *WasmdPrecompileTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _WasmdPrecompile.Contract.contract.Transact(opts, method, params...)
}

// QueryRaw is a free data retrieval call binding the contract method 0x506e35d0.
//
// Solidity: function queryRaw(string contractAddress, bytes queryData) view returns(bytes data)
func (_WasmdPrecompile *WasmdPrecompileCaller) QueryRaw(opts *bind.CallOpts, contractAddress string, queryData []byte) ([]byte, error) {
	var out []interface{}
	err := _WasmdPrecompile.contract.Call(opts, &out, "queryRaw", contractAddress, queryData)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// QueryRaw is a free data retrieval call binding the contract method 0x506e35d0.
//
// Solidity: function queryRaw(string contractAddress, bytes queryData) view returns(bytes data)
func (_WasmdPrecompile *WasmdPrecompileSession) QueryRaw(contractAddress string, queryData []byte) ([]byte, error) {
	return _WasmdPrecompile.Contract.QueryRaw(&_WasmdPrecompile.CallOpts, contractAddress, queryData)
}

// QueryRaw is a free data retrieval call binding the contract method 0x506e35d0.
//
// Solidity: function queryRaw(string contractAddress, bytes queryData) view returns(bytes data)
func (_WasmdPrecompile *WasmdPrecompileCallerSession) QueryRaw(contractAddress string, queryData []byte) ([]byte, error) {
	return _WasmdPrecompile.Contract.QueryRaw(&_WasmdPrecompile.CallOpts, contractAddress, queryData)
}

// QuerySmart is a free data retrieval call binding the contract method 0xaded76b6.
//
// Solidity: function querySmart(string contractAddress, bytes msg) view returns(bytes data)
func (_WasmdPrecompile *WasmdPrecompileCaller) QuerySmart(opts *bind.CallOpts, contractAddress string, msg []byte) ([]byte, error) {
	var out []interface{}
	err := _WasmdPrecompile.contract.Call(opts, &out, "querySmart", contractAddress, msg)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// QuerySmart is a free data retrieval call binding the contract method 0xaded76b6.
//
// Solidity: function querySmart(string contractAddress, bytes msg) view returns(bytes data)
func (_WasmdPrecompile *WasmdPrecompileSession) QuerySmart(contractAddress string, msg []byte) ([]byte, error) {
	return _WasmdPrecompile.Contract.QuerySmart(&_WasmdPrecompile.CallOpts, contractAddress, msg)
}

// QuerySmart is a free data retrieval call binding the contract method 0xaded76b6.
//
// Solidity: function querySmart(string contractAddress, bytes msg) view returns(bytes data)
func (_WasmdPrecompile *WasmdPrecompileCallerSession) QuerySmart(contractAddress string, msg []byte) ([]byte, error) {
	return _WasmdPrecompile.Contract.QuerySmart(&_WasmdPrecompile.CallOpts, contractAddress, msg)
}

// Execute is a paid mutator transaction binding the contract method 0x61ffaee4.
//
// Solidity: function execute(string contractAddress, bytes msg, (string,uint256)[] coins) returns(bool success)
func (_WasmdPrecompile *WasmdPrecompileTransactor) Execute(opts *bind.TransactOpts, contractAddress string, msg []byte, coins []Coin) (*types.Transaction, error) {
	return _WasmdPrecompile.contract.Transact(opts, "execute", contractAddress, msg, coins)
}

// Execute is a paid mutator transaction binding the contract method 0x61ffaee4.
//
// Solidity: function execute(string contractAddress, bytes msg, (string,uint256)[] coins) returns(bool success)
func (_WasmdPrecompile *WasmdPrecompileSession) Execute(contractAddress string, msg []byte, coins []Coin) (*types.Transaction, error) {
	return _WasmdPrecompile.Contract.Execute(&_WasmdPrecompile.TransactOpts, contractAddress, msg, coins)
}

// Execute is a paid mutator transaction binding the contract method 0x61ffaee4.
//
// Solidity: function execute(string contractAddress, bytes msg, (string,uint256)[] coins) returns(bool success)
func (_WasmdPrecompile *WasmdPrecompileTransactorSession) Execute(contractAddress string, msg []byte, coins []Coin) (*types.Transaction, error) {
	return _WasmdPrecompile.Contract.Execute(&_WasmdPrecompile.TransactOpts, contractAddress, msg, coins)
}

// Instantiate is a paid mutator transaction binding the contract method 0x3fd60967.
//
// Solidity: function instantiate(address admin, uint64 codeID, string label, bytes msg, (string,uint256)[] coins) returns(bool success)
func (_WasmdPrecompile *WasmdPrecompileTransactor) Instantiate(opts *bind.TransactOpts, admin common.Address, codeID uint64, label string, msg []byte, coins []Coin) (*types.Transaction, error) {
	return _WasmdPrecompile.contract.Transact(opts, "instantiate", admin, codeID, label, msg, coins)
}

// Instantiate is a paid mutator transaction binding the contract method 0x3fd60967.
//
// Solidity: function instantiate(address admin, uint64 codeID, string label, bytes msg, (string,uint256)[] coins) returns(bool success)
func (_WasmdPrecompile *WasmdPrecompileSession) Instantiate(admin common.Address, codeID uint64, label string, msg []byte, coins []Coin) (*types.Transaction, error) {
	return _WasmdPrecompile.Contract.Instantiate(&_WasmdPrecompile.TransactOpts, admin, codeID, label, msg, coins)
}

// Instantiate is a paid mutator transaction binding the contract method 0x3fd60967.
//
// Solidity: function instantiate(address admin, uint64 codeID, string label, bytes msg, (string,uint256)[] coins) returns(bool success)
func (_WasmdPrecompile *WasmdPrecompileTransactorSession) Instantiate(admin common.Address, codeID uint64, label string, msg []byte, coins []Coin) (*types.Transaction, error) {
	return _WasmdPrecompile.Contract.Instantiate(&_WasmdPrecompile.TransactOpts, admin, codeID, label, msg, coins)
}

// WasmdPrecompileContractExecutedIterator is returned from FilterContractExecuted and is used to iterate over the raw logs and unpacked data for ContractExecuted events raised by the WasmdPrecompile contract.
type WasmdPrecompileContractExecutedIterator struct {
	Event *WasmdPrecompileContractExecuted // Event containing the contract specifics and raw log

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
func (it *WasmdPrecompileContractExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WasmdPrecompileContractExecuted)
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
		it.Event = new(WasmdPrecompileContractExecuted)
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
func (it *WasmdPrecompileContractExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WasmdPrecompileContractExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WasmdPrecompileContractExecuted represents a ContractExecuted event raised by the WasmdPrecompile contract.
type WasmdPrecompileContractExecuted struct {
	ContractAddress common.Hash
	Caller          common.Address
	Data            []byte
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterContractExecuted is a free log retrieval operation binding the contract event 0x74eed1284ef48754e4f0a43308ca1f693db2da7a2e057b9ed70fe95bc4432e19.
//
// Solidity: event ContractExecuted(string indexed contractAddress, address indexed caller, bytes data)
func (_WasmdPrecompile *WasmdPrecompileFilterer) FilterContractExecuted(opts *bind.FilterOpts, contractAddress []string, caller []common.Address) (*WasmdPrecompileContractExecutedIterator, error) {

	var contractAddressRule []interface{}
	for _, contractAddressItem := range contractAddress {
		contractAddressRule = append(contractAddressRule, contractAddressItem)
	}
	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}

	logs, sub, err := _WasmdPrecompile.contract.FilterLogs(opts, "ContractExecuted", contractAddressRule, callerRule)
	if err != nil {
		return nil, err
	}
	return &WasmdPrecompileContractExecutedIterator{contract: _WasmdPrecompile.contract, event: "ContractExecuted", logs: logs, sub: sub}, nil
}

// WatchContractExecuted is a free log subscription operation binding the contract event 0x74eed1284ef48754e4f0a43308ca1f693db2da7a2e057b9ed70fe95bc4432e19.
//
// Solidity: event ContractExecuted(string indexed contractAddress, address indexed caller, bytes data)
func (_WasmdPrecompile *WasmdPrecompileFilterer) WatchContractExecuted(opts *bind.WatchOpts, sink chan<- *WasmdPrecompileContractExecuted, contractAddress []string, caller []common.Address) (event.Subscription, error) {

	var contractAddressRule []interface{}
	for _, contractAddressItem := range contractAddress {
		contractAddressRule = append(contractAddressRule, contractAddressItem)
	}
	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}

	logs, sub, err := _WasmdPrecompile.contract.WatchLogs(opts, "ContractExecuted", contractAddressRule, callerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WasmdPrecompileContractExecuted)
				if err := _WasmdPrecompile.contract.UnpackLog(event, "ContractExecuted", log); err != nil {
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

// ParseContractExecuted is a log parse operation binding the contract event 0x74eed1284ef48754e4f0a43308ca1f693db2da7a2e057b9ed70fe95bc4432e19.
//
// Solidity: event ContractExecuted(string indexed contractAddress, address indexed caller, bytes data)
func (_WasmdPrecompile *WasmdPrecompileFilterer) ParseContractExecuted(log types.Log) (*WasmdPrecompileContractExecuted, error) {
	event := new(WasmdPrecompileContractExecuted)
	if err := _WasmdPrecompile.contract.UnpackLog(event, "ContractExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WasmdPrecompileContractInstantiatedIterator is returned from FilterContractInstantiated and is used to iterate over the raw logs and unpacked data for ContractInstantiated events raised by the WasmdPrecompile contract.
type WasmdPrecompileContractInstantiatedIterator struct {
	Event *WasmdPrecompileContractInstantiated // Event containing the contract specifics and raw log

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
func (it *WasmdPrecompileContractInstantiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WasmdPrecompileContractInstantiated)
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
		it.Event = new(WasmdPrecompileContractInstantiated)
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
func (it *WasmdPrecompileContractInstantiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WasmdPrecompileContractInstantiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WasmdPrecompileContractInstantiated represents a ContractInstantiated event raised by the WasmdPrecompile contract.
type WasmdPrecompileContractInstantiated struct {
	Caller          common.Address
	CodeID          uint64
	ContractAddress string
	Data            []byte
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterContractInstantiated is a free log retrieval operation binding the contract event 0x512f93783417953ca4442aa5f46656d4a4066f5680cc06b8ecde58b325c5380e.
//
// Solidity: event ContractInstantiated(address indexed caller, uint64 indexed codeID, string contractAddress, bytes data)
func (_WasmdPrecompile *WasmdPrecompileFilterer) FilterContractInstantiated(opts *bind.FilterOpts, caller []common.Address, codeID []uint64) (*WasmdPrecompileContractInstantiatedIterator, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var codeIDRule []interface{}
	for _, codeIDItem := range codeID {
		codeIDRule = append(codeIDRule, codeIDItem)
	}

	logs, sub, err := _WasmdPrecompile.contract.FilterLogs(opts, "ContractInstantiated", callerRule, codeIDRule)
	if err != nil {
		return nil, err
	}
	return &WasmdPrecompileContractInstantiatedIterator{contract: _WasmdPrecompile.contract, event: "ContractInstantiated", logs: logs, sub: sub}, nil
}

// WatchContractInstantiated is a free log subscription operation binding the contract event 0x512f93783417953ca4442aa5f46656d4a4066f5680cc06b8ecde58b325c5380e.
//
// Solidity: event ContractInstantiated(address indexed caller, uint64 indexed codeID, string contractAddress, bytes data)
func (_WasmdPrecompile *WasmdPrecompileFilterer) WatchContractInstantiated(opts *bind.WatchOpts, sink chan<- *WasmdPrecompileContractInstantiated, caller []common.Address, codeID []uint64) (event.Subscription, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}
	var codeIDRule []interface{}
	for _, codeIDItem := range codeID {
		codeIDRule = append(codeIDRule, codeIDItem)
	}

	logs, sub, err := _WasmdPrecompile.contract.WatchLogs(opts, "ContractInstantiated", callerRule, codeIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WasmdPrecompileContractInstantiated)
				if err := _WasmdPrecompile.contract.UnpackLog(event, "ContractInstantiated", log); err != nil {
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

// ParseContractInstantiated is a log parse operation binding the contract event 0x512f93783417953ca4442aa5f46656d4a4066f5680cc06b8ecde58b325c5380e.
//
// Solidity: event ContractInstantiated(address indexed caller, uint64 indexed codeID, string contractAddress, bytes data)
func (_WasmdPrecompile *WasmdPrecompileFilterer) ParseContractInstantiated(log types.Log) (*WasmdPrecompileContractInstantiated, error) {
	event := new(WasmdPrecompileContractInstantiated)
	if err := _WasmdPrecompile.contract.UnpackLog(event, "ContractInstantiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
