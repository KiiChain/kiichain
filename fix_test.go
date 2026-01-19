	// Fix: use proper 20-byte address instead of string
	nonExistentValidator := sdk.ValAddress(make([]byte, 20))