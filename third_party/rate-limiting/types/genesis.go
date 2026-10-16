package types

import (
	"errors"
	fmt "fmt"
	"strconv"
	"strings"
	time "time"

	errorsmod "cosmossdk.io/errors"
)

// Splits a pending packet of the form {channelId}/{sequenceNumber}/{denom} into
// the channel Id, sequence number, and denom respectively.
func ParsePendingPacketId(pendingPacketId string) (channelId string, sequence uint64, denom string, err error) {
	splits := strings.SplitN(pendingPacketId, "/", 3)
	if len(splits) != 3 {
		return "", 0, "", fmt.Errorf("invalid pending packet (%s), must be of form: {channelId}/{sequenceNumber}/{denom}", pendingPacketId)
	}
	channelId = splits[0]
	sequenceString := splits[1]
	denom = splits[2]
	if denom == "" {
		return "", 0, "", fmt.Errorf("invalid pending packet (%s), denom must be specified", pendingPacketId)
	}

	sequence, err = strconv.ParseUint(sequenceString, 10, 64)
	if err != nil {
		return "", 0, "", errorsmod.Wrapf(err, "unable to parse sequence number (%s) from pending packet, %s", sequenceString, err)
	}

	return channelId, sequence, denom, nil
}

// ValidatePendingPacketParts validates the string fields used by pending packet
// collection keys. Pending packet keys are stored as (channelId, denom,
// sequence), where channelId and denom are non-terminal string keys and cannot
// contain the collections string delimiter byte 0x00.
func ValidatePendingPacketParts(channelId, denom string) error {
	if err := validatePendingPacketChannelId(channelId); err != nil {
		return err
	}

	if denom == "" {
		return errors.New("pending packet denom must be specified")
	}
	if strings.ContainsRune(denom, '\x00') {
		return errors.New("pending packet denom cannot contain 0x00")
	}

	return nil
}

func validatePendingPacketChannelId(channelId string) error {
	if len(channelId) > PendingSendPacketChannelLength {
		return errorsmod.Wrapf(ErrInvalidChannelId, "channel %s with length %d is greater than the allowed length %d", channelId, len(channelId), PendingSendPacketChannelLength)
	}
	if strings.ContainsRune(channelId, '\x00') {
		return errorsmod.Wrapf(ErrInvalidChannelId, "channel ID %q cannot contain 0x00", channelId)
	}

	return nil
}

// IsLegacyPendingPacketId returns true if the pending packet ID is in the pre-denom
// genesis format. These IDs cannot be safely migrated to denom-scoped markers and
// are dropped on import, mirroring the store migration behavior.
func IsLegacyPendingPacketId(pendingPacketId string) bool {
	splits := strings.Split(pendingPacketId, "/")
	if len(splits) != 2 {
		return false
	}

	_, err := strconv.ParseUint(splits[1], 10, 64)
	if err != nil {
		return false
	}

	return validatePendingPacketChannelId(splits[0]) == nil
}

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                           DefaultParams(),
		RateLimits:                       []RateLimit{},
		WhitelistedAddressPairs:          []WhitelistedAddressPair{},
		BlacklistedDenoms:                []string{},
		PendingSendPacketSequenceNumbers: []string{},
		PendingRecvPacketSequenceNumbers: []string{},
		HourEpoch: HourEpoch{
			EpochNumber: 0,
			Duration:    time.Hour,
		},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate the format of the pending send packets
	for _, pendingPacketId := range gs.PendingSendPacketSequenceNumbers {
		if err := validatePendingPacketId(pendingPacketId); err != nil {
			return err
		}
	}
	for _, pendingPacketId := range gs.PendingRecvPacketSequenceNumbers {
		if err := validatePendingPacketId(pendingPacketId); err != nil {
			return err
		}
	}

	// Verify the epoch hour duration is specified
	if gs.HourEpoch.Duration == 0 {
		return errors.New("hour epoch duration must be specified")
	}

	// If the hour epoch has been initialized already (epoch number != 0), validate and then use it
	if gs.HourEpoch.EpochNumber > 0 {
		if gs.HourEpoch.EpochStartTime.Equal(time.Time{}) {
			return errors.New("if hour epoch number is non-empty, epoch time must be initialized")
		}
		if gs.HourEpoch.EpochStartHeight == 0 {
			return errors.New("if hour epoch number is non-empty, epoch height must be initialized")
		}
	}

	return nil
}

func validatePendingPacketId(pendingPacketId string) error {
	channelOrClientId, _, denom, err := ParsePendingPacketId(pendingPacketId)
	if err != nil {
		if IsLegacyPendingPacketId(pendingPacketId) {
			return nil
		}
		return err
	}

	return ValidatePendingPacketParts(channelOrClientId, denom)
}
