package keeper

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"testing"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	kiiparams "github.com/kiichain/kiichain/v1/app/params"
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

	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"cosmossdk.io/log"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	dbm "github.com/cosmos/cosmos-db"
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
	// Create the KV store to each module (each one needs this to store its data)
	keyAccount := storetypes.NewKVStoreKey(authTypes.StoreKey)
	keyBank := storetypes.NewKVStoreKey(bankTypes.StoreKey)
	keyDist := storetypes.NewKVStoreKey(distTypes.StoreKey)
	keyStaking := storetypes.NewKVStoreKey(stakingTypes.StoreKey)
	keyOracle := storetypes.NewKVStoreKey(types.StoreKey)
	keyParams := storetypes.NewKVStoreKey(paramsTypes.StoreKey)

	keys := storetypes.NewKVStoreKeys(authTypes.StoreKey, bankTypes.StoreKey, distTypes.StoreKey, stakingTypes.StoreKey)

	authority := authtypes.NewModuleAddress("gov")

	memKeys := storetypes.NewMemoryStoreKeys(types.MemStoreKey)          // Create the memory KV store for the oracle
	tKeyParams := storetypes.NewTransientStoreKey(paramsTypes.TStoreKey) // create a KV store for temporal parameters

	db := dbm.NewMemDB()                                                                         // Create on memory DB
	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())            // create the multistore to handle the KV stores
	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Now().UTC()}, false, log.NewNopLogger()) // Create new context
	encodingConfig := kiiparams.MakeEncodingConfig()
	appCodec, legacyAmino := encodingConfig.Marshaler, encodingConfig.Amino

	// mount each KVStore on the multistore (ms)
	ms.MountStoreWithDB(keyAccount, storetypes.StoreTypeIAVL, db)                   // mount as Merkle trees type
	ms.MountStoreWithDB(keyBank, storetypes.StoreTypeIAVL, db)                      // mount as Merkle trees type
	ms.MountStoreWithDB(keyParams, storetypes.StoreTypeIAVL, db)                    // mount as Merkle trees type
	ms.MountStoreWithDB(tKeyParams, storetypes.StoreTypeIAVL, db)                   // mount as Merkle trees type
	ms.MountStoreWithDB(keyOracle, storetypes.StoreTypeIAVL, db)                    // mount as Merkle trees type
	ms.MountStoreWithDB(keyStaking, storetypes.StoreTypeIAVL, db)                   // mount as Merkle trees type
	ms.MountStoreWithDB(keyDist, storetypes.StoreTypeIAVL, db)                      // mount as Merkle trees type
	ms.MountStoreWithDB(memKeys[types.MemStoreKey], storetypes.StoreTypeMemory, db) // mount as temporal memory type

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
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)
	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[bankTypes.StoreKey]),
		accountKeeper,
		blackListAddrs,
		authority.String(),
		log.NewNopLogger(),
	)

	// Set Staking module on my testing environment
	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingTypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		authority.String(),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)
	stakingParams := stakingTypes.DefaultParams()
	stakingParams.BondDenom = utils.MicroKiiDenom
	stakingKeeper.SetParams(ctx, stakingParams)

	// Set distribution module on my testing environment
	distKeeper := distkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distTypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		authTypes.FeeCollectorName,
		authority.String(),
	)

	distKeeper.FeePool.Set(ctx, distTypes.InitialFeePool())
	distParams := distTypes.DefaultParams()
	distParams.CommunityTax = math.LegacyNewDecWithPrec(2, 2)        // 0.02
	distParams.BaseProposerReward = math.LegacyNewDecWithPrec(1, 2)  // 0.01
	distParams.BonusProposerReward = math.LegacyNewDecWithPrec(4, 2) // 0.04
	distKeeper.Params.Set(ctx, distParams)
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
	err = bankKeeper.SendCoinsFromModuleToModule(ctx, faucetAccountName, stakingTypes.NotBondedPoolName, sdk.NewCoins(sdk.NewCoin(utils.MicroKiiDenom, math.NewInt(100))))
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
		StakingKeeper: *stakingKeeper,
		DistKeeper:    distKeeper,
	}
}

// NewTestMsgCreateValidator simulate the message used on create a validator
// this function should be used ONLY FOR TESTING
func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey cryptotypes.PubKey, amount math.Int) *stakingTypes.MsgCreateValidator {
	rate := math.LegacyNewDecWithPrec(5, 2)                    // 0.05
	selfDelegation := sdk.NewCoin(utils.MicroKiiDenom, amount) // Create kii coin
	commission := stakingTypes.NewCommissionRates(rate, rate, rate)
	msg, _ := stakingTypes.NewMsgCreateValidator(address.String(), pubKey, selfDelegation, stakingTypes.Description{}, commission, math.OneInt()) // create a new MsgCreateValidator instance

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
