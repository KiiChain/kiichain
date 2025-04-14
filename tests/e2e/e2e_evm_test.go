package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (s *IntegrationTestSuite) testEVMQueries(jsonRCP string) {
	s.Run("eth_blockNumber", func() {
		res, err := httpEVMPostJSON(jsonRCP, "eth_blockNumber", []interface{}{})
		s.Require().NoError(err)

		blockNumber, err := parseResultAsHex(res)
		s.Require().NoError(err)
		s.Require().True(strings.HasPrefix(blockNumber, "0x"))
	})

	s.Run("eth_chainId", func() {
		res, err := httpEVMPostJSON(jsonRCP, "eth_chainId", []interface{}{})
		s.Require().NoError(err)

		chainID, err := parseResultAsHex(res)
		s.Require().NoError(err)
		s.Require().Equal(chainID, "0x3f2")
	})

	s.Run("eth_getBalance on zero address", func() {
		res, err := httpEVMPostJSON(jsonRCP, "eth_getBalance", []interface{}{
			"0x0000000000000000000000000000000000000000", "latest",
		})
		s.Require().NoError(err)

		balance, err := parseResultAsHex(res)
		s.Require().NoError(err)
		s.Require().True(strings.HasPrefix(balance, "0x0"))
	})

	s.Run("web3_clientVersion", func() {
		res, err := httpEVMPostJSON(jsonRCP, "web3_clientVersion", []interface{}{})
		s.Require().NoError(err)

		_, ok := res["result"].(string)
		s.Require().True(ok)
	})
}

// httpEVMPostJSON creates a post with the EVM format
func httpEVMPostJSON(url, method string, params []interface{}) (map[string]interface{}, error) {
	// Create the payload with the json format
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	}
	data, _ := json.Marshal(payload)

	// Get the response
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the result
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

// parseResultAsHex parse the result as json
func parseResultAsHex(resp map[string]interface{}) (string, error) {
	if result, ok := resp["result"].(string); ok {
		return result, nil
	}
	return "", fmt.Errorf("result not found or not a string")
}
