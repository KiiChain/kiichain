package types

import (
	"encoding/binary"

	"cosmossdk.io/collections"
)

const (
	ModuleName = "ratelimit"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

var (
	PathKeyPrefix      = KeyPrefix("path")
	RateLimitKeyPrefix = KeyPrefix("rate-limit")
	// PendingSendPacketPrefix is the legacy pending send packet prefix. It is
	// only used by migrations that clear old pending packet state.
	PendingSendPacketPrefix = KeyPrefix("pending-send-packet")
	// PendingReceivePacketPrefix is the legacy pending receive packet prefix. It
	// is only used by migrations that clear old pending packet state.
	PendingReceivePacketPrefix = KeyPrefix("pending-receive-packet")
	DenomBlacklistKeyPrefix    = KeyPrefix("denom-blacklist")
	AddressWhitelistKeyPrefix  = KeyPrefix("address-blacklist")
	HourEpochKey               = KeyPrefix("hour-epoch")

	PendingSendPacketsKey    = collections.NewPrefix(0)
	PendingReceivePacketsKey = collections.NewPrefix(1)

	PendingSendPacketChannelLength int = 64
)

// Get the rate limit byte key built from the denom and channelId
func GetRateLimitItemKey(denom string, channelId string) []byte {
	return append(KeyPrefix(denom), KeyPrefix(channelId)...)
}

// Get the pending packet key from the channel ID and sequence number
// The channel ID must be fixed length to allow for extracting the underlying
// values from a key
func GetPendingPacketKey(channelId string, sequenceNumber uint64) ([]byte, error) {
	if err := validatePendingPacketChannelId(channelId); err != nil {
		return nil, err
	}
	channelIdBz := make([]byte, PendingSendPacketChannelLength)
	copy(channelIdBz, channelId)

	sequenceNumberBz := make([]byte, 8)
	binary.BigEndian.PutUint64(sequenceNumberBz, sequenceNumber)

	return append(channelIdBz, sequenceNumberBz...), nil
}

// GetPendingSendPacketKey returns the pending packet key for a send packet.
// Deprecated: use GetPendingPacketKey instead.
func GetPendingSendPacketKey(channelId string, sequenceNumber uint64) ([]byte, error) {
	return GetPendingPacketKey(channelId, sequenceNumber)
}

// Get the whitelist path key from a sender and receiver address
func GetAddressWhitelistKey(sender, receiver string) []byte {
	return append(KeyPrefix(sender), KeyPrefix(receiver)...)
}
