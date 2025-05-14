package types

// Query is the query type for the EVM module on wasmbindings
type Query struct {
	EthCall *EthCall `json:"eth_call,omitempty"`
}

// EthCall is the query type for the EthCall query
type EthCall struct {
	Contract string `json:"contract"`
	Data     string `json:"data"`
}

// EthCallResponse is the response type for the EthCall query
type EthCallResponse struct {
	Data string `json:"data"`
}
