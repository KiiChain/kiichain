package e2e

import (
	"fmt"
	"net/http"
)

// /*
// RestRegression tests the continuity of critical endpoints that node operators, block explorers, and ecosystem participants depend on.
// Test Node REST Endpoints:
// 1. http://host:1317/validatorsets/latest
// 2. http://host:1317/validatorsets/{height}
// 3. http://host:1317/blocks/latest
// 4. http://host:1317/blocks/{height}
// 5. http://host:1317/syncing
// 6. http://host:1317/node_info
// 7. http://host:1317/txs
// Test Module REST Endpoints
// 1. Bank total
// 2. Auth params
// 3. Distribution for Community Pool
// 4. Evidence
// 5. Gov proposals
// 6. Mint params
// 7. Slashing params
// 8. Staking params
// */
const (
	valSetLatestPath                    = "/cosmos/base/tendermint/v1beta1/validatorsets/latest"
	valSetHeightPath                    = "/cosmos/base/tendermint/v1beta1/validatorsets/1"
	blocksLatestPath                    = "/cosmos/base/tendermint/v1beta1/blocks/latest"
	blocksHeightPath                    = "/cosmos/base/tendermint/v1beta1/blocks/1"
	syncingPath                         = "/cosmos/base/tendermint/v1beta1/syncing"
	nodeInfoPath                        = "/cosmos/base/tendermint/v1beta1/node_info"
	transactionsPath                    = "/cosmos/tx/v1beta1/txs?query=tx.height=9999999999"
	bankTotalModuleQueryPath            = "/cosmos/bank/v1beta1/supply"
	authParamsModuleQueryPath           = "/cosmos/auth/v1beta1/params"
	distributionCommPoolModuleQueryPath = "/cosmos/distribution/v1beta1/community_pool"
	evidenceModuleQueryPath             = "/cosmos/evidence/v1beta1/evidence"
	govPropsModuleQueryPath             = "/cosmos/gov/v1beta1/proposals"
	slashingParamsModuleQueryPath       = "/cosmos/slashing/v1beta1/params"
	stakingParamsModuleQueryPath        = "/cosmos/staking/v1beta1/params"
	missingPath                         = "/missing_endpoint"
	localMinGasPriceQueryPath           = "/cosmos/base/node/v1beta1/config"

	// EVM endpoints
	evmBaseFee        = "/cosmos/evm/vm/v1/base_fee"
	evmParams         = "/cosmos/evm/vm/v1/params"
	evmConfig         = "/cosmos/evm/vm/v1/config"
	feeMarketParams   = "/cosmos/evm/feemarket/v1/params"
	feeMarketBaseFee  = "/cosmos/evm/feemarket/v1/base_fee"
	feeMarketBlockGas = "/cosmos/evm/feemarket/v1/block_gas"
	erc20Params       = "/cosmos/evm/erc20/v1/token_pairs"
	erc20TokenPairs   = "/cosmos/evm/erc20/v1/params"
)

func (s *IntegrationTestSuite) testRestInterfaces() {
	s.Run("test rest interfaces", func() {
		var (
			valIdx        = 0
			c             = s.chainA
			endpointURL   = fmt.Sprintf("http://%s", s.valResources[c.id][valIdx].GetHostPort("1317/tcp"))
			testEndpoints = []struct {
				Path           string
				ExpectedStatus int
			}{
				// Client Endpoints
				{nodeInfoPath, 200},
				{syncingPath, 200},
				{valSetLatestPath, 200},
				{valSetHeightPath, 200},
				{blocksLatestPath, 200},
				{blocksHeightPath, 200},
				{transactionsPath, 200},
				// Module Endpoints
				{bankTotalModuleQueryPath, 200},
				{authParamsModuleQueryPath, 200},
				{distributionCommPoolModuleQueryPath, 200},
				{evidenceModuleQueryPath, 200},
				{govPropsModuleQueryPath, 200},
				{slashingParamsModuleQueryPath, 200},
				{stakingParamsModuleQueryPath, 200},
				{missingPath, 501},
				{localMinGasPriceQueryPath, 200},
				// EVM endpoints
				{evmBaseFee, 200},
				{evmParams, 200},
				{evmConfig, 200},
				{feeMarketParams, 200},
				{feeMarketBaseFee, 200},
				{feeMarketBlockGas, 200},
				{erc20Params, 200},
				{erc20TokenPairs, 200},
			}
		)

		for _, endpoint := range testEndpoints {
			resp, err := http.Get(fmt.Sprintf("%s%s", endpointURL, endpoint.Path))
			s.NoError(err, fmt.Sprintf("failed to get endpoint: %s%s", endpointURL, endpoint.Path))

			_, err = readJSON(resp)
			s.NoError(err, fmt.Sprintf("failed to read body of endpoint: %s%s", endpointURL, endpoint.Path))

			s.EqualValues(resp.StatusCode, endpoint.ExpectedStatus, fmt.Sprintf("invalid status from endpoint: : %s%s", endpointURL, endpoint.Path))
		}
	})
}
