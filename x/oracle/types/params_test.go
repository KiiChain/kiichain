package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestParamsValid(t *testing.T) {
	p1 := DefaultParams()
	err := p1.Validate()
	require.NoError(t, err)

	// minus vote period
	p1.VotePeriod = 0
	err = p1.Validate()
	require.Error(t, err)

	// small vote threshold
	p2 := DefaultParams()
	p2.VoteThreshold = math.LegacyZeroDec()
	err = p2.Validate()
	require.Error(t, err)

	// negative reward band
	p3 := DefaultParams()
	p3.RewardBand = math.LegacyNewDecWithPrec(-1, 2)
	err = p3.Validate()
	require.Error(t, err)

	// negative slash fraction
	p4 := DefaultParams()
	p4.SlashFraction = math.LegacyNewDec(-1)
	err = p4.Validate()
	require.Error(t, err)

	// negative min valid per window
	p5 := DefaultParams()
	p5.MinValidPerWindow = math.LegacyNewDec(-1)
	err = p5.Validate()
	require.Error(t, err)

	// small slash window
	p6 := DefaultParams()
	p6.SlashWindow = 0
	err = p6.Validate()
	require.Error(t, err)

	// empty name
	p7 := DefaultParams()
	p7.Whitelist[0].Name = ""
	err = p7.Validate()
	require.Error(t, err)

	// slash window not divisible
	p8 := DefaultParams()
	p8.SlashWindow = 2
	p8.VotePeriod = 3
	err = p8.Validate()
	require.Error(t, err)

	p9 := DefaultParams()
	require.NotNil(t, p9.String())
}

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	require.Equal(t, DefaultSlashFraction, params.SlashFraction)
	require.Equal(t, DefaultLookbackDuration, params.LookbackDuration)
}
