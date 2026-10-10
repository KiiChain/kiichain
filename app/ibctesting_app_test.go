package kiichain

import (
	ibctesting "github.com/cosmos/ibc-go/v11/testing"
)

// Keep this assertion in a test file. ibc-go's testing package imports the
// simapp, which registers a second ratelimit error set and prints that clash
// on every kiichaind start.
var _ ibctesting.TestingApp = (*KiichainApp)(nil)
