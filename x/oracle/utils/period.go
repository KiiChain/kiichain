package utils

const (
	BlocksPerMinute = uint64(75)
	BlocksPerHour   = BlocksPerMinute * 60
	BlocksPerDay    = BlocksPerMinute * 24
	BlocksPerWeek   = BlocksPerDay * 7
	BlocksPerMonth  = BlocksPerDay * 30
	BlocksPerYear   = BlocksPerDay * 365
)
