package bindings

import "github.com/kiichain/kiichain3/x/epoch/types"

type KiiEpochQuery struct {
	// queries the current Epoch
	Epoch *types.QueryEpochRequest `json:"epoch,omitempty"`
}
