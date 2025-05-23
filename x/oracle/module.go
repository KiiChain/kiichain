package oracle

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/kiichain/kiichain/v1/x/oracle/client/cli"
	"github.com/kiichain/kiichain/v1/x/oracle/client/rest"
	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}      // Indirect implement the AppModule interface
	_ module.AppModuleBasic = AppModuleBasic{} // Indirect implement the AppModuleBasic interface
)

// ********************* IMPLEMENT AppModuleBasic INTERFACE ******************

// AppModuleBasic defines the basic application module
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the module name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterCodec(cdc)
}

// RegisterInterfaces registers the request messages on the tx rpc
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns the default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs a genesis state validation
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	err := cdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return types.ValidateGenesis(&data)
}

// ValidateGenesisStream performs a genesis validation in a streaming fashion
func (appModule AppModuleBasic) ValidateGenesisStream(cdc codec.JSONCodec, config client.TxEncodingConfig, genesisCh <-chan json.RawMessage) error {
	for genesis := range genesisCh {
		err := appModule.ValidateGenesis(cdc, config, genesis)
		if err != nil {
			return err
		}
	}
	return nil
}

// RegisterRESTRoutes registers the REST routes
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, router *mux.Router) {
	rest.RegisterRoutes(clientCtx, router)
}

// RegisterGRPCGatewayRoutes registers the gRPC query routes
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
	if err != nil {
		panic(err)
	}
}

// GetTxCmd returns the cli tx commands for the module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the cli query commands for the module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// ****************************************************************************

// AppModule implements an application module (AppModule interface)
type AppModule struct {
	AppModuleBasic
	Kepper        keeper.Keeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{
			cdc: cdc,
		},
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		Kepper:        keeper,
	}
}

// ********************* IMPLEMENT AppModule INTERFACE ************************
// ConsensusVersion returns the version the module's version
func (AppModule) ConsensusVersion() uint64 { return 6 }

// RegisterServices registers the module services
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.Kepper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.Kepper))
}

// RegisterInvariants
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis trigger the genesis initialization
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	genesis := &types.GenesisState{}
	cdc.MustUnmarshalJSON(data, genesis)
	InitGenesis(ctx, am.Kepper, genesis)
	return nil
}

// ExportGenesis returns the current genesis state as json
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genesis := ExportGenesis(ctx, am.Kepper)
	return cdc.MustMarshalJSON(&genesis)
}

// ExportGenesisStream returns genesis state as json bytes in a streaming fashion
func (am AppModule) ExportGenesisStream(ctx sdk.Context, cdc codec.JSONCodec) <-chan json.RawMessage {
	ch := make(chan json.RawMessage)
	go func() {
		ch <- am.ExportGenesis(ctx, cdc)
		close(ch)
	}()
	return ch
}

// LegacyQuerierHandler returns the module sdk.Querier (deprecated)
func (am AppModule) LegacyQuerierHandler(_ *codec.LegacyAmino) sdk.Querier {
	return nil
}

// Route returns the module's routing key (deprecated)
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.Kepper))
}

// QuerierRoute returns the module's querier router name
func (am AppModule) QuerierRoute() string { return types.QuerierRoute }

// BeginBlock returns the module's begin blocker
func (am AppModule) BeginBlock(_ sdk.Context, _ int64) {}

// MidBlock returns the module's mid blocker
func (am AppModule) MidBlock(ctx sdk.Context, _ int64) {
	MidBlocker(ctx, am.Kepper)
}

// EndBlock returns the module's end blocker
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) (res []abci.ValidatorUpdate) {
	Endblocker(ctx, am.Kepper)
	return []abci.ValidatorUpdate{}
}

// ****************************************************************************
