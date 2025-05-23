package keeper

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
	"github.com/stretchr/testify/require"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simparams "github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	dbm "github.com/tendermint/tm-db"
)

const faucetAccountName = "faucet"

var (
	InitTokens   = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	InitialCoins = sdk.NewCoins(sdk.NewCoin(utils.MicroKiiDenom, InitTokens))
	kiiCoins     = sdk.NewCoins(sdk.NewCoin(utils.MicroKiiDenom, InitTokens.MulRaw(int64(len(Addrs)+2))))

	OracleDecPrecision = 8

	ValPubKeys = simapp.CreateTestPubKeys(7) // Return 7 public keys for testing

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
	// Create the KV store to each module (each one needs this to store its data)
	keyAccount := sdk.NewKVStoreKey(authTypes.StoreKey)
	keyBank := sdk.NewKVStoreKey(bankTypes.StoreKey)
	keyDist := sdk.NewKVStoreKey(distTypes.StoreKey)
	keyStaking := sdk.NewKVStoreKey(stakingTypes.StoreKey)
	keyOracle := sdk.NewKVStoreKey(types.StoreKey)
	keyParams := sdk.NewKVStoreKey(paramsTypes.StoreKey)

	memKeys := sdk.NewMemoryStoreKeys(types.MemStoreKey)          // Create the memory KV store for the oracle
	tKeyParams := sdk.NewTransientStoreKey(paramsTypes.TStoreKey) // create a KV store for temporal parameters

	db := dbm.NewMemDB()                                                                         // Create on memory DB
	ms := store.NewCommitMultiStore(db)                                                          // create the multistore to handle the KV stores
	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Now().UTC()}, false, log.NewNopLogger()) // Create new context
	encodingConfig := MakeEncodingConfig()
	appCodec, legacyAmino := encodingConfig.Marshaler, encodingConfig.Amino

	// mount each KVStore on the multistore (ms)
	ms.MountStoreWithDB(keyAccount, sdk.StoreTypeIAVL, db)                   // mount as Merkle trees type
	ms.MountStoreWithDB(keyBank, sdk.StoreTypeIAVL, db)                      // mount as Merkle trees type
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)                    // mount as Merkle trees type
	ms.MountStoreWithDB(tKeyParams, sdk.StoreTypeIAVL, db)                   // mount as Merkle trees type
	ms.MountStoreWithDB(keyOracle, sdk.StoreTypeIAVL, db)                    // mount as Merkle trees type
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)                   // mount as Merkle trees type
	ms.MountStoreWithDB(keyDist, sdk.StoreTypeIAVL, db)                      // mount as Merkle trees type
	ms.MountStoreWithDB(memKeys[types.MemStoreKey], sdk.StoreTypeMemory, db) // mount as temporal memory type

	require.NoError(t, ms.LoadLatestVersion()) // Test multistore doesn't returns error

	// Set permissions and blacklist for accounts
	blackListAddrs := map[string]bool{
		authTypes.FeeCollectorName:     true,
		stakingTypes.NotBondedPoolName: true,
		stakingTypes.BondedPoolName:    true,
		distTypes.ModuleName:           true,
		faucetAccountName:              true,
		types.ModuleName:               true,
	}

	// Define account's permissions
	maccPerms := map[string][]string{
		authTypes.FeeCollectorName:     {authTypes.Minter},
		stakingTypes.NotBondedPoolName: nil,
		stakingTypes.BondedPoolName:    {authTypes.Burner, authTypes.Staking},
		distTypes.ModuleName:           {authTypes.Burner, authTypes.Staking},
		faucetAccountName:              {authTypes.Minter},
		types.ModuleName:               {authTypes.Minter, authTypes.Burner},
	}

	// Init account, bank and staking keepers
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, keyParams, tKeyParams)
	accountKeeper := authkeeper.NewAccountKeeper(appCodec, keyAccount, paramsKeeper.Subspace(authTypes.ModuleName), authTypes.ProtoBaseAccount, maccPerms)
	bankKeeper := bankkeeper.NewBaseKeeper(appCodec, keyBank, accountKeeper, paramsKeeper.Subspace(bankTypes.ModuleName), blackListAddrs)

	// Set Staking module on my testing environment
	stakingKeeper := stakingkeeper.NewKeeper(appCodec, keyStaking, accountKeeper, bankKeeper, paramsKeeper.Subspace(stakingTypes.ModuleName))
	stakingParams := stakingTypes.DefaultParams()
	stakingParams.BondDenom = utils.MicroKiiDenom
	stakingKeeper.SetParams(ctx, stakingParams)

	// Set distribution module on my testing environment
	distKeeper := distkeeper.NewKeeper(appCodec, keyDist, paramsKeeper.Subspace(distTypes.ModuleName),
		accountKeeper, bankKeeper, stakingKeeper, authTypes.FeeCollectorName, blackListAddrs)

	distKeeper.SetFeePool(ctx, distTypes.InitialFeePool())
	distParams := distTypes.DefaultParams()
	distParams.CommunityTax = sdk.NewDecWithPrec(2, 2)        // 0.02
	distParams.BaseProposerReward = sdk.NewDecWithPrec(1, 2)  // 0.01
	distParams.BonusProposerReward = sdk.NewDecWithPrec(4, 2) // 0.04
	distKeeper.SetParams(ctx, distParams)                     // Assign new params on the module
	stakingKeeper.SetHooks(stakingTypes.NewMultiStakingHooks(distKeeper.Hooks()))

	// Create empty module accounts and assign permissions
	faucetAcc := authTypes.NewEmptyModuleAccount(faucetAccountName, authTypes.Minter, authTypes.Burner) // Account with the tokens
	feeCollectorAcc := authTypes.NewEmptyModuleAccount(authTypes.FeeCollectorName)                      // Create fee collector account
	notBondedPoolAcc := authTypes.NewEmptyModuleAccount(stakingTypes.NotBondedPoolName, authTypes.Burner, authTypes.Staking)
	bondPoolAcc := authTypes.NewEmptyModuleAccount(stakingTypes.BondedPoolName, authTypes.Burner, authTypes.Staking)
	distAcc := authTypes.NewEmptyModuleAccount(distTypes.ModuleName)
	oracleAcc := authTypes.NewEmptyModuleAccount(types.ModuleName, authTypes.Minter)

	// Assign accounts on the account keeper
	accountKeeper.SetModuleAccount(ctx, faucetAcc)
	accountKeeper.SetModuleAccount(ctx, feeCollectorAcc)
	accountKeeper.SetModuleAccount(ctx, notBondedPoolAcc)
	accountKeeper.SetModuleAccount(ctx, bondPoolAcc)
	accountKeeper.SetModuleAccount(ctx, distAcc)
	accountKeeper.SetModuleAccount(ctx, oracleAcc)

	// Create total supply of my testing env and mint on the faucetAcc
	totalSupply := kiiCoins
	err := bankKeeper.MintCoins(ctx, faucetAccountName, totalSupply)
	require.NoError(t, err) // Validate the operation

	// Verify the faucet balance
	balance := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(faucetAccountName), utils.MicroKiiDenom)
	require.True(t, balance.IsGTE(kiiCoins[0]), "Faucet account does not have enough funds")

	// Send some tokens to not_bonded_tokens_pool account
	err = bankKeeper.SendCoinsFromModuleToModule(ctx, faucetAccountName, stakingTypes.NotBondedPoolName, sdk.NewCoins(sdk.NewCoin(utils.MicroKiiDenom, sdk.NewInt(100))))
	require.NoError(t, err) // Validate the operation

	// Send initial funds to testing accounts
	for _, addr := range Addrs {
		accountKeeper.SetAccount(ctx, authTypes.NewBaseAccountWithAddress(addr)) // Ensure the account exists
		err := bankKeeper.SendCoinsFromModuleToAccount(ctx, faucetAccountName, addr, InitialCoins)
		require.NoError(t, err) // Validate the operation
	}

	// check the module account set
	addr := accountKeeper.GetModuleAddress(types.ModuleName)
	require.NotNil(t, addr, "Oracle account was not set")

	// Set Oracle module
	oracleKeeper := NewKeeper(appCodec, keyOracle, memKeys[types.MemStoreKey], paramsKeeper.Subspace(types.ModuleName),
		accountKeeper, bankKeeper, stakingKeeper, distTypes.ModuleName)

	oracleParams := types.DefaultParams()
	oracleKeeper.SetParams(ctx, oracleParams)

	// Set the desired denoms
	for _, denom := range oracleParams.Whitelist {
		oracleKeeper.SetVoteTarget(ctx, denom.Name)
	}

	return TestInput{
		Ctx:           ctx,
		Cdc:           legacyAmino,
		AccountKeeper: accountKeeper,
		BankKeeper:    bankKeeper,
		OracleKeeper:  oracleKeeper,
		StakingKeeper: stakingKeeper,
		DistKeeper:    distKeeper,
	}
}

