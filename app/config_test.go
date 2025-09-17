package kiichain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	kiichain "github.com/kiichain/kiichain/v4/app"
)

func TestParseChainID(t *testing.T) {
	tests := []struct {
		name        string
		chainID     string
		expected    uint64
		expectError bool
		errContains string
	}{
		// Valid cases
		{
			name:     "valid localchain",
			chainID:  "localchain_1010-1",
			expected: 1010,
		},
		{
			name:     "valid testnet",
			chainID:  "testnet_1336-1",
			expected: 1336,
		},
		{
			name:     "valid testing with large number",
			chainID:  "testing_262144-155",
			expected: 262144,
		},
		{
			name:     "valid with minimum values",
			chainID:  "a_1-1",
			expected: 1,
		},
		{
			name:     "valid with maximum uint64 compatible values",
			chainID:  "chainname_18446744073709551615-1",
			expected: 18446744073709551615,
		},

		// Invalid cases - format errors
		{
			name:        "empty string",
			chainID:     "",
			expectError: true,
			errContains: "does not match Evmos format",
		},
		{
			name:        "missing separator",
			chainID:     "localchain1010-1",
			expectError: true,
			errContains: "does not match Evmos format",
		},
		{
			name:        "wrong separator order",
			chainID:     "localchain-1010_1",
			expectError: true,
			errContains: "does not match Evmos format",
		},
		{
			name:        "missing EIP155 part",
			chainID:     "localchain_-1",
			expectError: true,
			errContains: "does not match Evmos format",
		},
		{
			name:        "missing epoch part",
			chainID:     "localchain_1010-",
			expectError: true,
			errContains: "does not match Evmos format",
		},
		{
			name:        "missing chain name",
			chainID:     "_1010-1",
			expectError: true,
			errContains: "does not match Evmos format",
		},
		{
			name:        "too long chain ID",
			chainID:     "thischainidiswaytoolongandexceedsfortyeightchars_1010-1",
			expectError: true,
			errContains: "cannot exceed 48 chars",
		},

		// Invalid cases - EIP155 number errors
		{
			name:        "EIP155 starts with zero",
			chainID:     "chain_0123-1",
			expectError: true,
			errContains: "does not match Evmos format",
		},
		{
			name:        "EIP155 contains non-digit",
			chainID:     "chain_12a34-1",
			expectError: true,
			errContains: "does not match Evmos format",
		},
		{
			name:        "EIP155 zero value",
			chainID:     "chain_0-1",
			expectError: true,
			errContains: "does not match Evmos format",
		},
		{
			name:        "EIP155 negative number",
			chainID:     "chain_-123-1",
			expectError: true,
			errContains: "does not match Evmos format",
		},

		// Edge cases
		{
			name:     "whitespace around",
			chainID:  "  localchain_1010-1  ",
			expected: 1010,
		},
		{
			name:        "mixed case chain name",
			chainID:     "LocalChain_1010-1",
			expectError: true,
			errContains: "does not match Evmos format",
		},
		{
			name:        "uppercase chain name",
			chainID:     "LOCALCHAIN_1010-1",
			expectError: true,
			errContains: "does not match Evmos format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := kiichain.ParseChainID(tt.chainID)

			if tt.expectError {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
