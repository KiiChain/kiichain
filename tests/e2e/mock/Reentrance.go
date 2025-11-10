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

// ReentranceMetaData contains all meta data concerning the Reentrance contract.
var ReentranceMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"gasToConsume\",\"type\":\"uint256\"}],\"name\":\"ping\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"wasmBech32\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"}],\"name\":\"reentranceQuery\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"wasmBech32\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"iterations\",\"type\":\"uint256\"}],\"name\":\"spin\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"n\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x6080604052348015600e575f5ffd5b506108488061001c5f395ff3fe608060405234801561000f575f5ffd5b506004361061003f575f3560e01c8063773acdef1461004357806385530c7414610073578063a7c826f7146100a3575b5f5ffd5b61005d60048036038101906100589190610298565b6100d3565b60405161006a91906102d2565b60405180910390f35b61008d600480360381019061008891906104c5565b610119565b60405161009a91906102d2565b60405180910390f35b6100bd60048036038101906100b8919061054d565b6101cb565b6040516100ca9190610623565b60405180910390f35b5f5f5a90505f8382116100e6575f6100f3565b83826100f29190610670565b5b90505f5f90505b815a111561010d576001810190506100fa565b602a9350505050919050565b5f5f5f90505b828110156101c35761100173ffffffffffffffffffffffffffffffffffffffff1663aded76b686866040518363ffffffff1660e01b81526004016101649291906106f5565b5f60405180830381865afa15801561017e573d5f5f3e3d5ffd5b505050506040513d5f823e3d601f19601f820116820180604052508101906101a69190610798565b506001816101b491906107df565b9150808060010191505061011f565b509392505050565b606061100173ffffffffffffffffffffffffffffffffffffffff1663aded76b684846040518363ffffffff1660e01b815260040161020a9291906106f5565b5f60405180830381865afa158015610224573d5f5f3e3d5ffd5b505050506040513d5f823e3d601f19601f8201168201806040525081019061024c9190610798565b905092915050565b5f604051905090565b5f5ffd5b5f5ffd5b5f819050919050565b61027781610265565b8114610281575f5ffd5b50565b5f813590506102928161026e565b92915050565b5f602082840312156102ad576102ac61025d565b5b5f6102ba84828501610284565b91505092915050565b6102cc81610265565b82525050565b5f6020820190506102e55f8301846102c3565b92915050565b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b610339826102f3565b810181811067ffffffffffffffff8211171561035857610357610303565b5b80604052505050565b5f61036a610254565b90506103768282610330565b919050565b5f67ffffffffffffffff82111561039557610394610303565b5b61039e826102f3565b9050602081019050919050565b828183375f83830152505050565b5f6103cb6103c68461037b565b610361565b9050828152602081018484840111156103e7576103e66102ef565b5b6103f28482856103ab565b509392505050565b5f82601f83011261040e5761040d6102eb565b5b813561041e8482602086016103b9565b91505092915050565b5f67ffffffffffffffff82111561044157610440610303565b5b61044a826102f3565b9050602081019050919050565b5f61046961046484610427565b610361565b905082815260208101848484011115610485576104846102ef565b5b6104908482856103ab565b509392505050565b5f82601f8301126104ac576104ab6102eb565b5b81356104bc848260208601610457565b91505092915050565b5f5f5f606084860312156104dc576104db61025d565b5b5f84013567ffffffffffffffff8111156104f9576104f8610261565b5b610505868287016103fa565b935050602084013567ffffffffffffffff81111561052657610525610261565b5b61053286828701610498565b925050604061054386828701610284565b9150509250925092565b5f5f604083850312156105635761056261025d565b5b5f83013567ffffffffffffffff8111156105805761057f610261565b5b61058c858286016103fa565b925050602083013567ffffffffffffffff8111156105ad576105ac610261565b5b6105b985828601610498565b9150509250929050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f6105f5826105c3565b6105ff81856105cd565b935061060f8185602086016105dd565b610618816102f3565b840191505092915050565b5f6020820190508181035f83015261063b81846105eb565b905092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f61067a82610265565b915061068583610265565b925082820390508181111561069d5761069c610643565b5b92915050565b5f81519050919050565b5f82825260208201905092915050565b5f6106c7826106a3565b6106d181856106ad565b93506106e18185602086016105dd565b6106ea816102f3565b840191505092915050565b5f6040820190508181035f83015261070d81856106bd565b9050818103602083015261072181846105eb565b90509392505050565b5f61073c61073784610427565b610361565b905082815260208101848484011115610758576107576102ef565b5b6107638482856105dd565b509392505050565b5f82601f83011261077f5761077e6102eb565b5b815161078f84826020860161072a565b91505092915050565b5f602082840312156107ad576107ac61025d565b5b5f82015167ffffffffffffffff8111156107ca576107c9610261565b5b6107d68482850161076b565b91505092915050565b5f6107e982610265565b91506107f483610265565b925082820190508082111561080c5761080b610643565b5b9291505056fea2646970667358221220407835872f2f4d0bc80ebeef7df50e807d9ea632d2bdfc01befc11a14aa24c7a64736f6c634300081c0033",
}

