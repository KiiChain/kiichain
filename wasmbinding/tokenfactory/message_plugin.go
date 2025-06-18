package tokenfactory

import (
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	tfbindingtypes "github.com/kiichain/kiichain/v2/wasmbinding/tokenfactory/types"
	"github.com/kiichain/kiichain/v2/wasmbinding/utils"
	tokenfactorykeeper "github.com/kiichain/kiichain/v2/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/kiichain/kiichain/v2/x/tokenfactory/types"
)

// CustomMessenger is a wrapper for the token factory message plugin
type CustomMessenger struct {
	bank         bankkeeper.Keeper
	tokenFactory *tokenfactorykeeper.Keeper
}

// DispatchMsg implements keeper.Messenger
func (m *CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg tfbindingtypes.Msg) (events []sdk.Event, data [][]byte, msgResponses [][]*types.Any, err error) {
	// Match the message
	switch {
	case msg.CreateDenom != nil:
		return m.CreateDenom(ctx, contractAddr, msg.CreateDenom)
	case msg.MintTokens != nil:
		return m.MintTokens(ctx, contractAddr, msg.MintTokens)
	case msg.ChangeAdmin != nil:
		return m.ChangeAdmin(ctx, contractAddr, msg.ChangeAdmin)
	case msg.BurnTokens != nil:
		return m.BurnTokens(ctx, contractAddr, msg.BurnTokens)
	case msg.SetMetadata != nil:
		return m.SetMetadata(ctx, contractAddr, msg.SetMetadata)
	case msg.ForceTransfer != nil:
		return m.ForceTransfer(ctx, contractAddr, msg.ForceTransfer)
	default:
		return nil, nil, utils.EmptyMsgResp, wasmvmtypes.UnsupportedRequest{Kind: "unknown token factory msg variant"}
	}
}

// NewCustomMessenger returns a reference to a new CustomMessenger
func NewCustomMessenger(bank bankkeeper.Keeper, tokenFactory *tokenfactorykeeper.Keeper) *CustomMessenger {
	return &CustomMessenger{
		bank:         bank,
		tokenFactory: tokenFactory,
	}
}

// CreateDenom creates a new token denom
func (m *CustomMessenger) CreateDenom(ctx sdk.Context, contractAddr sdk.AccAddress, createDenom *tfbindingtypes.CreateDenom) ([]sdk.Event, [][]byte, [][]*types.Any, error) {
	bz, err := PerformCreateDenom(m.tokenFactory, m.bank, ctx, contractAddr, createDenom)
	if err != nil {
		return nil, nil, utils.EmptyMsgResp, errorsmod.Wrap(err, "perform create denom")
	}
	return nil, [][]byte{bz}, utils.EmptyMsgResp, nil
}

// PerformCreateDenom is used with createDenom to create a token denom; validates the msgCreateDenom.
func PerformCreateDenom(f *tokenfactorykeeper.Keeper, b bankkeeper.Keeper, ctx sdk.Context, contractAddr sdk.AccAddress, createDenom *tfbindingtypes.CreateDenom) ([]byte, error) {
	if createDenom == nil {
		return nil, wasmvmtypes.InvalidRequest{Err: "create denom null create denom"}
	}

	msgServer := tokenfactorykeeper.NewMsgServerImpl(*f)

	msgCreateDenom := tokenfactorytypes.NewMsgCreateDenom(contractAddr.String(), createDenom.Subdenom)

	if err := msgCreateDenom.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "failed validating MsgCreateDenom")
	}

	// Create denom
	resp, err := msgServer.CreateDenom(
		ctx,
		msgCreateDenom,
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "creating denom")
	}

	if createDenom.Metadata != nil {
		newDenom := resp.NewTokenDenom
		err := PerformSetMetadata(f, b, ctx, contractAddr, newDenom, *createDenom.Metadata)
		if err != nil {
			return nil, errorsmod.Wrap(err, "setting metadata")
		}
	}

	return resp.Marshal()
}

// MintTokens mints tokens of a specified denom to an address.
func (m *CustomMessenger) MintTokens(ctx sdk.Context, contractAddr sdk.AccAddress, mint *tfbindingtypes.MintTokens) ([]sdk.Event, [][]byte, [][]*types.Any, error) {
	err := PerformMint(m.tokenFactory, m.bank, ctx, contractAddr, mint)
	if err != nil {
		return nil, nil, utils.EmptyMsgResp, errorsmod.Wrap(err, "perform mint")
	}
	return nil, nil, utils.EmptyMsgResp, nil
}

