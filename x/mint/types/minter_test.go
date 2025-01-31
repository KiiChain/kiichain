package types_test

import (
	"fmt"
	"testing"
	"time"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain3/app"
	epochTypes "github.com/kiichain/kiichain3/x/epoch/types"
	"github.com/kiichain/kiichain3/x/mint/types"
	"github.com/stretchr/testify/require"
)

func TestParamsUkii(t *testing.T) {
	params := types.DefaultParams()
	err := params.Validate()
	require.Nil(t, err)

	params.MintDenom = "kii"
	err = params.Validate()
	require.NotNil(t, err)
}

func TestDaysBetween(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		date1    string
		date2    string
		expected uint64
	}{
		{
			name:     "Same day",
			date1:    "2023-04-20T00:00:00Z",
			date2:    "2023-04-20T23:59:59Z",
			expected: 0,
		},
		{
			name:     "25 days apart",
			date1:    "2023-04-24T00:00:00Z",
			date2:    "2023-05-19T00:00:00Z",
			expected: 25,
		},
		{
			name:     "One day apart",
			date1:    "2023-04-20T00:00:00Z",
			date2:    "2023-04-21T00:00:00Z",
			expected: 1,
		},
		{
			name:     "Five days apart",
			date1:    "2023-04-20T00:00:00Z",
			date2:    "2023-04-25T00:00:00Z",
			expected: 5,
		},
		{
			name:     "Inverted dates",
			date1:    "2023-04-25T00:00:00Z",
			date2:    "2023-04-20T00:00:00Z",
			expected: 5,
		},
		{
			name:     "Less than 24 hours apart, crossing day boundary",
			date1:    "2023-04-20T23:00:00Z",
			date2:    "2023-04-21T22:59:59Z",
			expected: 1,
		},
		{
			name:     "Exactly 24 hours apart",
			date1:    "2023-04-20T12:34:56Z",
			date2:    "2023-04-21T12:34:56Z",
			expected: 1,
		},
		{
			name:     "One minute less than 24 hours apart",
			date1:    "2023-04-20T12:34:56Z",
			date2:    "2023-04-21T12:33:56Z",
			expected: 1,
		},
		{
			name:     "Inverted dates with times",
			date1:    "2023-04-25T15:30:00Z",
			date2:    "2023-04-20T10:00:00Z",
			expected: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			date1, _ := time.Parse(time.RFC3339, tc.date1)
			date2, _ := time.Parse(time.RFC3339, tc.date2)

			result := types.DaysBetween(date1, date2)

			if result != tc.expected {
				t.Errorf("Expected days between %s and %s to be %d, but got %d", tc.date1, tc.date2, tc.expected, result)
			}
		})
	}
}

