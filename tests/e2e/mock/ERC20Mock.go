// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package mock

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

// ERC20MockMetaData contains all meta data concerning the ERC20Mock contract.
var ERC20MockMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"allowance\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"needed\",\"type\":\"uint256\"}],\"name\":\"ERC20InsufficientAllowance\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"balance\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"needed\",\"type\":\"uint256\"}],\"name\":\"ERC20InsufficientBalance\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"approver\",\"type\":\"address\"}],\"name\":\"ERC20InvalidApprover\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"ERC20InvalidReceiver\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"ERC20InvalidSender\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"ERC20InvalidSpender\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f5ffd5b506040518060400160405280600981526020016845524332304d6f636b60b81b815250604051806040016040528060048152602001634532304d60e01b815250816003908161005e919061010b565b50600461006b828261010b565b5050506101c5565b634e487b7160e01b5f52604160045260245ffd5b600181811c9082168061009b57607f821691505b6020821081036100b957634e487b7160e01b5f52602260045260245ffd5b50919050565b601f82111561010657805f5260205f20601f840160051c810160208510156100e45750805b601f840160051c820191505b81811015610103575f81556001016100f0565b50505b505050565b81516001600160401b0381111561012457610124610073565b610138816101328454610087565b846100bf565b6020601f82116001811461016a575f83156101535750848201515b5f19600385901b1c1916600184901b178455610103565b5f84815260208120601f198516915b828110156101995787850151825560209485019460019092019101610179565b50848210156101b657868401515f19600387901b60f8161c191681555b50505050600190811b01905550565b6107a2806101d25f395ff3fe608060405234801561000f575f5ffd5b50600436106100a6575f3560e01c806340c10f191161006e57806340c10f191461011f57806370a082311461013457806395d89b411461015c5780639dc29fac14610164578063a9059cbb14610177578063dd62ed3e1461018a575f5ffd5b806306fdde03146100aa578063095ea7b3146100c857806318160ddd146100eb57806323b872dd146100fd578063313ce56714610110575b5f5ffd5b6100b26101c2565b6040516100bf9190610612565b60405180910390f35b6100db6100d6366004610662565b610252565b60405190151581526020016100bf565b6002545b6040519081526020016100bf565b6100db61010b36600461068a565b61026b565b604051601281526020016100bf565b61013261012d366004610662565b61028e565b005b6100ef6101423660046106c4565b6001600160a01b03165f9081526020819052604090205490565b6100b261029c565b610132610172366004610662565b6102ab565b6100db610185366004610662565b6102b5565b6100ef6101983660046106e4565b6001600160a01b039182165f90815260016020908152604080832093909416825291909152205490565b6060600380546101d190610715565b80601f01602080910402602001604051908101604052809291908181526020018280546101fd90610715565b80156102485780601f1061021f57610100808354040283529160200191610248565b820191905f5260205f20905b81548152906001019060200180831161022b57829003601f168201915b5050505050905090565b5f3361025f8185856102c2565b60019150505b92915050565b5f336102788582856102d4565b610283858585610355565b506001949350505050565b61029882826103b2565b5050565b6060600480546101d190610715565b61029882826103e6565b5f3361025f818585610355565b6102cf838383600161041a565b505050565b6001600160a01b038381165f908152600160209081526040808320938616835292905220545f1981101561034f578181101561034157604051637dc7a0d960e11b81526001600160a01b038416600482015260248101829052604481018390526064015b60405180910390fd5b61034f84848484035f61041a565b50505050565b6001600160a01b03831661037e57604051634b637e8f60e11b81525f6004820152602401610338565b6001600160a01b0382166103a75760405163ec442f0560e01b81525f6004820152602401610338565b6102cf8383836104ec565b6001600160a01b0382166103db5760405163ec442f0560e01b81525f6004820152602401610338565b6102985f83836104ec565b6001600160a01b03821661040f57604051634b637e8f60e11b81525f6004820152602401610338565b610298825f836104ec565b6001600160a01b0384166104435760405163e602df0560e01b81525f6004820152602401610338565b6001600160a01b03831661046c57604051634a1406b160e11b81525f6004820152602401610338565b6001600160a01b038085165f908152600160209081526040808320938716835292905220829055801561034f57826001600160a01b0316846001600160a01b03167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925846040516104de91815260200190565b60405180910390a350505050565b6001600160a01b038316610516578060025f82825461050b919061074d565b909155506105869050565b6001600160a01b0383165f90815260208190526040902054818110156105685760405163391434e360e21b81526001600160a01b03851660048201526024810182905260448101839052606401610338565b6001600160a01b0384165f9081526020819052604090209082900390555b6001600160a01b0382166105a2576002805482900390556105c0565b6001600160a01b0382165f9081526020819052604090208054820190555b816001600160a01b0316836001600160a01b03167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef8360405161060591815260200190565b60405180910390a3505050565b602081525f82518060208401528060208501604085015e5f604082850101526040601f19601f83011684010191505092915050565b80356001600160a01b038116811461065d575f5ffd5b919050565b5f5f60408385031215610673575f5ffd5b61067c83610647565b946020939093013593505050565b5f5f5f6060848603121561069c575f5ffd5b6106a584610647565b92506106b360208501610647565b929592945050506040919091013590565b5f602082840312156106d4575f5ffd5b6106dd82610647565b9392505050565b5f5f604083850312156106f5575f5ffd5b6106fe83610647565b915061070c60208401610647565b90509250929050565b600181811c9082168061072957607f821691505b60208210810361074757634e487b7160e01b5f52602260045260245ffd5b50919050565b8082018082111561026557634e487b7160e01b5f52601160045260245ffdfea2646970667358221220d187705abdb3c91c3a2be0a014b5b1ca678c5683470a789ba4e7be48142b265464736f6c634300081d0033",
}