// PerformMint used with mintTokens to validate the mint message and mint through token factory.
func PerformMint(f *tokenfactorykeeper.Keeper, b bankkeeper.Keeper, ctx sdk.Context, contractAddr sdk.AccAddress, mint *tfbindingtypes.MintTokens) error {
	if mint == nil {
		return wasmvmtypes.InvalidRequest{Err: "mint token null mint"}
	}
	rcpt, err := parseAddress(mint.MintToAddress)
	if err != nil {
		return err
	}

	coin := sdk.Coin{Denom: mint.Denom, Amount: mint.Amount}
	sdkMsg := tokenfactorytypes.NewMsgMint(contractAddr.String(), coin)

	if err = sdkMsg.ValidateBasic(); err != nil {
		return err
	}

	// Mint through token factory / message server
	msgServer := tokenfactorykeeper.NewMsgServerImpl(*f)
	_, err = msgServer.Mint(ctx, sdkMsg)
	if err != nil {
		return errorsmod.Wrap(err, "minting coins from message")
	}

	if b.BlockedAddr(rcpt) {
		return errorsmod.Wrapf(err, "minting coins to blocked address %s", rcpt.String())
	}

	err = b.SendCoins(ctx, contractAddr, rcpt, sdk.NewCoins(coin))
	if err != nil {
		return errorsmod.Wrap(err, "sending newly minted coins from message")
	}
	return nil
}

// ChangeAdmin changes the admin.
func (m *CustomMessenger) ChangeAdmin(ctx sdk.Context, contractAddr sdk.AccAddress, changeAdmin *tfbindingtypes.ChangeAdmin) ([]sdk.Event, [][]byte, [][]*types.Any, error) {
	err := ChangeAdmin(m.tokenFactory, ctx, contractAddr, changeAdmin)
	if err != nil {
		return nil, nil, utils.EmptyMsgResp, errorsmod.Wrap(err, "failed to change admin")
	}
	return nil, nil, utils.EmptyMsgResp, nil
}

// ChangeAdmin is used with changeAdmin to validate changeAdmin messages and to dispatch.
func ChangeAdmin(f *tokenfactorykeeper.Keeper, ctx sdk.Context, contractAddr sdk.AccAddress, changeAdmin *tfbindingtypes.ChangeAdmin) error {
	if changeAdmin == nil {
		return wasmvmtypes.InvalidRequest{Err: "changeAdmin is nil"}
	}
	newAdminAddr, err := parseAddress(changeAdmin.NewAdminAddress)
	if err != nil {
		return err
	}

	changeAdminMsg := tokenfactorytypes.NewMsgChangeAdmin(contractAddr.String(), changeAdmin.Denom, newAdminAddr.String())
	if err := changeAdminMsg.ValidateBasic(); err != nil {
		return err
	}

	msgServer := tokenfactorykeeper.NewMsgServerImpl(*f)
	_, err = msgServer.ChangeAdmin(ctx, changeAdminMsg)
	if err != nil {
		return errorsmod.Wrap(err, "failed changing admin from message")
	}
	return nil
}

// BurnTokens burns tokens.
func (m *CustomMessenger) BurnTokens(ctx sdk.Context, contractAddr sdk.AccAddress, burn *tfbindingtypes.BurnTokens) ([]sdk.Event, [][]byte, [][]*types.Any, error) {
	err := PerformBurn(m.tokenFactory, ctx, contractAddr, burn)
	if err != nil {
		return nil, nil, utils.EmptyMsgResp, errorsmod.Wrap(err, "perform burn")
	}
	return nil, nil, utils.EmptyMsgResp, nil
}

// PerformBurn performs token burning after validating tokenBurn message.
func PerformBurn(f *tokenfactorykeeper.Keeper, ctx sdk.Context, contractAddr sdk.AccAddress, burn *tfbindingtypes.BurnTokens) error {
	if burn == nil {
		return wasmvmtypes.InvalidRequest{Err: "burn token null mint"}
	}

	coin := sdk.Coin{Denom: burn.Denom, Amount: burn.Amount}
	sdkMsg := tokenfactorytypes.NewMsgBurn(contractAddr.String(), coin)
	if burn.BurnFromAddress != "" {
		sdkMsg = tokenfactorytypes.NewMsgBurnFrom(contractAddr.String(), coin, burn.BurnFromAddress)
	}

	if err := sdkMsg.ValidateBasic(); err != nil {
		return err
	}

	// Burn through token factory / message server
	msgServer := tokenfactorykeeper.NewMsgServerImpl(*f)
	_, err := msgServer.Burn(ctx, sdkMsg)
	if err != nil {
		return errorsmod.Wrap(err, "burning coins from message")
	}
	return nil
}

// ForceTransfer moves tokens.
func (m *CustomMessenger) ForceTransfer(ctx sdk.Context, contractAddr sdk.AccAddress, forcetransfer *tfbindingtypes.ForceTransfer) ([]sdk.Event, [][]byte, [][]*types.Any, error) {
	err := PerformForceTransfer(m.tokenFactory, ctx, contractAddr, forcetransfer)
	if err != nil {
		return nil, nil, utils.EmptyMsgResp, errorsmod.Wrap(err, "perform force transfer")
	}
	return nil, nil, utils.EmptyMsgResp, nil
}

