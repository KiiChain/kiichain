package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain/v3/x/rewards/types"
)

func TestParamsValidateBasic(t *testing.T) {
	type fields struct {
		GovernanceMinDeposit string
		TokenDenom           string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success - valid params",
			fields: fields{
				TokenDenom: "akii",
			},
			wantErr: false,
		},
		{
			name: "invalid - empty token denom",
			fields: fields{
				TokenDenom: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := types.Params{
				TokenDenom: tt.fields.TokenDenom,
			}
			if err := p.ValidateBasic(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultParams(t *testing.T) {
	// Test that default params are valid
	defaultParams := types.DefaultParams()
	require.NoError(t, defaultParams.ValidateBasic())

	// Verify specific default values
	require.Equal(t, "akii", defaultParams.TokenDenom)
}
