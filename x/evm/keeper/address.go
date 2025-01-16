package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kiichain/kiichain3/x/evm/types"
)

func (k *Keeper) SetAddressMapping(ctx sdk.Context, kiiAddress sdk.AccAddress, evmAddress common.Address) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.EVMAddressToKiiAddressKey(evmAddress), kiiAddress)
	store.Set(types.KiiAddressToEVMAddressKey(kiiAddress), evmAddress[:])
	if !k.accountKeeper.HasAccount(ctx, kiiAddress) {
		k.accountKeeper.SetAccount(ctx, k.accountKeeper.NewAccountWithAddress(ctx, kiiAddress))
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeAddressAssociated,
		sdk.NewAttribute(types.AttributeKeyKiiAddress, kiiAddress.String()),
		sdk.NewAttribute(types.AttributeKeyEvmAddress, evmAddress.Hex()),
	))
}

func (k *Keeper) DeleteAddressMapping(ctx sdk.Context, kiiAddress sdk.AccAddress, evmAddress common.Address) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.EVMAddressToKiiAddressKey(evmAddress))
	store.Delete(types.KiiAddressToEVMAddressKey(kiiAddress))
}

func (k *Keeper) GetEVMAddress(ctx sdk.Context, kiiAddress sdk.AccAddress) (common.Address, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KiiAddressToEVMAddressKey(kiiAddress))
	addr := common.Address{}
	if bz == nil {
		return addr, false
	}
	copy(addr[:], bz)
	return addr, true
}

func (k *Keeper) GetEVMAddressOrDefault(ctx sdk.Context, kiiAddress sdk.AccAddress) common.Address {
	addr, ok := k.GetEVMAddress(ctx, kiiAddress)
	if ok {
		return addr
	}
	return common.BytesToAddress(kiiAddress)
}

func (k *Keeper) GetKiiAddress(ctx sdk.Context, evmAddress common.Address) (sdk.AccAddress, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.EVMAddressToKiiAddressKey(evmAddress))
	if bz == nil {
		return []byte{}, false
	}
	return bz, true
}

func (k *Keeper) GetKiiAddressOrDefault(ctx sdk.Context, evmAddress common.Address) sdk.AccAddress {
	addr, ok := k.GetKiiAddress(ctx, evmAddress)
	if ok {
		return addr
	}
	return sdk.AccAddress(evmAddress[:])
}

func (k *Keeper) IterateKiiAddressMapping(ctx sdk.Context, cb func(evmAddr common.Address, kiiAddr sdk.AccAddress) bool) {
	iter := prefix.NewStore(ctx.KVStore(k.storeKey), types.EVMAddressToKiiAddressKeyPrefix).Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		evmAddr := common.BytesToAddress(iter.Key())
		kiiAddr := sdk.AccAddress(iter.Value())
		if cb(evmAddr, kiiAddr) {
			break
		}
	}
}

// A sdk.AccAddress may not receive funds from bank if it's the result of direct-casting
// from an EVM address AND the originating EVM address has already been associated with
// a true (i.e. derived from the same pubkey) sdk.AccAddress.
func (k *Keeper) CanAddressReceive(ctx sdk.Context, addr sdk.AccAddress) bool {
	directCast := common.BytesToAddress(addr) // casting goes both directions since both address formats have 20 bytes
	associatedAddr, isAssociated := k.GetKiiAddress(ctx, directCast)
	// if the associated address is the cast address itself, allow the address to receive (e.g. EVM contract addresses)
	return associatedAddr.Equals(addr) || !isAssociated // this means it's either a cast address that's not associated yet, or not a cast address at all.
}