// ERC20MockABI is the input ABI used to generate the binding from.
// Deprecated: Use ERC20MockMetaData.ABI instead.
var ERC20MockABI = ERC20MockMetaData.ABI

// ERC20MockBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ERC20MockMetaData.Bin instead.
var ERC20MockBin = ERC20MockMetaData.Bin

// DeployERC20Mock deploys a new Ethereum contract, binding an instance of ERC20Mock to it.
func DeployERC20Mock(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ERC20Mock, error) {
	parsed, err := ERC20MockMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ERC20MockBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ERC20Mock{ERC20MockCaller: ERC20MockCaller{contract: contract}, ERC20MockTransactor: ERC20MockTransactor{contract: contract}, ERC20MockFilterer: ERC20MockFilterer{contract: contract}}, nil
}

// ERC20Mock is an auto generated Go binding around an Ethereum contract.
type ERC20Mock struct {
	ERC20MockCaller     // Read-only binding to the contract
	ERC20MockTransactor // Write-only binding to the contract
	ERC20MockFilterer   // Log filterer for contract events
}

// ERC20MockCaller is an auto generated read-only Go binding around an Ethereum contract.
type ERC20MockCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20MockTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ERC20MockTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20MockFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ERC20MockFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20MockSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ERC20MockSession struct {
	Contract     *ERC20Mock        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ERC20MockCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ERC20MockCallerSession struct {
	Contract *ERC20MockCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// ERC20MockTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ERC20MockTransactorSession struct {
	Contract     *ERC20MockTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// ERC20MockRaw is an auto generated low-level Go binding around an Ethereum contract.
type ERC20MockRaw struct {
	Contract *ERC20Mock // Generic contract binding to access the raw methods on
}

// ERC20MockCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ERC20MockCallerRaw struct {
	Contract *ERC20MockCaller // Generic read-only contract binding to access the raw methods on
}

// ERC20MockTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ERC20MockTransactorRaw struct {
	Contract *ERC20MockTransactor // Generic write-only contract binding to access the raw methods on
}

