package types

type Query struct {
	HexToBech32 *HexToBech32 `json:"hex_to_bech32,omitempty"`
	Bech32ToHex *Bech32ToHex `json:"bech32_to_hex,omitempty"`
}

// HexToBech32 is a query to convert a hex address to a bech32 address
type HexToBech32 struct {
	Address string `json:"address"`
	Prefix  string `json:"prefix"`
}

// Bech32ToHex is a query to convert a bech32 address to a hex address
type Bech32ToHex struct {
	Address string `json:"address"`
}

// HexToBech32Response is the response for the hex to bech32 query
type HexToBech32Response struct {
	Address string `json:"address"`
}

// Bech32ToHexResponse is the response for the bech32 to hex query
type Bech32ToHexResponse struct {
	Address string `json:"address"`
}
