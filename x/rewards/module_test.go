package rewards

import (
	"context"
	"testing"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkruntime "github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kiichain/kiichain/v7/x/rewards/keeper"
	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

type nopBank struct{}

func (nopBank) SendCoinsFromModuleToModule(context.Context, string, string, sdk.Coins) error {
	return nil
}

func (nopBank) GetBalance(context.Context, sdk.AccAddress, string) sdk.Coin {
	return sdk.Coin{}
}

type nopStaking struct{}

func (nopStaking) BondedRatio(context.Context) (math.LegacyDec, error) {
	return math.LegacyZeroDec(), nil
}

func TestModuleRegistrations(t *testing.T) {
	basic := NewAppModuleBasic()
	amino := codec.NewLegacyAmino()
	basic.RegisterLegacyAminoCodec(amino)

	registry := codectypes.NewInterfaceRegistry()
	basic.RegisterInterfaces(registry)

	clientCtx := client.Context{}.WithCodec(codec.NewProtoCodec(registry)).WithInterfaceRegistry(registry)
	basic.RegisterGRPCGatewayRoutes(clientCtx, runtime.NewServeMux())

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	cdc := codec.NewProtoCodec(registry)
	k := keeper.NewKeeper(
		cdc,
		sdkruntime.NewKVStoreService(storeKey),
		nopBank{},
		nopStaking{},
		authtypes.NewModuleAddress("gov").String(),
		authtypes.FeeCollectorName,
	)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Time: time.Now().UTC()}, false, log.NewNopLogger())
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))
	require.NoError(t, k.RewardPool.Set(ctx, types.RewardPool{}))

	am := NewAppModule(k, nopBank{})
	msgRouter := baseapp.NewMsgServiceRouter()
	queryRouter := baseapp.NewGRPCQueryRouter()
	msgRouter.SetInterfaceRegistry(registry)
	queryRouter.SetInterfaceRegistry(registry)
	cfg := module.NewConfigurator(cdc, msgRouter, queryRouter)
	mm := module.NewManager(am)
	require.NoError(t, mm.RegisterServices(cfg))

	vm := mm.GetVersionMap()
	vm[types.ModuleName] = 1
	_, err := mm.RunMigrations(ctx, cfg, vm)
	require.NoError(t, err)

	require.Panics(t, func() {
		am.RegisterServices(cfg)
	})
}
