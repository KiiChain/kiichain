package utils

import (
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v3/app/keepers"
)

// InstallNewPrecompiles is a placeholder for installing new precompiles.
func InstallNewPrecompiles(ctx sdk.Context, keepers *keepers.AppKeepers, precompiles []common.Address) error {
	// Log the upgrade
	ctx.Logger().Info("Installing new precompile...")

	// Install the new address
	return keepers.EVMKeeper.EnableStaticPrecompiles(ctx, precompiles...)
}
