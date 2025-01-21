#!/bin/bash
set -e

teardown() {
    echo "Cleaning up resources..."
    make docker-cluster-stop-integration
}

# If any line fails, it will tear down docker
trap teardown EXIT

# This test works as the following
# 1. Preparation
# - We prepare by copying the old binary to the build path
# - The binary must be located on integration_test/upgrade/old_binary and named as kiichaind
# 2. Start-up
# - Start the docker environment
# - We do a chain upgrade on node 0, 1 and 2
# - We wait for a few moments and do the upgrade on node 3

# 1. Preparation
echo "Preparing the old binary"
mkdir -p build/
sudo cp integration_test/upgrade/old_binary/kiichaind build/

# 2. Start docker environment
echo "Starting the docker environment"
sudo rm build/generated/launch.complete
sudo make docker-cluster-start-skipbuild-integration > /dev/null 2>&1 &

# Wait for liveness
until [ $(cat build/generated/launch.complete |wc -l) = 4 ]
do
    echo "Containers are note initialized yet, sleeping..."
    sleep 1
done
echo "Nodes have started successfully. Sleeping for 10 seconds..."
sleep 10

# 2. Run the compatibility test
echo "Starting the compatibility test..."
python3 integration_test/scripts/runner.py integration_test/upgrade/upgrade_test.yaml

# HERE YOU CAN APPLY OTHER TESTS BEFORE THE TEAR DOWN
python3 integration_test/scripts/runner.py integration_test/authz_module/send_authorization_test.yaml
python3 integration_test/scripts/runner.py integration_test/bank_module/send_funds_test.yaml
python3 integration_test/scripts/runner.py integration_test/staking_module/staking_test.yaml

# Start the RPC node
# sudo make run-rpc-node-skipbuild-integration > /dev/null 2>&1 &
# integration_test/evm_module/scripts/evm_tests.sh
