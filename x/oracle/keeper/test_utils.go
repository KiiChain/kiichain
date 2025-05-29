package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distribtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramsproptypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	erc20types "github.com/cosmos/evm/x/erc20/types"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	kiiparams "github.com/kiichain/kiichain/v1/app/params"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
	tokenfactorytypes "github.com/kiichain/kiichain/v1/x/tokenfactory/types"
)

const faucetAccountName = "faucet"

var (
	InitTokens   = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	InitialCoins = sdk.NewCoins(sdk.NewCoin(utils.MicroKiiDenom, InitTokens))
	kiiCoins     = sdk.NewCoins(sdk.NewCoin(utils.MicroKiiDenom, InitTokens.MulRaw(int64(len(Addrs)+2))))

	OracleDecPrecision = 8

	ValPubKeys = CreateTestPubKeys(7) // Return 7 public keys for testing

	pubKeys = []crypto.PubKey{
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
	}

	Addrs = []sdk.AccAddress{
		sdk.AccAddress(pubKeys[0].Address()),
		sdk.AccAddress(pubKeys[1].Address()),
		sdk.AccAddress(pubKeys[2].Address()),
		sdk.AccAddress(pubKeys[3].Address()),
		sdk.AccAddress(pubKeys[4].Address()),
		sdk.AccAddress(pubKeys[5].Address()),
		sdk.AccAddress(pubKeys[6].Address()),
	}

	ValAddrs = []sdk.ValAddress{
		sdk.ValAddress(pubKeys[0].Address()),
		sdk.ValAddress(pubKeys[1].Address()),
		sdk.ValAddress(pubKeys[2].Address()),
		sdk.ValAddress(pubKeys[3].Address()),
		sdk.ValAddress(pubKeys[4].Address()),
		sdk.ValAddress(pubKeys[5].Address()),
		sdk.ValAddress(pubKeys[6].Address()),
	}
)

// ModuleBasics register the basic app modules
var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	distribution.AppModuleBasic{},
	staking.AppModuleBasic{},
	params.AppModuleBasic{},
)

// TestInput nolint
type TestInput struct {
	Ctx           sdk.Context
	Cdc           *codec.LegacyAmino
	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
	OracleKeeper  Keeper
	StakingKeeper stakingkeeper.Keeper
	DistKeeper    distkeeper.Keeper
}

