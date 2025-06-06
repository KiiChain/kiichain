package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain/v1/x/rewards/types"
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
				GovernanceMinDeposit: "1000000000000000000000", // 1000 kii
				TokenDenom:           "akii",
			},
			wantErr: false,
		},
		{
			name: "sucess - zero min deposit",
			fields: fields{
				GovernanceMinDeposit: "0",
				TokenDenom:           "akii",
			},
			wantErr: false,
		},
		{
			name: "invalid - empty token denom",
			fields: fields{
				GovernanceMinDeposit: "1000000000000000000000",
				TokenDenom:           "",
			},
			wantErr: true,
		},
		{
			name: "invalid - negative min deposit",
			fields: fields{
				GovernanceMinDeposit: "-1000000000000000000000",
				TokenDenom:           "akii",
			},
			wantErr: true,
		},
		{
			name: "invalid - non-numeric min deposit",
			fields: fields{
				GovernanceMinDeposit: "notanumber",
				TokenDenom:           "akii",
			},
			wantErr: true,
		},
		{
			name: "invalid - empty min deposit",
			fields: fields{
				GovernanceMinDeposit: "",
				TokenDenom:           "akii",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := types.Params{
				GovernanceMinDeposit: tt.fields.GovernanceMinDeposit,
				TokenDenom:           tt.fields.TokenDenom,
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
	require.Equal(t, "1000000000000000000000", defaultParams.GovernanceMinDeposit)
	require.Equal(t, "akii", defaultParams.TokenDenom)
}
