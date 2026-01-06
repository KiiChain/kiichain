package oracle

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestParseTwapsArgsValidation tests the validation in ParseGetTwapsArgs
func TestParseTwapsArgsValidation(t *testing.T) {
	testCases := []struct {
		name        string
		args        []interface{}
		expectError bool
		errContains string
	}{
		{
			name:        "valid - positive lookback period",
			args:        []interface{}{big.NewInt(100)},
			expectError: false,
		},
		{
			name:        "invalid - nil pointer causes panic without fix",
			args:        []interface{}{(*big.Int)(nil)},
			expectError: true,
			errContains: "invalid lookback period",
		},
		{
			name:        "invalid - zero lookback period should be rejected",
			args:        []interface{}{big.NewInt(0)},
			expectError: true,
			errContains: "lookback period must be positive",
		},
		{
			name:        "invalid - negative lookback period",
			args:        []interface{}{big.NewInt(-1)},
			expectError: true,
			errContains: "lookback period must be positive",
		},
		{
			name: "invalid - overflow value larger than uint64",
			args: []interface{}{new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil)},
			expectError: true,
			errContains: "lookback period overflow",
		},
		{
			name:        "invalid - wrong type (string instead of *big.Int)",
			args:        []interface{}{"100"},
			expectError: true,
			errContains: "invalid lookback period",
		},
		{
			name:        "invalid - wrong number of arguments (empty)",
			args:        []interface{}{},
			expectError: true,
			errContains: "invalid number of arguments",
		},
		{
			name:        "invalid - wrong number of arguments (too many)",
			args:        []interface{}{big.NewInt(100), big.NewInt(200)},
			expectError: true,
			errContains: "invalid number of arguments",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseGetTwapsArgs(tc.args)
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}
