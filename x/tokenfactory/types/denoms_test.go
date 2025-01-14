package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	appparams "github.com/kiichain/kiichain3/app/params"
	"github.com/kiichain/kiichain3/x/tokenfactory/types"
)

func TestDecomposeDenoms(t *testing.T) {
	appparams.SetAddressPrefixes()
	for _, tc := range []struct {
		desc  string
		denom string
		valid bool
	}{
		{
			desc:  "empty is invalid",
			denom: "",
			valid: false,
		},
		{
			desc:  "normal",
			denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/bitcoin",
			valid: true,
		},
		{
			desc:  "multiple slashes in subdenom",
			denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/bitcoin/1",
			valid: true,
		},
		{
			desc:  "no subdenom",
			denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/",
			valid: true,
		},
		{
			desc:  "incorrect prefix",
			denom: "ibc/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/bitcoin",
			valid: false,
		},
		{
			desc:  "subdenom of only slashes",
			denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/////",
			valid: true,
		},
		{
			desc:  "too long name",
			denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			valid: false,
		},
		{
			desc:  "too long creator name",
			denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczsasdfasdfasdfasdfasdfasdfadfasdfasdfasdfasdfasdfas/bitcoin",
			valid: false,
		},
		{
			desc:  "empty subdenom",
			denom: "factory/kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs/",
			valid: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, _, err := types.DeconstructDenom(tc.denom)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetTokenDenom(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		creator  string
		subdenom string
		valid    bool
	}{
		{
			desc:     "normal",
			creator:  "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
			subdenom: "bitcoin",
			valid:    true,
		},
		{
			desc:     "multiple slashes in subdenom",
			creator:  "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
			subdenom: "bitcoin/1",
			valid:    true,
		},
		{
			desc:     "no subdenom",
			creator:  "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
			subdenom: "",
			valid:    true,
		},
		{
			desc:     "subdenom of only slashes",
			creator:  "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
			subdenom: "/////",
			valid:    true,
		},
		{
			desc:     "too long name",
			creator:  "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
			subdenom: "adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			valid:    false,
		},
		{
			desc:     "subdenom is exactly max length",
			creator:  "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
			subdenom: "bitcoinfsadfsdfeadfsafwefsefsefsdfsdafasefsf",
			valid:    true,
		},
		{
			desc:     "creator is exactly max length",
			creator:  "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczshjkljkljkljkljkljkljkljkljkljkljk",
			subdenom: "bitcoin",
			valid:    true,
		},
		{
			desc:     "empty subdenom",
			creator:  "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
			subdenom: "",
			valid:    true,
		},
		{
			desc:     "non standard UTF-8",
			creator:  "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
			subdenom: "\u2603",
			valid:    false,
		},
		{
			desc:     "non standard ASCII",
			creator:  "kii1y3pxq5dp900czh0mkudhjdqjq5m8cpmm4hvczs",
			subdenom: "\n\t",
			valid:    false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := types.GetTokenDenom(tc.creator, tc.subdenom)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
