package feeabstraction

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/client/cli"
	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// Interface inference
var (
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ module.HasGenesisBasics = AppModuleBasic{}
	// _ module.AppModule        = AppModule{}
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
