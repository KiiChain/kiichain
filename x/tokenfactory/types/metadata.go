package types

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const MaxDenomMetadataSize = 10 * 1024

func ValidateMetadataSize(metadata banktypes.Metadata) error {
	if size := metadata.Size(); size > MaxDenomMetadataSize {
		return ErrMetadataTooLarge.Wrapf("got %d bytes, max is %d bytes", size, MaxDenomMetadataSize)
	}
	return nil
}