// NewERC20Mock creates a new instance of ERC20Mock, bound to a specific deployed contract.
func NewERC20Mock(address common.Address, backend bind.ContractBackend) (*ERC20Mock, error) {
	contract, err := bindERC20Mock(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ERC20Mock{ERC20MockCaller: ERC20MockCaller{contract: contract}, ERC20MockTransactor: ERC20MockTransactor{contract: contract}, ERC20MockFilterer: ERC20MockFilterer{contract: contract}}, nil
}

// NewERC20MockCaller creates a new read-only instance of ERC20Mock, bound to a specific deployed contract.
func NewERC20MockCaller(address common.Address, caller bind.ContractCaller) (*ERC20MockCaller, error) {
	contract, err := bindERC20Mock(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20MockCaller{contract: contract}, nil
}

// NewERC20MockTransactor creates a new write-only instance of ERC20Mock, bound to a specific deployed contract.
func NewERC20MockTransactor(address common.Address, transactor bind.ContractTransactor) (*ERC20MockTransactor, error) {
	contract, err := bindERC20Mock(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20MockTransactor{contract: contract}, nil
}

// NewERC20MockFilterer creates a new log filterer instance of ERC20Mock, bound to a specific deployed contract.
func NewERC20MockFilterer(address common.Address, filterer bind.ContractFilterer) (*ERC20MockFilterer, error) {
	contract, err := bindERC20Mock(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ERC20MockFilterer{contract: contract}, nil
}

// bindERC20Mock binds a generic wrapper to an already deployed contract.
func bindERC20Mock(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ERC20MockMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20Mock *ERC20MockRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ERC20Mock.Contract.ERC20MockCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20Mock *ERC20MockRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20Mock.Contract.ERC20MockTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20Mock *ERC20MockRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20Mock.Contract.ERC20MockTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20Mock *ERC20MockCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ERC20Mock.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20Mock *ERC20MockTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20Mock.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20Mock *ERC20MockTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20Mock.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_ERC20Mock *ERC20MockCaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ERC20Mock.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_ERC20Mock *ERC20MockSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _ERC20Mock.Contract.Allowance(&_ERC20Mock.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_ERC20Mock *ERC20MockCallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _ERC20Mock.Contract.Allowance(&_ERC20Mock.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_ERC20Mock *ERC20MockCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ERC20Mock.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_ERC20Mock *ERC20MockSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _ERC20Mock.Contract.BalanceOf(&_ERC20Mock.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_ERC20Mock *ERC20MockCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _ERC20Mock.Contract.BalanceOf(&_ERC20Mock.CallOpts, account)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ERC20Mock *ERC20MockCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ERC20Mock.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ERC20Mock *ERC20MockSession) Decimals() (uint8, error) {
	return _ERC20Mock.Contract.Decimals(&_ERC20Mock.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ERC20Mock *ERC20MockCallerSession) Decimals() (uint8, error) {
	return _ERC20Mock.Contract.Decimals(&_ERC20Mock.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ERC20Mock *ERC20MockCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _ERC20Mock.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ERC20Mock *ERC20MockSession) Name() (string, error) {
	return _ERC20Mock.Contract.Name(&_ERC20Mock.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ERC20Mock *ERC20MockCallerSession) Name() (string, error) {
	return _ERC20Mock.Contract.Name(&_ERC20Mock.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ERC20Mock *ERC20MockCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _ERC20Mock.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ERC20Mock *ERC20MockSession) Symbol() (string, error) {
	return _ERC20Mock.Contract.Symbol(&_ERC20Mock.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ERC20Mock *ERC20MockCallerSession) Symbol() (string, error) {
	return _ERC20Mock.Contract.Symbol(&_ERC20Mock.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ERC20Mock *ERC20MockCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ERC20Mock.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ERC20Mock *ERC20MockSession) TotalSupply() (*big.Int, error) {
	return _ERC20Mock.Contract.TotalSupply(&_ERC20Mock.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ERC20Mock *ERC20MockCallerSession) TotalSupply() (*big.Int, error) {
	return _ERC20Mock.Contract.TotalSupply(&_ERC20Mock.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_ERC20Mock *ERC20MockTransactor) Approve(opts *bind.TransactOpts, spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.contract.Transact(opts, "approve", spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_ERC20Mock *ERC20MockSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.Contract.Approve(&_ERC20Mock.TransactOpts, spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_ERC20Mock *ERC20MockTransactorSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.Contract.Approve(&_ERC20Mock.TransactOpts, spender, value)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address account, uint256 amount) returns()
func (_ERC20Mock *ERC20MockTransactor) Burn(opts *bind.TransactOpts, account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.contract.Transact(opts, "burn", account, amount)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address account, uint256 amount) returns()
func (_ERC20Mock *ERC20MockSession) Burn(account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.Contract.Burn(&_ERC20Mock.TransactOpts, account, amount)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address account, uint256 amount) returns()
func (_ERC20Mock *ERC20MockTransactorSession) Burn(account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.Contract.Burn(&_ERC20Mock.TransactOpts, account, amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address account, uint256 amount) returns()
func (_ERC20Mock *ERC20MockTransactor) Mint(opts *bind.TransactOpts, account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.contract.Transact(opts, "mint", account, amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address account, uint256 amount) returns()
func (_ERC20Mock *ERC20MockSession) Mint(account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.Contract.Mint(&_ERC20Mock.TransactOpts, account, amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address account, uint256 amount) returns()
func (_ERC20Mock *ERC20MockTransactorSession) Mint(account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.Contract.Mint(&_ERC20Mock.TransactOpts, account, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_ERC20Mock *ERC20MockTransactor) Transfer(opts *bind.TransactOpts, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.contract.Transact(opts, "transfer", to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_ERC20Mock *ERC20MockSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.Contract.Transfer(&_ERC20Mock.TransactOpts, to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_ERC20Mock *ERC20MockTransactorSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.Contract.Transfer(&_ERC20Mock.TransactOpts, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_ERC20Mock *ERC20MockTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.contract.Transact(opts, "transferFrom", from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_ERC20Mock *ERC20MockSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.Contract.TransferFrom(&_ERC20Mock.TransactOpts, from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_ERC20Mock *ERC20MockTransactorSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _ERC20Mock.Contract.TransferFrom(&_ERC20Mock.TransactOpts, from, to, value)
}

// ERC20MockApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the ERC20Mock contract.
type ERC20MockApprovalIterator struct {
	Event *ERC20MockApproval // Event containing the contract specifics and raw log

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
func (it *ERC20MockApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20MockApproval)
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
		it.Event = new(ERC20MockApproval)
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
func (it *ERC20MockApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20MockApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20MockApproval represents a Approval event raised by the ERC20Mock contract.
type ERC20MockApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ERC20Mock *ERC20MockFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*ERC20MockApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ERC20Mock.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &ERC20MockApprovalIterator{contract: _ERC20Mock.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ERC20Mock *ERC20MockFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ERC20MockApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ERC20Mock.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20MockApproval)
				if err := _ERC20Mock.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ERC20Mock *ERC20MockFilterer) ParseApproval(log types.Log) (*ERC20MockApproval, error) {
	event := new(ERC20MockApproval)
	if err := _ERC20Mock.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ERC20MockTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the ERC20Mock contract.
type ERC20MockTransferIterator struct {
	Event *ERC20MockTransfer // Event containing the contract specifics and raw log

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
func (it *ERC20MockTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20MockTransfer)
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
		it.Event = new(ERC20MockTransfer)
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
func (it *ERC20MockTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20MockTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20MockTransfer represents a Transfer event raised by the ERC20Mock contract.
type ERC20MockTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ERC20Mock *ERC20MockFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*ERC20MockTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ERC20Mock.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &ERC20MockTransferIterator{contract: _ERC20Mock.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ERC20Mock *ERC20MockFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ERC20MockTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ERC20Mock.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20MockTransfer)
				if err := _ERC20Mock.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ERC20Mock *ERC20MockFilterer) ParseTransfer(log types.Log) (*ERC20MockTransfer, error) {
	event := new(ERC20MockTransfer)
	if err := _ERC20Mock.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
