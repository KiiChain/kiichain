package feeabstraction

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kiichain/kiichain/v4/x/feeabstraction/client/cli"
	"github.com/kiichain/kiichain/v4/x/feeabstraction/keeper"
	"github.com/kiichain/kiichain/v4/x/feeabstraction/types"
)

// Interface inference
var (
	_ module.AppModuleBasic     = AppModuleBasic{}
	_ module.HasGenesisBasics   = AppModuleBasic{}
	_ appmodule.HasBeginBlocker = AppModule{}
	_ module.AppModule          = AppModule{}
	_ module.HasABCIGenesis     = AppModule{}
)

// ConsensusVersion defines the current x/feeabstraction module consensus version
const ConsensusVersion = 1

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic implements the AppModuleBasic interface for the capability module
type AppModuleBasic struct{}

// NewAppModuleBasic returns a new app module basic
func NewAppModuleBasic() AppModuleBasic {
	return AppModuleBasic{}
}

// Name returns the x/feeabstraction module name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAmnioCodec register the amino codecs in the legacy format
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces register all the interfaces from proto
func (a AppModuleBasic) RegisterInterfaces(cdc codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(cdc)
}

// RegisterGRPCGatewayRoutes register all the GRPC routes for the project
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
	if err != nil {
		panic(err)
	}
}

// DefaultGenesis returns the x/feeabstraction default genesis
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	// Unmarshal the genesis state
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	// Validate and return
	return genState.Validate()
}

// GetTxCmd returns the module root Tx CLI command
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the module root query CLI command
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule defines the AppModule while keeping the AppModuleBasic interface
type AppModule struct {
	// Wrap the AppModuleBasic
	AppModuleBasic

	// Has the keeper as param
	keeper keeper.Keeper
}

// NewAppModule returns the a new AppModule
func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(),
		keeper:         k,
	}
}

// IsAppModule implement the AppModule interface
func (AppModule) IsAppModule() {}

// IsOnePerModuleType implements the AppModule interface
func (AppModule) IsOnePerModuleType() {}

// Name returns the x/feeabstraction module name
func (am AppModule) Name() string { return am.AppModuleBasic.Name() }

// QuerierRoute returns the x/feeabstraction module query route key
func (AppModule) QuerierRoute() string { return types.QuerierRoute }

// RegisterServices registers the GRPC query and msg servers
func (am AppModule) RegisterServices(c module.Configurator) {
	types.RegisterMsgServer(c.MsgServer(), keeper.NewMsgServer(am.keeper))
	types.RegisterQueryServer(c.QueryServer(), keeper.NewQuerier(am.keeper))
}

// RegisterInvariants register the module invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs the module genesis initialization
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	// Unmarshal the genesis state
	var genState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genState)

	// Initialize the genesis
	err := am.keeper.InitGenesis(ctx, genState)
	if err != nil {
		panic(err)
	}
	// Return no validator updates
	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports the module genesis in raw json bytes
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	// Export the keeper contents as a genesis state
	genState, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion returns the module consensus version
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// BeginBlock returns the begin blocker for the module
func (am AppModule) BeginBlock(ctx context.Context) error {
	return am.keeper.BeginBlocker(ctx)
}