// PerformForceTransfer performs token moving after validating tokenForceTransfer message.
func PerformForceTransfer(f *tokenfactorykeeper.Keeper, ctx sdk.Context, contractAddr sdk.AccAddress, forcetransfer *tfbindingtypes.ForceTransfer) error {
	if forcetransfer == nil {
		return wasmvmtypes.InvalidRequest{Err: "force transfer null"}
	}

	_, err := parseAddress(forcetransfer.FromAddress)
	if err != nil {
		return err
	}

	_, err = parseAddress(forcetransfer.ToAddress)
	if err != nil {
		return err
	}

	coin := sdk.Coin{Denom: forcetransfer.Denom, Amount: forcetransfer.Amount}
	sdkMsg := tokenfactorytypes.NewMsgForceTransfer(contractAddr.String(), coin, forcetransfer.FromAddress, forcetransfer.ToAddress)

	if err := sdkMsg.ValidateBasic(); err != nil {
		return err
	}

	// Transfer through token factory / message server
	msgServer := tokenfactorykeeper.NewMsgServerImpl(*f)
	_, err = msgServer.ForceTransfer(ctx, sdkMsg)
	if err != nil {
		return errorsmod.Wrap(err, "force transferring from message")
	}
	return nil
}

// createDenom creates a new token denom
func (m *CustomMessenger) SetMetadata(ctx sdk.Context, contractAddr sdk.AccAddress, setMetadata *tfbindingtypes.SetMetadata) ([]sdk.Event, [][]byte, [][]*types.Any, error) {
	err := PerformSetMetadata(m.tokenFactory, m.bank, ctx, contractAddr, setMetadata.Denom, setMetadata.Metadata)
	if err != nil {
		return nil, nil, utils.EmptyMsgResp, errorsmod.Wrap(err, "perform set metadata")
	}
	return nil, nil, utils.EmptyMsgResp, nil
}

// PerformSetMetadata is used with setMetadata to add new metadata
// It also is called inside CreateDenom if optional metadata field is set
func PerformSetMetadata(f *tokenfactorykeeper.Keeper, b bankkeeper.Keeper, ctx sdk.Context, contractAddr sdk.AccAddress, denom string, metadata tfbindingtypes.Metadata) error {
	// ensure contract address is admin of denom
	auth, err := f.GetAuthorityMetadata(ctx, denom)
	if err != nil {
		return err
	}
	if auth.Admin != contractAddr.String() {
		return wasmvmtypes.InvalidRequest{Err: "only admin can set metadata"}
	}

	// ensure we are setting proper denom metadata (bank uses Base field, fill it if missing)
	if metadata.Base == "" {
		metadata.Base = denom
	} else if metadata.Base != denom {
		// this is the key that we set
		return wasmvmtypes.InvalidRequest{Err: "Base must be the same as denom"}
	}

	// Create and validate the metadata
	bankMetadata := WasmMetadataToSdk(metadata)
	if err := bankMetadata.Validate(); err != nil {
		return err
	}

	b.SetDenomMetaData(ctx, bankMetadata)
	return nil
}

// GetFullDenom is a function, not method, so the message_plugin can use it
func GetFullDenom(contract string, subDenom string) (string, error) {
	// Address validation
	if _, err := parseAddress(contract); err != nil {
		return "", err
	}
	fullDenom, err := tokenfactorytypes.GetTokenDenom(contract, subDenom)
	if err != nil {
		return "", errorsmod.Wrap(err, "validate sub-denom")
	}

	return fullDenom, nil
}

// parseAddress parses address from bech32 string and verifies its format.
func parseAddress(addr string) (sdk.AccAddress, error) {
	parsed, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, errorsmod.Wrap(err, "address from bech32")
	}
	err = sdk.VerifyAddressFormat(parsed)
	if err != nil {
		return nil, errorsmod.Wrap(err, "verify address format")
	}
	return parsed, nil
}

// WasmMetadataToSdk converts a tokenfactory metadata to sdk metadata
func WasmMetadataToSdk(metadata tfbindingtypes.Metadata) banktypes.Metadata {
	denoms := []*banktypes.DenomUnit{}
	for _, unit := range metadata.DenomUnits {
		denoms = append(denoms, &banktypes.DenomUnit{
			Denom:    unit.Denom,
			Exponent: unit.Exponent,
			Aliases:  unit.Aliases,
		})
	}
	return banktypes.Metadata{
		Description: metadata.Description,
		Display:     metadata.Display,
		Base:        metadata.Base,
		Name:        metadata.Name,
		Symbol:      metadata.Symbol,
		DenomUnits:  denoms,
	}
}

// SdkMetadataToWasm converts a sdk metadata to tokenfactory metadata
func SdkMetadataToWasm(metadata banktypes.Metadata) *tfbindingtypes.Metadata {
	denoms := []tfbindingtypes.DenomUnit{}
	for _, unit := range metadata.DenomUnits {
		denoms = append(denoms, tfbindingtypes.DenomUnit{
			Denom:    unit.Denom,
			Exponent: unit.Exponent,
			Aliases:  unit.Aliases,
		})
	}
	return &tfbindingtypes.Metadata{
		Description: metadata.Description,
		Display:     metadata.Display,
		Base:        metadata.Base,
		Name:        metadata.Name,
		Symbol:      metadata.Symbol,
		DenomUnits:  denoms,
	}
}
