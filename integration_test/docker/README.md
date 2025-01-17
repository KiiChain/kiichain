## Prerequisite

### Install Docker and Docker Compose
MacOS:
```sh
# The easiest and recommended way to get Docker and
# Docker Compose is to install Docker Desktop here:
https://docs.docker.com/desktop/install/mac-install/
```

Ubuntu:
```sh
# Follow the below link to install docker on ubuntu
https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository
# Follow the below link to install standalone docker compose
https://docs.docker.com/compose/install/other/
```

## Local Cluster

Detailed instruction: see the `Makefile`.

**To start 4 node cluster**

This will start a 4 node kii chain cluster as well as having the oracle price feeder run for each node.
```sh
# If this is the first time or you want to rebuild the binary:
sudo make docker-cluster-start-integration
```
All the logs and genesis files will be generated under the temporary build/generated folder.

```sh
# To monitor logs after cluster is started
tail -f build/generated/logs/kiichaind-0.log
```

**To ssh into a single node**
```sh
# List all containers
docker ps -a
# SSH into a running container
docker exec -it [container_name] /bin/bash
```

## State Sync RPC Node

Requirement: Follow the above steps to start a 4 node docker cluster before starting any state sync node

```sh
# Be sure to start up a 4-node cluster before you start a state sync node
make docker-cluster-start-integration
# Wait for at least a few minutes till the latest block height exceed 500 (this can be changed via app.toml)
kiichaind status |jq
# Start up a state sync node
make run-rpc-node-skipbuild-integration
```

## Local Debugging & Testing
One of the fanciest thing of using docker is fast iteration. Here we support:
- Being able to make changes locally and start up the chain to see the immediate impact
- Being able to make changes to local dependency repo (Cosmo SDK/Tendermint) and start the chain with the latest changes without bumping or release any binary version


In order to make local debugging work, you can follow these steps:
```sh
# Clone your dependency repo and put them under the same path as kii-chain
cd kii-chain
cd ../
git clone https://github.com/kiichain/kii-tendermint.git
git clone https://github.com/kiichain/kii-cosmos.git

# Modify go.mod file to point to local repo, must use the exact same path as below:
go mod edit -replace github.com/cosmos/cosmos-sdk=../kii-cosmos
go mod edit -replace github.com/tendermint/tendermint=../kii-tendermint

# You are good to go now! Make changes as you wish to any of the dependency repo and run docker to test it out.
```
****