// MakeEncodingConfig prepares the codification env
func MakeEncodingConfig() simparams.EncodingConfig {
	amino := codec.NewLegacyAmino() // create codificator (using amino)
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)     // create protobuf codificador
	txCfg := tx.NewTxConfig(marshaler, tx.DefaultSignModes) // create tx system to encode and decode txs

	// Register the standar types on amino and interfaceRegistry
	std.RegisterInterfaces(interfaceRegistry)
	std.RegisterLegacyAminoCodec(amino)

	// Register all basic modules on amino and interfaceRegistry
	ModuleBasics.RegisterLegacyAminoCodec(amino)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)

	// Register the Oracle module on amino and interfaceRegistry
	types.RegisterCodec(amino)
	types.RegisterInterfaces(interfaceRegistry)

	return simparams.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}

// NewTestMsgCreateValidator simulate the message used on create a validator
// this function should be used ONLY FOR TESTING
func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey cryptotypes.PubKey, amount sdk.Int) *stakingTypes.MsgCreateValidator {
	rate := sdk.NewDecWithPrec(5, 2)                           // 0.05
	selfDelegation := sdk.NewCoin(utils.MicroKiiDenom, amount) // Create kii coin
	commission := stakingTypes.NewCommissionRates(rate, rate, rate)
	msg, _ := stakingTypes.NewMsgCreateValidator(address, pubKey, selfDelegation, stakingTypes.Description{}, commission, sdk.OneInt()) // create a new MsgCreateValidator instance

	return msg
}
