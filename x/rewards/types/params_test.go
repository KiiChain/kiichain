package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

func TestParamsValidateBasic(t *testing.T) {
	valid := types.DefaultParams()

	tests := []struct {
		name    string
		mutate  func(p *types.Params)
		wantErr bool
	}{
		{
			name:    "success - default params",
			mutate:  func(p *types.Params) {},
			wantErr: false,
		},
		{
			name: "invalid - empty token denom",
			mutate: func(p *types.Params) {
				p.TokenDenom = ""
			},
			wantErr: true,
		},
		{
			name: "invalid - goal bonded zero",
			mutate: func(p *types.Params) {
				p.GoalBonded = math.LegacyZeroDec()
			},
			wantErr: true,
		},
		{
			name: "invalid - goal bonded above one",
			mutate: func(p *types.Params) {
				p.GoalBonded = math.LegacyNewDecWithPrec(101, 2)
			},
			wantErr: true,
		},
		{
			name: "invalid - inflation min negative",
			mutate: func(p *types.Params) {
				p.InflationMin = math.LegacyNewDec(-1)
			},
			wantErr: true,
		},
		{
			name: "invalid - inflation max below min",
			mutate: func(p *types.Params) {
				p.InflationMin = math.LegacyNewDecWithPrec(10, 2)
				p.InflationMax = math.LegacyNewDecWithPrec(5, 2)
			},
			wantErr: true,
		},
		{
			name: "invalid - supply base negative",
			mutate: func(p *types.Params) {
				p.SupplyBase = math.NewInt(-1)
			},
			wantErr: true,
		},
		{
			name: "invalid - inflation rate change zero",
			mutate: func(p *types.Params) {
				p.InflationRateChange = math.LegacyZeroDec()
			},
			wantErr: true,
		},
		{
			name: "invalid - inflation rate change above one",
			mutate: func(p *types.Params) {
				p.InflationRateChange = math.LegacyNewDecWithPrec(101, 2)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := valid
			tt.mutate(&p)
			err := p.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultParams(t *testing.T) {
	defaultParams := types.DefaultParams()
	require.NoError(t, defaultParams.ValidateBasic())
	require.Equal(t, "akii", defaultParams.TokenDenom)
	require.True(t, defaultParams.SupplyBase.IsZero())
	require.True(t, math.LegacyNewDecWithPrec(67, 2).Equal(defaultParams.GoalBonded))
	require.True(t, math.LegacyNewDecWithPrec(20, 2).Equal(defaultParams.InflationMax))
	require.True(t, defaultParams.InflationMin.IsZero())
	require.True(t, math.LegacyNewDecWithPrec(13, 2).Equal(defaultParams.InflationRateChange))
}