func TestGetReleaseAmountToday(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		minter         types.Minter
		currentTime    time.Time
		expectedAmount uint64
		errContains    string
	}{
		{
			name: "Regular scenario",
			minter: types.Minter{
				StartDate:           "2023-04-01",
				EndDate:             "2023-04-10",
				Denom:               "test",
				TotalMintAmount:     100,
				RemainingMintAmount: 60,
				LastMintDate:        "2023-04-04",
			},
			currentTime:    time.Date(2023, 4, 5, 0, 0, 0, 0, time.UTC),
			expectedAmount: 12,
		},
		{
			name: "Don't mint on the same day",
			minter: types.Minter{
				StartDate:           "2023-04-01",
				EndDate:             "2023-04-10",
				Denom:               "test",
				TotalMintAmount:     100,
				RemainingMintAmount: 60,
				LastMintDate:        "2023-04-05",
			},
			currentTime:    time.Date(2023, 4, 5, 0, 0, 0, 0, time.UTC),
			expectedAmount: 0,
		},
		{
			name: "No days left but remaining mint amount",
			minter: types.Minter{
				StartDate:           "2023-04-01",
				EndDate:             "2023-04-10",
				Denom:               "test",
				TotalMintAmount:     100,
				RemainingMintAmount: 60,
				LastMintDate:        "2023-04-11",
			},
			currentTime:    time.Date(2023, 4, 13, 0, 0, 0, 0, time.UTC),
			expectedAmount: 60,
		},
		{
			name: "Past end date",
			minter: types.Minter{
				StartDate:           "2023-04-01",
				EndDate:             "2023-04-10",
				Denom:               "test",
				TotalMintAmount:     100,
				RemainingMintAmount: 60,
				LastMintDate:        "2023-04-09",
			},
			currentTime:    time.Date(2023, 4, 11, 0, 0, 0, 0, time.UTC),
			expectedAmount: 60,
		},
		{
			name: "No remaining mint amount",
			minter: types.Minter{
				StartDate:           "2023-04-01",
				EndDate:             "2023-04-10",
				Denom:               "test",
				TotalMintAmount:     100,
				RemainingMintAmount: 0,
				LastMintDate:        "2023-04-05",
			},
			currentTime:    time.Date(2023, 4, 5, 0, 0, 0, 0, time.UTC),
			expectedAmount: 0,
		},
		{
			name: "Not yet started",
			minter: types.NewMinter(
				"2023-04-01",
				"2023-04-10",
				"test",
				100,
			),
			currentTime:    time.Date(2023, 4, 0, 0, 0, 0, 0, time.UTC),
			expectedAmount: 0,
		},
		{
			name: "First day",
			minter: types.NewMinter(
				"2023-04-01",
				"2023-04-10",
				"test",
				100,
			),
			currentTime:    time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
			expectedAmount: 11,
		},
		{
			name: "One day mint",
			minter: types.NewMinter(
				"2023-04-01",
				"2023-04-01",
				"test",
				100,
			),
			currentTime:    time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
			expectedAmount: 100,
		},
		{
			name: "One day mint - alreaddy minted",
			minter: types.Minter{
				StartDate:           "2023-04-01",
				EndDate:             "2023-04-01",
				Denom:               "test",
				TotalMintAmount:     100,
				RemainingMintAmount: 0,
				LastMintAmount:      100,
				LastMintDate:        "2023-04-01",
				LastMintHeight:      0,
			},
			currentTime:    time.Date(2023, 4, 1, 0, 1, 0, 0, time.UTC),
			expectedAmount: 0,
		},
		{
			name:           "No minter",
			minter:         types.InitialMinter(),
			currentTime:    time.Date(2023, 4, 5, 0, 0, 0, 0, time.UTC),
			expectedAmount: 0,
		},
		{
			name: "Invalid - bad start date",
			minter: types.NewMinter(
				"20-23-04-01", // Bad start date
				"2023-04-10",
				"test",
				100,
			),
			currentTime:    time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
			expectedAmount: 11,
			errContains:    "invalid start date for current minter",
		},
		{
			name: "Invalid - bad end date",
			minter: types.NewMinter(
				"2023-04-01",
				"20-23-04-10", // Bad end date
				"test",
				100,
			),
			currentTime:    time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
			expectedAmount: 11,
			errContains:    "invalid end date for current minter",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get the release amount
			releaseAmount, err := tc.minter.GetReleaseAmountToday(tc.currentTime.UTC())

			// Check for error
			if tc.errContains == "" {
				require.NoError(t, err, "Expected GetReleaseAmountToday to contain no error")
			} else {
				require.ErrorContains(
					t,
					err,
					tc.errContains,
					fmt.Sprintf("Expected GetReleaseAmountToday to contain error with %s", tc.errContains),
				)
				return
			}
			// Check the amount
			releaseAmountUint := releaseAmount.AmountOf(tc.minter.Denom).Uint64()
			if releaseAmountUint != tc.expectedAmount {
				t.Errorf("Expected release amount to be %d, but got %d", tc.expectedAmount, releaseAmountUint)
			}
		})
	}
}

