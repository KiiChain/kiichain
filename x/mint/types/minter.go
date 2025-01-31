package types

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain3/utils/metrics"
	epochTypes "github.com/kiichain/kiichain3/x/epoch/types"
)

// NewMinter returns a new Minter object with the given inflation and annual
// provisions values.
func NewMinter(
	startDate string,
	endDate string,
	denom string,
	totalMintAmount uint64,
) Minter {
	return Minter{
		StartDate:           startDate,
		EndDate:             endDate,
		Denom:               denom,
		TotalMintAmount:     totalMintAmount,
		RemainingMintAmount: totalMintAmount,
		LastMintDate:        time.Time{}.Format(TokenReleaseDateFormat),
		LastMintHeight:      0,
		LastMintAmount:      0,
	}
}

// InitialMinter returns an initial Minter object with default values with no previous mints
func InitialMinter() Minter {
	return NewMinter(
		time.Time{}.Format(TokenReleaseDateFormat),
		time.Time{}.Format(TokenReleaseDateFormat),
		sdk.DefaultBondDenom,
		0,
	)
}

// DefaultInitialMinter returns a default initial Minter object for a new chain
// which uses an inflation rate of 0%.
func DefaultInitialMinter() Minter {
	return InitialMinter()
}

// validate minter
func ValidateMinter(minter Minter) error {
	if minter.GetTotalMintAmount() < minter.GetRemainingMintAmount() {
		return fmt.Errorf("total mint amount cannot be less than remaining mint amount")
	}

	// Get the end date
	endDate, err := minter.GetEndDateTime()
	if err != nil {
		return err
	}

	// Get the start date
	startDate, err := minter.GetStartDateTime()
	if err != nil {
		return err
	}

	// Compare the end and start date
	// End must be after start
	if endDate.Before(startDate) {
		return fmt.Errorf("end date must be after start date %s < %s", endDate, startDate)
	}
	return validateMintDenom(minter.Denom)
}

// GetLastMintDateTime returns the last time tokens were minted
func (m *Minter) GetLastMintDateTime() (time.Time, error) {
	lastMintedDateTime, err := time.Parse(TokenReleaseDateFormat, m.GetLastMintDate())
	if err != nil {
		// This should not happen as the date is validated when the minter is created
		return time.Time{}, fmt.Errorf("invalid end date for current minter: %s, minter=%s", err, m.String())
	}
	return lastMintedDateTime.UTC(), nil
}

// GetStartDateTime returns the minter last start date
func (m *Minter) GetStartDateTime() (time.Time, error) {
	startDateTime, err := time.Parse(TokenReleaseDateFormat, m.GetStartDate())
	if err != nil {
		// This should not happen as the date is validated when the minter is created
		return time.Time{}, fmt.Errorf("invalid start date for current minter: %s, minter=%s", err, m.String())
	}
	return startDateTime.UTC(), nil
}

// GetEndDateTime returns the minter last end date
func (m *Minter) GetEndDateTime() (time.Time, error) {
	endDateTime, err := time.Parse(TokenReleaseDateFormat, m.GetEndDate())
	if err != nil {
		// This should not happen as the date is validated when the minter is created
		return time.Time{}, fmt.Errorf("invalid end date for current minter: %s, minter=%s", err, m.String())
	}
	return endDateTime.UTC(), nil
}

// GetLastMintAmountCoin returns the minter last mint amount
func (m Minter) GetLastMintAmountCoin() sdk.Coin {
	return sdk.NewCoin(m.GetDenom(), sdk.NewInt(int64(m.GetLastMintAmount())))
}

// GetReleaseAmountToday takes the current time and returns the amount to be release as mint in coins
func (m *Minter) GetReleaseAmountToday(currentTime time.Time) (sdk.Coins, error) {
	// Get the amount to release today
	amountToRelease, err := m.getReleaseAmountToday(currentTime.UTC())
	if err != nil {
		return nil, err
	}

	// Transform into a coin and return
	amountToReleaseCoin := sdk.NewCoins(sdk.NewCoin(m.GetDenom(), sdk.NewInt(int64(amountToRelease))))
	return amountToReleaseCoin, nil
}

// RecordSuccessfulMint records a successful mint process
func (m *Minter) RecordSuccessfulMint(ctx sdk.Context, epoch epochTypes.Epoch, mintedAmount uint64) {
	m.RemainingMintAmount -= mintedAmount
	m.LastMintDate = epoch.CurrentEpochStartTime.Format(TokenReleaseDateFormat)
	m.LastMintHeight = uint64(epoch.CurrentEpochHeight)
	m.LastMintAmount = mintedAmount
	metrics.SetCoinsMinted(mintedAmount, m.GetDenom())
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeMint,
			sdk.NewAttribute(AttributeMintEpoch, fmt.Sprintf("%d", epoch.GetCurrentEpoch())),
			sdk.NewAttribute(AttribtueMintDate, m.GetLastMintDate()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, fmt.Sprintf("%d", mintedAmount)),
		),
	)
}

// getReleaseAmountToday returns the amount to the released as mint value on the day
func (m *Minter) getReleaseAmountToday(currentTime time.Time) (uint64, error) {
	// Get the start date
	startDate, err := m.GetStartDateTime()
	if err != nil {
		return 0, err
	}

	// Not yet started or already minted today
	if currentTime.Before(startDate) || currentTime.Format(TokenReleaseDateFormat) == m.GetLastMintDate() {
		return 0, nil
	}

	// Get the end date
	endDate, err := m.GetEndDateTime()
	if err != nil {
		return 0, err
	}

	// Get the number of days left
	numberOfDaysLeft, err := m.GetNumberOfDaysLeft(currentTime)
	if err != nil {
		return 0, err
	}

	// If it's already past the end date then release the remaining amount likely caused by outage
	if currentTime.After(endDate) || numberOfDaysLeft == 0 {
		return m.GetRemainingMintAmount(), nil
	}

	return m.GetRemainingMintAmount() / numberOfDaysLeft, nil
}

// GetNumberOfDaysLeft returns the number of days left before next mint
func (m *Minter) GetNumberOfDaysLeft(currentTime time.Time) (uint64, error) {
	// Get the end date
	endDate, err := m.GetEndDateTime()
	if err != nil {
		return 0, err
	}

	// If the last mint date is after the start date then use the last mint date as there's an ongoing release
	daysBetween := DaysBetween(currentTime, endDate)
	return daysBetween, nil
}

// OngoingRelease returns if there's a ongoing mint release
func (m *Minter) OngoingRelease() bool {
	return m.GetRemainingMintAmount() != 0
}

// DaysBetween returns days in between two dates
func DaysBetween(a, b time.Time) uint64 {
	// Convert both times to UTC before comparing
	aYear, aMonth, aDay := a.UTC().Date()
	a = time.Date(aYear, aMonth, aDay, 0, 0, 0, 0, time.UTC)
	bYear, bMonth, bDay := b.UTC().Date()
	b = time.Date(bYear, bMonth, bDay, 0, 0, 0, 0, time.UTC)

	// Always return a positive value between the dates
	if a.Before(b) {
		a, b = b, a
	}
	hours := a.Sub(b).Hours()
	return uint64(hours / 24)
}
