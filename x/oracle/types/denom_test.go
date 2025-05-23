package types

import "testing"

type testStruct struct {
	name      string
	denomList DenomList
	denom     string
	want      bool
}

func TestDenomList(t *testing.T) {
	testData := []testStruct{

		{
			name: "denomination present",
			denomList: DenomList{
				{Name: "USD"},
				{Name: "EUR"},
				{Name: "INR"},
			},
			denom: "EUR",
			want:  true,
		},
		{
			name: "denomination absent",
			denomList: DenomList{
				{Name: "USD"},
				{Name: "EUR"},
				{Name: "INR"},
			},
			denom: "JPY",
			want:  false,
		},
		{
			name:      "empty list",
			denomList: DenomList{},
			denom:     "USD",
			want:      false,
		},
	}

	// Run testData
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got := test.denomList.Contains(test.denom)
			if got != test.want {
				t.Errorf("DenomList.Contains() = %v, want %v", got, test.want)
			}
		})
	}

}
