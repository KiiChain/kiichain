package oracle

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/kiichain/kiichain/v1/x/oracle/client/cli"
	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/spf13/cobra"
)

var (
	_ module.AppModule      = AppModule{}      // Indirect implement the AppModule interface
	_ module.AppModuleBasic = AppModuleBasic{} // Indirect implement the AppModuleBasic interface
)

// ConsensusVersion defines the current x/oracle module consensus version.
const ConsensusVersion = 1

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic defines the basic application module
type AppModuleBasic struct {
	cdc codec.Codec
}

// NewAppModuleBasic creates a new AppModuleBasic object
func NewAppModuleBasic(cdc codec.Codec) AppModuleBasic {
	return AppModuleBasic{
		cdc: cdc,
	}
}

// Name returns the module name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the module's types on the LegacyAmino codec
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
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	err := cdc.UnmarshalJSON(bz, &genState)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterRESTRoutes registers the REST routes
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, router *mux.Router) {
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

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

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

// IsAppModule implements the AppModule interface
func (AppModule) IsAppModule() {}

// IsOnePerModuleType implements the AppModule interface
func (AppModule) IsOnePerModuleType() {}

// Name returns the module name
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// QuerierRoute returns the module's querier router name
func (am AppModule) QuerierRoute() string { return types.QuerierRoute }

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

// ConsensusVersion returns the version the module's version
func (AppModule) ConsensusVersion() uint64 { return 6 }

// EndBlock returns the module's end blocker
func (am AppModule) EndBlock(ctx sdk.Context) (res []abci.ValidatorUpdate) {
	MidBlocker(ctx, am.Kepper)
	Endblocker(ctx, am.Kepper)
	return []abci.ValidatorUpdate{}
}
