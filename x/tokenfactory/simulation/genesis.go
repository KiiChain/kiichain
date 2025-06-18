package simulation

import (
	"math/rand"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	appparams "github.com/kiichain/kiichain/v2/app/params"
	"github.com/kiichain/kiichain/v2/x/tokenfactory/types"
)

func RandDenomCreationFeeParam(r *rand.Rand) sdk.Coins {
	amount := r.Int63n(10_000_000)
	return sdk.NewCoins(sdk.NewCoin(appparams.DisplayDenom, sdkmath.NewInt(amount)))
}

func RandomizedGenState(simstate *module.SimulationState) {
	tfGenesis := types.DefaultGenesis()

	_, err := simstate.Cdc.MarshalJSON(tfGenesis)
	if err != nil {
		panic(err)
	}

	simstate.GenState[types.ModuleName] = simstate.Cdc.MustMarshalJSON(tfGenesis)
}