// ReentranceABI is the input ABI used to generate the binding from.
// Deprecated: Use ReentranceMetaData.ABI instead.
var ReentranceABI = ReentranceMetaData.ABI

// ReentranceBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ReentranceMetaData.Bin instead.
var ReentranceBin = ReentranceMetaData.Bin

// DeployReentrance deploys a new Ethereum contract, binding an instance of Reentrance to it.
func DeployReentrance(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Reentrance, error) {
	parsed, err := ReentranceMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ReentranceBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Reentrance{ReentranceCaller: ReentranceCaller{contract: contract}, ReentranceTransactor: ReentranceTransactor{contract: contract}, ReentranceFilterer: ReentranceFilterer{contract: contract}}, nil
}

// Reentrance is an auto generated Go binding around an Ethereum contract.
type Reentrance struct {
	ReentranceCaller     // Read-only binding to the contract
	ReentranceTransactor // Write-only binding to the contract
	ReentranceFilterer   // Log filterer for contract events
}

// ReentranceCaller is an auto generated read-only Go binding around an Ethereum contract.
type ReentranceCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReentranceTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ReentranceTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReentranceFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ReentranceFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReentranceSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ReentranceSession struct {
	Contract     *Reentrance       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ReentranceCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ReentranceCallerSession struct {
	Contract *ReentranceCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// ReentranceTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ReentranceTransactorSession struct {
	Contract     *ReentranceTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// ReentranceRaw is an auto generated low-level Go binding around an Ethereum contract.
type ReentranceRaw struct {
	Contract *Reentrance // Generic contract binding to access the raw methods on
}

// ReentranceCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ReentranceCallerRaw struct {
	Contract *ReentranceCaller // Generic read-only contract binding to access the raw methods on
}

// ReentranceTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ReentranceTransactorRaw struct {
	Contract *ReentranceTransactor // Generic write-only contract binding to access the raw methods on
}

