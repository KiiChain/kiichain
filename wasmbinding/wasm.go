package wasmbinding

import (
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	tfbinding "github.com/kiichain/kiichain/v1/wasmbinding/tokenfactory"
	tokenfactorykeeper "github.com/kiichain/kiichain/v1/x/tokenfactory/keeper"
)

// RegisterCustomPlugins registers custom plugins for the wasm module
func RegisterCustomPlugins(
	bank bankkeeper.Keeper,
	tokenFactory *tokenfactorykeeper.Keeper,
) []wasmkeeper.Option {
	// Register custom query plugins
	tokenFactoryQueryPlugin := tfbinding.NewQueryPlugin(bank, tokenFactory)

	// Create the central query plugin
	queryPlugin := NewQueryPlugin(tokenFactoryQueryPlugin)

	// Register custom message handler decorators
	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(queryPlugin),
	})

	// Create the custom messenger to the token factory
	tokenFactoryMessenger := tfbinding.NewCustomMessenger(bank, tokenFactory)

	// Initialize the decorator for the custom messenger
	messengerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(bank, tokenFactoryMessenger),
	)

	// Register custom message handlers
	return []wasmkeeper.Option{
		queryPluginOpt,
		messengerDecoratorOpt,
	}
}
