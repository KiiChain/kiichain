package bech32_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain/v2/wasmbinding/bech32"
	bech32bindingtypes "github.com/kiichain/kiichain/v2/wasmbinding/bech32/types"
)

// TestHexToBech32 tests the HexToBech32 function of the bech32 module
func TestHexToBech32(t *testing.T) {
	testCases := []struct {
		name        string
		request     bech32bindingtypes.HexToBech32
		expResponse *bech32bindingtypes.HexToBech32Response
		errContains string
	}{
		{
			name: "valid conversion",
			request: bech32bindingtypes.HexToBech32{
				Address: "0x7cB61D4117AE31a12E393a1Cfa3BaC666481D02E",
				Prefix:  "kii",
			},
			expResponse: &bech32bindingtypes.HexToBech32Response{
				Address: "kii10jmp6sgh4cc6zt3e8gw05wavvejgr5pwfe2u6n",
			},
		},
		{
			name: "valid conversion 2",
			request: bech32bindingtypes.HexToBech32{
				Address: "0xA18344d76Cf42dB408db7f27d1167BaeBeDFe1Ee",
				Prefix:  "kii",
			},
			expResponse: &bech32bindingtypes.HexToBech32Response{
				Address: "kii15xp5f4mv7skmgzxm0unaz9nm46ldlc0w93d8qa",
			},
		},
		{
			name: "valid conversion - short address",
			request: bech32bindingtypes.HexToBech32{
				Address: "0xA18344d76",
				Prefix:  "kii",
			},
			expResponse: &bech32bindingtypes.HexToBech32Response{
				Address: "kii1qqqqqqqqqqqqqqqqqqqqqqqqpgvrgntkmjtsz8",
			},
		},
		{
			name: "valid conversion - empty address",
			request: bech32bindingtypes.HexToBech32{
				Address: "",
				Prefix:  "kii",
			},
			expResponse: &bech32bindingtypes.HexToBech32Response{
				Address: "kii1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq2g643z",
			},
		},
		{
			name: "valid conversion - cosmos address",
			request: bech32bindingtypes.HexToBech32{
				Address: "0x7cB61D4117AE31a12E393a1Cfa3BaC666481D02E",
				Prefix:  "cosmos",
			},
			expResponse: &bech32bindingtypes.HexToBech32Response{
				Address: "cosmos10jmp6sgh4cc6zt3e8gw05wavvejgr5pwsjskvv",
			},
		},
		{
			name: "invalid conversion - empty prefix",
			request: bech32bindingtypes.HexToBech32{
				Address: "0x7cB61D4117AE31a12E393a1Cfa3BaC666481D02E",
				Prefix:  "",
			},
			errContains: "prefix cannot be empty",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Apply the query
			res, err := bech32.HandleHexToBech32(tc.request)

			// Check the error
			if tc.errContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expResponse, res)
			}
		})
	}
}

// TestBech32ToHex tests the Bech32ToHex function of the bech32 module
func TestBech32ToHex(t *testing.T) {
	testCases := []struct {
		name        string
		request     bech32bindingtypes.Bech32ToHex
		expResponse *bech32bindingtypes.Bech32ToHexResponse
		errContains string
	}{
		{
			name: "valid conversion",
			request: bech32bindingtypes.Bech32ToHex{
				Address: "kii10jmp6sgh4cc6zt3e8gw05wavvejgr5pwfe2u6n",
			},
			expResponse: &bech32bindingtypes.Bech32ToHexResponse{
				Address: "0x7cB61D4117AE31a12E393a1Cfa3BaC666481D02E",
			},
		},
		{
			name: "invalid - invalid bech32 address",
			request: bech32bindingtypes.Bech32ToHex{
				Address: "randomaddress",
			},
			errContains: "invalid bech32 address",
		},
		{
			name: "invalid - invalid bech32 address",
			request: bech32bindingtypes.Bech32ToHex{
				Address: "kiiiiiiiiiiiiiiiiiiiiiiiiiiiiii10jmp6sgh4cc6zt3e8gw05wavvejgr5pwfe2u6n",
			},
			errContains: "decoding bech32 failed",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Apply the query
			res, err := bech32.HandleBech32ToHex(tc.request)

			// Check the error
			if tc.errContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expResponse, res)
			}
		})
	}
}