// NewReentrance creates a new instance of Reentrance, bound to a specific deployed contract.
func NewReentrance(address common.Address, backend bind.ContractBackend) (*Reentrance, error) {
	contract, err := bindReentrance(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Reentrance{ReentranceCaller: ReentranceCaller{contract: contract}, ReentranceTransactor: ReentranceTransactor{contract: contract}, ReentranceFilterer: ReentranceFilterer{contract: contract}}, nil
}

// NewReentranceCaller creates a new read-only instance of Reentrance, bound to a specific deployed contract.
func NewReentranceCaller(address common.Address, caller bind.ContractCaller) (*ReentranceCaller, error) {
	contract, err := bindReentrance(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ReentranceCaller{contract: contract}, nil
}

// NewReentranceTransactor creates a new write-only instance of Reentrance, bound to a specific deployed contract.
func NewReentranceTransactor(address common.Address, transactor bind.ContractTransactor) (*ReentranceTransactor, error) {
	contract, err := bindReentrance(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ReentranceTransactor{contract: contract}, nil
}

// NewReentranceFilterer creates a new log filterer instance of Reentrance, bound to a specific deployed contract.
func NewReentranceFilterer(address common.Address, filterer bind.ContractFilterer) (*ReentranceFilterer, error) {
	contract, err := bindReentrance(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ReentranceFilterer{contract: contract}, nil
}

// bindReentrance binds a generic wrapper to an already deployed contract.
func bindReentrance(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ReentranceMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Reentrance *ReentranceRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Reentrance.Contract.ReentranceCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Reentrance *ReentranceRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Reentrance.Contract.ReentranceTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Reentrance *ReentranceRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Reentrance.Contract.ReentranceTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Reentrance *ReentranceCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Reentrance.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Reentrance *ReentranceTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Reentrance.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Reentrance *ReentranceTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Reentrance.Contract.contract.Transact(opts, method, params...)
}

// Ping is a free data retrieval call binding the contract method 0x773acdef.
//
// Solidity: function ping(uint256 gasToConsume) view returns(uint256)
func (_Reentrance *ReentranceCaller) Ping(opts *bind.CallOpts, gasToConsume *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Reentrance.contract.Call(opts, &out, "ping", gasToConsume)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Ping is a free data retrieval call binding the contract method 0x773acdef.
//
// Solidity: function ping(uint256 gasToConsume) view returns(uint256)
func (_Reentrance *ReentranceSession) Ping(gasToConsume *big.Int) (*big.Int, error) {
	return _Reentrance.Contract.Ping(&_Reentrance.CallOpts, gasToConsume)
}

// Ping is a free data retrieval call binding the contract method 0x773acdef.
//
// Solidity: function ping(uint256 gasToConsume) view returns(uint256)
func (_Reentrance *ReentranceCallerSession) Ping(gasToConsume *big.Int) (*big.Int, error) {
	return _Reentrance.Contract.Ping(&_Reentrance.CallOpts, gasToConsume)
}

// ReentranceQuery is a free data retrieval call binding the contract method 0xa7c826f7.
//
// Solidity: function reentranceQuery(string wasmBech32, bytes payload) view returns(bytes)
func (_Reentrance *ReentranceCaller) ReentranceQuery(opts *bind.CallOpts, wasmBech32 string, payload []byte) ([]byte, error) {
	var out []interface{}
	err := _Reentrance.contract.Call(opts, &out, "reentranceQuery", wasmBech32, payload)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// ReentranceQuery is a free data retrieval call binding the contract method 0xa7c826f7.
//
// Solidity: function reentranceQuery(string wasmBech32, bytes payload) view returns(bytes)
func (_Reentrance *ReentranceSession) ReentranceQuery(wasmBech32 string, payload []byte) ([]byte, error) {
	return _Reentrance.Contract.ReentranceQuery(&_Reentrance.CallOpts, wasmBech32, payload)
}

// ReentranceQuery is a free data retrieval call binding the contract method 0xa7c826f7.
//
// Solidity: function reentranceQuery(string wasmBech32, bytes payload) view returns(bytes)
func (_Reentrance *ReentranceCallerSession) ReentranceQuery(wasmBech32 string, payload []byte) ([]byte, error) {
	return _Reentrance.Contract.ReentranceQuery(&_Reentrance.CallOpts, wasmBech32, payload)
}

// Spin is a free data retrieval call binding the contract method 0x85530c74.
//
// Solidity: function spin(string wasmBech32, bytes payload, uint256 iterations) view returns(uint256 n)
func (_Reentrance *ReentranceCaller) Spin(opts *bind.CallOpts, wasmBech32 string, payload []byte, iterations *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Reentrance.contract.Call(opts, &out, "spin", wasmBech32, payload, iterations)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Spin is a free data retrieval call binding the contract method 0x85530c74.
//
// Solidity: function spin(string wasmBech32, bytes payload, uint256 iterations) view returns(uint256 n)
func (_Reentrance *ReentranceSession) Spin(wasmBech32 string, payload []byte, iterations *big.Int) (*big.Int, error) {
	return _Reentrance.Contract.Spin(&_Reentrance.CallOpts, wasmBech32, payload, iterations)
}

// Spin is a free data retrieval call binding the contract method 0x85530c74.
//
// Solidity: function spin(string wasmBech32, bytes payload, uint256 iterations) view returns(uint256 n)
func (_Reentrance *ReentranceCallerSession) Spin(wasmBech32 string, payload []byte, iterations *big.Int) (*big.Int, error) {
	return _Reentrance.Contract.Spin(&_Reentrance.CallOpts, wasmBech32, payload, iterations)
}