func TestGetNumberOfDaysLeft(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		minter           types.Minter
		expectedDaysLeft uint64
		currentTime      time.Time
		errContains      string
	}{
		{
			name: "Regular scenario",
			minter: types.Minter{
				StartDate:           "2023-04-01",
				EndDate:             "2023-04-10",
				Denom:               "test",
				TotalMintAmount:     100,
				RemainingMintAmount: 60,
				LastMintDate:        "2023-04-05",
			},
			currentTime:      time.Date(2023, 4, 5, 0, 0, 0, 0, time.UTC),
			expectedDaysLeft: 5,
		},
		{
			name: "No days left but amount left",
			minter: types.Minter{
				StartDate:           "2023-04-01",
				EndDate:             "2023-04-10",
				Denom:               "test",
				TotalMintAmount:     100,
				RemainingMintAmount: 60,
				LastMintDate:        "2023-04-10",
			},
			currentTime:      time.Date(2023, 4, 10, 0, 0, 0, 0, time.UTC),
			expectedDaysLeft: 0,
		},
		{
			name: "Past end date",
			minter: types.Minter{
				StartDate:           "2023-04-01",
				EndDate:             "2023-04-10",
				Denom:               "test",
				TotalMintAmount:     100,
				RemainingMintAmount: 60,
				LastMintDate:        "2023-04-09",
			},
			currentTime:      time.Date(2023, 4, 9, 0, 0, 0, 0, time.UTC),
			expectedDaysLeft: 1,
		},
		{
			name: "Regular end date",
			minter: types.NewMinter(
				"2023-04-24",
				"2023-05-19",
				"test",
				100,
			),
			expectedDaysLeft: 25,
			currentTime:      time.Date(2023, 4, 24, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "No remaining mint amount",
			minter: types.Minter{
				StartDate:           "2023-04-01",
				EndDate:             "2023-04-10",
				Denom:               "test",
				TotalMintAmount:     100,
				RemainingMintAmount: 0,
				LastMintDate:        "2023-04-05",
			},
			currentTime:      time.Date(2023, 4, 5, 0, 0, 0, 0, time.UTC),
			expectedDaysLeft: 5,
		},
		{
			name: "First mint",
			minter: types.NewMinter(
				"2023-04-01",
				"2023-04-10",
				"test",
				100,
			),
			currentTime:      time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
			expectedDaysLeft: 9,
		},
		{
			name: "Invalid - Bad end date",
			minter: types.NewMinter(
				"2023-04-01",
				"20-23-04-10", // Broken date
				"test",
				100,
			),
			currentTime:      time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
			expectedDaysLeft: 0,
			errContains:      "invalid end date for current minter",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			daysLeft, err := tc.minter.GetNumberOfDaysLeft(tc.currentTime)

			// Check for error
			if tc.errContains == "" {
				require.NoError(t, err, "Expected GetNumberOfDaysLeft to contain no error")
			} else {
				require.ErrorContains(
					t,
					err,
					tc.errContains,
					fmt.Sprintf("Expected GetNumberOfDaysLeft to contain error with %s", tc.errContains),
				)
			}

			// Check the days left
			require.Equal(t, daysLeft, tc.expectedDaysLeft, fmt.Sprintf("Expected days left to be %d, but got %d", tc.expectedDaysLeft, daysLeft))
		})
	}
}

func TestNewMinter(t *testing.T) {
	m := types.NewMinter(
		time.Now().Format(types.TokenReleaseDateFormat),
		time.Now().AddDate(0, 0, 1).Format(types.TokenReleaseDateFormat),
		sdk.DefaultBondDenom,
		1000,
	)
	require.Equal(t, m.TotalMintAmount, m.RemainingMintAmount)
}

func TestInitialMinter(t *testing.T) {
	m := types.InitialMinter()
	require.Equal(t, uint64(0), m.TotalMintAmount)
	require.Equal(t, time.Time{}.Format(types.TokenReleaseDateFormat), m.StartDate)
	require.Equal(t, time.Time{}.Format(types.TokenReleaseDateFormat), m.EndDate)
}

func TestDefaultInitialMinter(t *testing.T) {
	m := types.DefaultInitialMinter()
	require.Equal(t, uint64(0), m.TotalMintAmount)
	require.Equal(t, time.Time{}.Format(types.TokenReleaseDateFormat), m.StartDate)
	require.Equal(t, time.Time{}.Format(types.TokenReleaseDateFormat), m.EndDate)
	require.False(t, m.OngoingRelease())
}

func TestValidateMinterBase(t *testing.T) {
	// Get the current test cases
	testCases := []struct {
		name           string
		minter         types.Minter
		ongoingRelease bool
		errContains    string
	}{
		{
			name:           "Good path - Ongoing release",
			ongoingRelease: true,
			minter: types.NewMinter(
				time.Now().Format(types.TokenReleaseDateFormat),
				time.Now().AddDate(0, 0, 1).Format(types.TokenReleaseDateFormat),
				sdk.DefaultBondDenom,
				1000,
			),
		},
		{
			name: "Bad path - Invalid denom",
			minter: types.NewMinter(
				time.Now().Format(types.TokenReleaseDateFormat),
				time.Now().AddDate(0, 0, 1).Format(types.TokenReleaseDateFormat),
				"invalid denom",
				1000,
			),
			errContains: "mint denom must be the same as the default bond denom",
		},
		{
			name: "Bad path - End date in the past",
			minter: types.NewMinter(
				time.Now().Format(types.TokenReleaseDateFormat),
				time.Now().AddDate(0, 0, -1).Format(types.TokenReleaseDateFormat),
				sdk.DefaultBondDenom,
				1000,
			),
			errContains: "end date must be after start date",
		},
		{
			name: "Bad path - Invalid initial date",
			minter: types.NewMinter(
				time.Now().Format("Random String"),
				time.Now().AddDate(0, 0, 1).Format(types.TokenReleaseDateFormat),
				sdk.DefaultBondDenom,
				1000,
			),
			errContains: "cannot parse \"Random String\"",
		},
		{
			name: "Bad path - Invalid end date",
			minter: types.NewMinter(
				time.Now().Format(types.TokenReleaseDateFormat),
				time.Now().AddDate(0, 0, 1).Format("Random String"),
				sdk.DefaultBondDenom,
				1000,
			),
			errContains: "cannot parse \"Random String\"",
		},
	}

	// Run all the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run the validate
			err := types.ValidateMinter(tc.minter)

			// Check for errors
			if tc.errContains == "" {
				require.NoError(t, err, "Expected minter to not contain errors")
				// Check for ongoing release
				require.Equal(
					t,
					tc.minter.OngoingRelease(),
					tc.ongoingRelease,
					fmt.Sprintf("Expected minter to have an ongoing release as %t", tc.ongoingRelease),
				)
			} else {
				require.ErrorContains(
					t,
					err,
					tc.errContains,
					fmt.Sprintf("Expected error to contain %s", tc.errContains),
				)
			}
		})
	}
}

