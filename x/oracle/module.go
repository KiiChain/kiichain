package oracle

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/kiichain/kiichain3/x/oracle/client/cli"
	"github.com/kiichain/kiichain3/x/oracle/client/rest"
	"github.com/kiichain/kiichain3/x/oracle/keeper"
	"github.com/kiichain/kiichain3/x/oracle/types"
	"github.com/spf13/cobra"
)

var (
	// _ module.AppModule      = AppModule{} // Indirect implement the AppModule interface
	_ module.AppModuleBasic = AppModule{} // Indirect implement the AppModuleBasic interface
)

// AppModule implements the Cosmos SDK AppModule interface
type AppModule struct {
	cdc    codec.Codec
	Kepper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		cdc:    cdc,
		Kepper: keeper,
	}
}

// ********************* IMPLEMENT AppModule INTERFACE ************************

// ****************************************************************************

// ********************* IMPLEMENT AppModuleBasic INTERFACE ******************
func (AppModule) Name() string {
	return types.ModuleName
}

func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterCodec(cdc)
}

func (AppModule) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

func (AppModule) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	err := cdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return types.ValidateGenesis(&data)
}

func (appModule AppModule) ValidateGenesisStream(cdc codec.JSONCodec, config client.TxEncodingConfig, genesisCh <-chan json.RawMessage) error {
	for genesis := range genesisCh {
		err := appModule.ValidateGenesis(cdc, config, genesis)
		if err != nil {
			return err
		}
	}
	return nil
}

func (AppModule) RegisterRESTRoutes(clientCtx client.Context, router *mux.Router) {
	rest.RegisterRoutes(clientCtx, router)
}

func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// TODO: Register gRPC query routes
}

func (AppModule) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

func (AppModule) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// ****************************************************************************
