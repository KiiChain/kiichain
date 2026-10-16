package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/ratelimit module sentinel errors
var (
	ErrRateLimitAlreadyExists = errorsmod.Register(ModuleName, 1,
		"ratelimit key duplicated")
	ErrRateLimitNotFound = errorsmod.Register(ModuleName, 2,
		"rate limit not found")
	ErrZeroChannelValue = errorsmod.Register(ModuleName, 3,
		"channel value is zero")
	ErrQuotaExceeded = errorsmod.Register(ModuleName, 4,
		"quota exceeded")
	ErrInvalidClientState = errorsmod.Register(ModuleName, 5,
		"unable to determine client state from channelId")
	ErrChannelNotFound = errorsmod.Register(ModuleName, 6,
		"channel does not exist")
	ErrDenomIsBlacklisted = errorsmod.Register(ModuleName, 7,
		"denom is blacklisted",
	)
	ErrInvalidChannelId = errorsmod.Register(ModuleName, 8,
		"invalid channel",
	)
	ErrAsyncPacketNotFound = errorsmod.Register(ModuleName, 9,
		"async packet not found",
	)
	ErrAsyncAckNil = errorsmod.Register(ModuleName, 10,
		"async ack is nil",
	)
	ErrWriteAcknowledgementWrapperNil = errorsmod.Register(ModuleName, 11,
		"write acknowledgement wrapper cannot be nil",
	)
	ErrChannelKeeperV2Nil = errorsmod.Register(ModuleName, 12,
		"channel keeper v2 cannot be nil",
	)
)
