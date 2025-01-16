package wasmbinding

import (
	"encoding/json"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain3/utils/metrics"
	epochwasm "github.com/kiichain/kiichain3/x/epoch/client/wasm"
	epochbindings "github.com/kiichain/kiichain3/x/epoch/client/wasm/bindings"
	epochtypes "github.com/kiichain/kiichain3/x/epoch/types"
	evmwasm "github.com/kiichain/kiichain3/x/evm/client/wasm"
	evmbindings "github.com/kiichain/kiichain3/x/evm/client/wasm/bindings"

	tokenfactorywasm "github.com/kiichain/kiichain3/x/tokenfactory/client/wasm"
	tokenfactorybindings "github.com/kiichain/kiichain3/x/tokenfactory/client/wasm/bindings"
	tokenfactorytypes "github.com/kiichain/kiichain3/x/tokenfactory/types"
)

type QueryPlugin struct {
	epochHandler        epochwasm.EpochWasmQueryHandler
	tokenfactoryHandler tokenfactorywasm.TokenFactoryWasmQueryHandler
	evmHandler          evmwasm.EVMQueryHandler
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(eh *epochwasm.EpochWasmQueryHandler, th *tokenfactorywasm.TokenFactoryWasmQueryHandler, evmh *evmwasm.EVMQueryHandler) *QueryPlugin {
	return &QueryPlugin{
		epochHandler:        *eh,
		tokenfactoryHandler: *th,
		evmHandler:          *evmh,
	}
}

func (qp QueryPlugin) HandleEpochQuery(ctx sdk.Context, queryData json.RawMessage) ([]byte, error) {
	var parsedQuery epochbindings.KiiEpochQuery
	if err := json.Unmarshal(queryData, &parsedQuery); err != nil {
		return nil, epochtypes.ErrParsingKiiEpochQuery
	}
	switch {
	case parsedQuery.Epoch != nil:
		res, err := qp.epochHandler.GetEpoch(ctx, parsedQuery.Epoch)
		if err != nil {
			return nil, err
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, epochtypes.ErrEncodingEpoch
		}

		return bz, nil
	default:
		return nil, epochtypes.ErrUnknownKiiEpochQuery
	}
}

func (qp QueryPlugin) HandleTokenFactoryQuery(ctx sdk.Context, queryData json.RawMessage) ([]byte, error) {
	var parsedQuery tokenfactorybindings.KiiTokenFactoryQuery
	if err := json.Unmarshal(queryData, &parsedQuery); err != nil {
		return nil, tokenfactorytypes.ErrParsingKiiTokenFactoryQuery
	}
	switch {
	case parsedQuery.DenomAuthorityMetadata != nil:
		res, err := qp.tokenfactoryHandler.GetDenomAuthorityMetadata(ctx, parsedQuery.DenomAuthorityMetadata)
		if err != nil {
			return nil, err
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, tokenfactorytypes.ErrEncodingDenomAuthorityMetadata
		}

		return bz, nil
	case parsedQuery.DenomsFromCreator != nil:
		res, err := qp.tokenfactoryHandler.GetDenomsFromCreator(ctx, parsedQuery.DenomsFromCreator)
		if err != nil {
			return nil, err
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, tokenfactorytypes.ErrEncodingDenomsFromCreator
		}

		return bz, nil
	default:
		return nil, tokenfactorytypes.ErrUnknownKiiTokenFactoryQuery
	}
}

func (qp QueryPlugin) HandleEVMQuery(ctx sdk.Context, queryData json.RawMessage) (res []byte, err error) {
	var queryType evmbindings.EVMQueryType
	var parsedQuery evmbindings.KiiEVMQuery
	if err := json.Unmarshal(queryData, &parsedQuery); err != nil {
		return nil, errors.New("invalid EVM query")
	}
	queryType = parsedQuery.GetQueryType()

	defer func() {
		metrics.IncrementErrorMetrics(string(queryType), err)
	}()

	switch queryType {
	case evmbindings.StaticCallType:
		c := parsedQuery.StaticCall
		return qp.evmHandler.HandleStaticCall(ctx, c.From, c.To, c.Data)
	case evmbindings.ERC20TransferType:
		c := parsedQuery.ERC20TransferPayload
		return qp.evmHandler.HandleERC20TransferPayload(ctx, c.Recipient, c.Amount)
	case evmbindings.ERC20TransferFromType:
		c := parsedQuery.ERC20TransferFromPayload
		return qp.evmHandler.HandleERC20TransferFromPayload(ctx, c.Owner, c.Recipient, c.Amount)
	case evmbindings.ERC20ApproveType:
		c := parsedQuery.ERC20ApprovePayload
		return qp.evmHandler.HandleERC20ApprovePayload(ctx, c.Spender, c.Amount)
	case evmbindings.ERC20AllowanceType:
		c := parsedQuery.ERC20Allowance
		return qp.evmHandler.HandleERC20Allowance(ctx, c.ContractAddress, c.Owner, c.Spender)
	case evmbindings.ERC20TokenInfoType:
		c := parsedQuery.ERC20TokenInfo
		return qp.evmHandler.HandleERC20TokenInfo(ctx, c.ContractAddress, c.Caller)
	case evmbindings.ERC20BalanceType:
		c := parsedQuery.ERC20Balance
		return qp.evmHandler.HandleERC20Balance(ctx, c.ContractAddress, c.Account)
	case evmbindings.ERC721OwnerType:
		c := parsedQuery.ERC721Owner
		return qp.evmHandler.HandleERC721Owner(ctx, c.Caller, c.ContractAddress, c.TokenID)
	case evmbindings.ERC721TransferType:
		c := parsedQuery.ERC721TransferPayload
		return qp.evmHandler.HandleERC721TransferPayload(ctx, c.From, c.Recipient, c.TokenID)
	case evmbindings.ERC721ApproveType:
		c := parsedQuery.ERC721ApprovePayload
		return qp.evmHandler.HandleERC721ApprovePayload(ctx, c.Spender, c.TokenID)
	case evmbindings.ERC721SetApprovalAllType:
		c := parsedQuery.ERC721SetApprovalAllPayload
		return qp.evmHandler.HandleERC721SetApprovalAllPayload(ctx, c.To, c.Approved)
	case evmbindings.ERC721ApprovedType:
		c := parsedQuery.ERC721Approved
		return qp.evmHandler.HandleERC721Approved(ctx, c.Caller, c.ContractAddress, c.TokenID)
	case evmbindings.ERC721IsApprovedForAllType:
		c := parsedQuery.ERC721IsApprovedForAll
		return qp.evmHandler.HandleERC721IsApprovedForAll(ctx, c.Caller, c.ContractAddress, c.Owner, c.Operator)
	case evmbindings.ERC721TotalSupplyType:
		c := parsedQuery.ERC721TotalSupply
		return qp.evmHandler.HandleERC721TotalSupply(ctx, c.Caller, c.ContractAddress)
	case evmbindings.ERC721NameSymbolType:
		c := parsedQuery.ERC721NameSymbol
		return qp.evmHandler.HandleERC721NameSymbol(ctx, c.Caller, c.ContractAddress)
	case evmbindings.ERC721UriType:
		c := parsedQuery.ERC721Uri
		return qp.evmHandler.HandleERC721Uri(ctx, c.Caller, c.ContractAddress, c.TokenID)
	case evmbindings.ERC721RoyaltyInfoType:
		c := parsedQuery.ERC721RoyaltyInfo
		return qp.evmHandler.HandleERC721RoyaltyInfo(ctx, c.Caller, c.ContractAddress, c.TokenID, c.SalePrice)
	case evmbindings.GetEvmAddressType:
		c := parsedQuery.GetEvmAddress
		return qp.evmHandler.HandleGetEvmAddress(ctx, c.KiiAddress)
	case evmbindings.GetKiiAddressType:
		c := parsedQuery.GetKiiAddress
		return qp.evmHandler.HandleGetKiiAddress(ctx, c.EvmAddress)
	case evmbindings.SupportsInterfaceType:
		c := parsedQuery.SupportsInterface
		return qp.evmHandler.HandleSupportsInterface(ctx, c.Caller, c.InterfaceID, c.ContractAddress)
	default:
		return nil, errors.New("unknown EVM query")
	}
}
