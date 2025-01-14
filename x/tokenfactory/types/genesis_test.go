package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain3/x/tokenfactory/types"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "different admin from creator",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "kii1hjfwcza3e3uzeznf3qthhakdr9juetl7uajv0t",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "empty admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "no admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/bitcoin",
					},
				},
			},
			valid: true,
		},
		{
			desc: "invalid admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "moose",
						},
					},
				},
			},
			valid: false,
		},
		{
			desc: "multiple denoms",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
					{
						Denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/litecoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "duplicate denoms",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
					{
						Denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
