package types

// Query is the query type for the EVM module on wasmbindings
type Query struct {
	EthCall          *EthCallRequest          `json:"eth_call,omitempty"`
	ERC20Information *ERC20InformationRequest `json:"erc20_information,omitempty"`
	ERC20Balance     *ERC20BalanceRequest     `json:"erc20_balance,omitempty"`
	ERC20Allowance   *ERC20AllowanceRequest   `json:"erc20_allowance,omitempty"`
}

// EthCallRequest is the query type for the EthCallRequest query
type EthCallRequest struct {
	Contract string `json:"contract"`
	Data     string `json:"data"`
}

// EthCallResponse is the response type for the EthCall query
type EthCallResponse struct {
	Data string `json:"data"`
}

// ERC20InformationRequest is the query type for the ERC20Information query
type ERC20InformationRequest struct {
	Contract string `json:"contract"`
}

// ERC20InformationRequest is the request type for the ERC20Information query
type ERC20InformationResponse struct {
	Decimals    uint8  `json:"decimals"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	TotalSupply string `json:"total_supply"`
}

// ERC20BalanceRequest is the query type for the ERC20Balance query
type ERC20BalanceRequest struct {
	Contract string `json:"contract"`
	Address  string `json:"address"`
}

// ERC20BalanceResponse is the response type for the ERC20Balance query
type ERC20BalanceResponse struct {
	Balance string `json:"balance"`
}

// ERC20AllowanceRequest is the query type for the ERC20Allowance query
type ERC20AllowanceRequest struct {
	Contract string `json:"contract"`
	Owner    string `json:"owner"`
	Spender  string `json:"spender"`
}

// ERC20AllowanceResponse is the response type for the ERC20Allowance query
type ERC20AllowanceResponse struct {
	Allowance string `json:"allowance"`
}