func TestGetLastMintDateTime(t *testing.T) {
	m := types.InitialMinter()
	_, err := time.Parse(types.TokenReleaseDateFormat, m.GetLastMintDate())
	require.NoError(t, err)
}

func TestGetStartDateTime(t *testing.T) {
	m := types.InitialMinter()
	_, err := time.Parse(types.TokenReleaseDateFormat, m.GetStartDate())
	require.NoError(t, err)
}

func TestGetEndDateTime(t *testing.T) {
	m := types.InitialMinter()
	_, err := time.Parse(types.TokenReleaseDateFormat, m.GetEndDate())
	require.NoError(t, err)
}

func TestGetLastMintAmountCoin(t *testing.T) {
	m := types.InitialMinter()
	coin := m.GetLastMintAmountCoin()
	require.Equal(t, sdk.NewInt(int64(0)), coin.Amount)
	require.Equal(t, sdk.DefaultBondDenom, coin.Denom)
}

func TestRecordSuccessfulMint(t *testing.T) {
	minter := types.NewMinter(
		time.Now().Format(types.TokenReleaseDateFormat),
		time.Now().Add(time.Hour*24*10).Format(types.TokenReleaseDateFormat),
		sdk.DefaultBondDenom,
		1000,
	)
	app := app.Setup(false, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	currentTime := time.Now().UTC()

	epoch := epochTypes.Epoch{
		CurrentEpochStartTime: currentTime,
		CurrentEpochHeight:    100,
	}

	minter.RecordSuccessfulMint(ctx, epoch, 100)

	// Check results
	if minter.GetRemainingMintAmount() != 900 {
		t.Errorf("Remaining mint amount was incorrect, got: %d, want: %d.", minter.GetRemainingMintAmount(), 900)
	}
}

func TestValidateMinter(t *testing.T) {
	minter := types.NewMinter(
		time.Now().Format(types.TokenReleaseDateFormat),
		time.Now().Add(time.Hour*24*10).Format(types.TokenReleaseDateFormat),
		sdk.DefaultBondDenom,
		1000,
	)

	err := types.ValidateMinter(minter)
	if err != nil {
		t.Errorf("Expected valid minter, got error: %v", err)
	}

	// Create invalid minter
	minter = types.NewMinter(
		time.Now().Add(time.Hour*24*10).Format(types.TokenReleaseDateFormat), // start date is after end date
		time.Now().Format(types.TokenReleaseDateFormat),
		sdk.DefaultBondDenom,
		1000,
	)

	err = types.ValidateMinter(minter)
	if err == nil {
		t.Errorf("Expected error, got valid minter")
	}
}

func TestGetLastMintDateTimeBase(t *testing.T) {
	// Create minter object
	minter := types.NewMinter(
		time.Now().Format(types.TokenReleaseDateFormat),
		time.Now().Add(time.Hour*24*10).Format(types.TokenReleaseDateFormat),
		sdk.DefaultBondDenom,
		1000,
	)

	// Call the function
	date, err := minter.GetLastMintDateTime()
	require.NoError(t, err, "Expected GetLastMintDateTime to return no error")

	// Check the result
	// It should be the zero time because we haven't minted anything yet
	if !date.IsZero() {
		t.Errorf("Expected zero time, got: %v", date)
	}
}