// CreateTestInput prepate the testing env, initializes modules, creates ctx,
// prepares memory storage and create testing accounts with funds
func CreateTestInput(t *testing.T) TestInput {
	t.Helper()
	// Start the keys
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		distribtypes.StoreKey,
		stakingtypes.StoreKey,
		paramsTypes.StoreKey,
		types.StoreKey,
		paramsTypes.TStoreKey,
	)

	authority := authtypes.NewModuleAddress("gov")

	cms := integration.CreateMultiStore(keys, log.NewTestLogger(t))                               // create the multistore to handle the KV stores
	ctx := sdk.NewContext(cms, tmproto.Header{Time: time.Now().UTC()}, false, log.NewNopLogger()) // Create new context

	encodingConfig := kiiparams.MakeEncodingConfig()
	// Register the auth interface
	banktypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	authtypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	authvesting.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	evidencetypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	cryptocodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	govv1types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	govv1beta1types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	paramsproptypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	paramsproptypes.RegisterLegacyAminoCodec(encodingConfig.Amino)

	upgradetypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	distribtypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ratelimittypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	tokenfactorytypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	// EVM register interfaces
	evmtypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	erc20types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	feemarkettypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	appCodec, legacyAmino := encodingConfig.Marshaler, encodingConfig.Amino

	// Set permissions and blacklist for accounts
	blackListAddrs := map[string]bool{
		authtypes.FeeCollectorName:     true,
		stakingtypes.NotBondedPoolName: true,
		stakingtypes.BondedPoolName:    true,
		distribtypes.ModuleName:        true,
		faucetAccountName:              true,
		types.ModuleName:               true,
	}

	// Define account's permissions
	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:     {authtypes.Minter},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		distribtypes.ModuleName:        {authtypes.Burner, authtypes.Staking},
		faucetAccountName:              {authtypes.Minter},
		types.ModuleName:               {authtypes.Minter, authtypes.Burner},
	}

	// Init account, bank and staking keepers
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authority.String(),
	)
	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blackListAddrs,
		authority.String(),
		log.NewNopLogger(),
	)

	// Set Staking module on my testing environment
	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		authority.String(),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = utils.MicroKiiDenom
	err := stakingKeeper.SetParams(ctx, stakingParams)
	require.NoError(t, err)

	// Set distribution module on my testing environment
	distKeeper := distkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distribtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authority.String(),
	)

	err = distKeeper.FeePool.Set(ctx, distribtypes.InitialFeePool())
	require.NoError(t, err)
	distParams := distribtypes.DefaultParams()
	distParams.CommunityTax = math.LegacyNewDecWithPrec(2, 2) // 0.02
	err = distKeeper.Params.Set(ctx, distParams)
	require.NoError(t, err)
	stakingKeeper.SetHooks(stakingtypes.NewMultiStakingHooks(distKeeper.Hooks()))

	// Create total supply of my testing env and mint on the faucetAcc
	totalSupply := kiiCoins
	err = bankKeeper.MintCoins(ctx, faucetAccountName, totalSupply)
	require.NoError(t, err) // Validate the operation

	// Verify the faucet balance
	balance := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(faucetAccountName), utils.MicroKiiDenom)
	require.True(t, balance.IsGTE(kiiCoins[0]), "Faucet account does not have enough funds")

	// Send some tokens to not_bonded_tokens_pool account
	err = bankKeeper.SendCoinsFromModuleToModule(ctx, faucetAccountName, stakingtypes.NotBondedPoolName, sdk.NewCoins(sdk.NewCoin(utils.MicroKiiDenom, math.NewInt(100))))
	require.NoError(t, err) // Validate the operation

	// Send initial funds to testing accounts
	for _, addr := range Addrs {
		// accountKeeper.SetAccount(ctx, authtypes.NewBaseAccountWithAddress(addr)) // Ensure the account exists
		err := bankKeeper.SendCoinsFromModuleToAccount(ctx, faucetAccountName, addr, InitialCoins)
		require.NoError(t, err) // Validate the operation
	}

	// check the module account set
	addr := accountKeeper.GetModuleAddress(types.ModuleName)
	require.NotNil(t, addr, "Oracle account was not set")

	// Set Oracle module
	oracleKeeper := NewKeeper(appCodec, runtime.NewKVStoreService(keys[types.StoreKey]),
		accountKeeper, bankKeeper, stakingKeeper, authority.String())

	oracleParams := types.DefaultParams()

	err = oracleKeeper.Params.Set(ctx, oracleParams)
	require.NoError(t, err)

	// Set the desired denoms
	for _, denom := range oracleParams.Whitelist {
		err = oracleKeeper.VoteTarget.Set(ctx, denom.Name, types.Denom{Name: denom.Name})
		require.NoError(t, err)
	}

	return TestInput{
		Ctx:           ctx,
		Cdc:           legacyAmino,
		AccountKeeper: accountKeeper,
		BankKeeper:    bankKeeper,
		OracleKeeper:  oracleKeeper,
		StakingKeeper: *stakingKeeper,
		DistKeeper:    distKeeper,
	}
}

// NewTestMsgCreateValidator simulate the message used on create a validator
// this function should be used ONLY FOR TESTING
func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey cryptotypes.PubKey, amount math.Int) *stakingtypes.MsgCreateValidator {
	// Get the validator rates (0.05)
	rate := math.LegacyNewDecWithPrec(5, 2)
	// Build the self delegation
	selfDelegation := sdk.NewCoin(utils.MicroKiiDenom, amount)
	// Build the commission rates
	commission := stakingtypes.NewCommissionRates(rate, rate, rate)

	// Get the message
	msg, _ := stakingtypes.NewMsgCreateValidator(address.String(), pubKey, selfDelegation, stakingtypes.Description{
		Moniker: fmt.Sprint("val-", address.String()),
	}, commission, math.OneInt()) // create a new MsgCreateValidator instance

	// Return the message
	return msg
}

// CreateTestPubKeys returns a total of numPubKeys public keys in ascending order.
func CreateTestPubKeys(numPubKeys int) []cryptotypes.PubKey {
	var publicKeys []cryptotypes.PubKey
	var buffer bytes.Buffer

	// start at 10 to avoid changing 1 to 01, 2 to 02, etc
	for i := 100; i < (numPubKeys + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AF") // base pubkey string
		buffer.WriteString(numString)                                                       // adding on final two digits to make pubkeys unique
		publicKeys = append(publicKeys, NewPubKeyFromHex(buffer.String()))
		buffer.Reset()
	}

	return publicKeys
}

// NewPubKeyFromHex returns a PubKey from a hex string.
func NewPubKeyFromHex(pk string) (res cryptotypes.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	if len(pkBytes) != ed25519.PubKeySize {
		panic(errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, "invalid pubkey size"))
	}
	return &ed25519.PubKey{Key: pkBytes}
}
