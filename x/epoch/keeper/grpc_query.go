package keeper

import (
	"github.com/kiichain/kiichain3/x/epoch/types"
)

var _ types.QueryServer = Keeper{}
