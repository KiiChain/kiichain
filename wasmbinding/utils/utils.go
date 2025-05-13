package utils

import (
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ConvertSdkCoinsToWasmCoins converts sdk type coins to wasm vm type coins
func ConvertSdkCoinsToWasmCoins(coins []sdk.Coin) []wasmvmtypes.Coin {
	var toSend []wasmvmtypes.Coin
	for _, coin := range coins {
		c := ConvertSdkCoinToWasmCoin(coin)
		toSend = append(toSend, c)
	}
	return toSend
}

// ConvertSdkCoinToWasmCoin converts a sdk type coin to a wasm vm type coin
func ConvertSdkCoinToWasmCoin(coin sdk.Coin) wasmvmtypes.Coin {
	return wasmvmtypes.Coin{
		Denom: coin.Denom,
		// Note: tokenfactory tokens have 18 decimal places, so 10^22 is common, no longer in u64 range
		Amount: coin.Amount.String(),
	}
}
