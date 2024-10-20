package state

import (
	"encoding/binary"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UkiiToSweiMultiplier Fields that were denominated in ukii will be converted to swei (1ukii = 10^12swei)
// for existing Ethereum application (which assumes 18 decimal points) to display properly.
var UkiiToSweiMultiplier = big.NewInt(1_000_000_000_000)
var SdkUkiiToSweiMultiplier = sdk.NewIntFromBigInt(UkiiToSweiMultiplier)

var CoinbaseAddressPrefix = []byte("evm_coinbase")

func GetCoinbaseAddress(txIdx int) sdk.AccAddress {
	txIndexBz := make([]byte, 8)
	binary.BigEndian.PutUint64(txIndexBz, uint64(txIdx))
	return append(CoinbaseAddressPrefix, txIndexBz...)
}

func SplitUkiiWeiAmount(amt *big.Int) (sdk.Int, sdk.Int) {
	wei := new(big.Int).Mod(amt, UkiiToSweiMultiplier)
	ukii := new(big.Int).Quo(amt, UkiiToSweiMultiplier)
	return sdk.NewIntFromBigInt(ukii), sdk.NewIntFromBigInt(wei)
}
