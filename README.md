# Kiichain3

![Banner!](assets/kii.png)

Kiichain Version 3 - Forked from [Sei Chain](https://github.com/sei-protocol/sei-chain).

# Documentation

Our documentation can be found at:
- [Kiichain docs](https://docs.kiiglobal.io)

# Run Single Local Node (Docker)

```shell
git clone git@github.com:KiiChain/kiichain3.git
cd kiichain3
```

```shell
nano docker/rpcnode/scripts/step1_configure.init.sh

# replace REPLACE_SYNC_RPC with 18.227.13.176:26669

# replace REPLACE_SYNC_PEERS with 42355192eb77b71edbaa7e03f38e335849993ca0@18.227.13.176:26668,f3232ca5248aeb38af1d99542316d3c784dbf6f2@3.15.3.149:26668
```

```shell
make run-prime-node # this is to run a local node

# once the container is running, verify that you can reach the rpc
# by visiting http://localhost:26669 on your browser.  You should see a list of rpc endpoints

# execute commands once the container is running:
docker ps

# retrieve the container id
docker exec -it <container id> sh

# you will then shell into the container

kiichaind # you should now see a list of all the sub commands that can be used with kiichaind
```

# Run Testnet Validator Node

## Step 1: Run Single Testnet Node (Docker)

```shell
git clone git@github.com:KiiChain/kiichain3.git
cd kiichain3
make run-rpc-node # this is to run a rpc node

# once the container is running, verify that you can reach the rpc
# by visiting http://localhost:26669 on your browser.  You should see a list of rpc endpoints

# execute commands once the container is running:
docker ps

# retrieve the container id
docker exec -it <container id> sh

# you will then shell into the container

kiichaind # you should now see a list of all the sub commands that can be used with kiichaind
```

## Step 2: Create a Key

```shell

# ensure you are in the docker container from the previous step and have access to kiichaind

kiichaind keys add <KEY NAME>

# get your kiichain address for the faucet step
kiichaind keys show <KEY NAME> -a
```

## Step 3: Get Testnet Tokens

Go to the Kiichain discord channel and request for kiichain3 testnet tokens from the faucet.

## Step 4: Create a Validator Transaction

```shell

# ensure you are in the docker container from the previous step and have access to kiichaind

kiichaind tx staking create-validator \
--from <KEY NAME> \
--chain-id  \
--moniker="<VALIDATOR NAME>" \
--commission-max-change-rate=0.01 \
--commission-max-rate=1.0 \
--commission-rate=0.05 \
--details="<description>" \
--security-contact="<contact information>" \
--website="<your website>" \
--pubkey $(kiichaind tendermint show-validator) \
--min-self-delegation="1" \
--amount <token delegation>ukii \
--node localhost:26669
```

# Testnet Genesis

**How to validate on the Kiichain Testnet**

_This is the Kiichain kiichain3 Testnet_

> Genesis [Published](https://github.com/KiiChain/kiichain3/blob/main/remote/genesis.json)

## Hardware Requirements

**Minimum**

- 64 GB RAM
- 1 TB NVME SSD
- 16 Cores (modern CPU's)

## Operating System

> Linux (x86_64) or Linux (amd64) Recommended Arch Linux

**Dependencies**

> Prerequisite: go1.18+ required.

- Arch Linux: `pacman -S go`
- Ubuntu: `sudo snap install go --classic`

> Prerequisite: git.

- Arch Linux: `pacman -S git`
- Ubuntu: `sudo apt-get install git`

> Optional requirement: GNU make.

- Arch Linux: `pacman -S make`
- Ubuntu: `sudo apt-get install make`

## Upgrading node

When a new upgrade has launched and your node is requesting upgrade it to continue with the validation process, follow these steps:

### 1. Pull all changes

```
$ git pull origin main
```

### 2. Run upgrade command

We have created a command which interacts with the cosmovisor tool, installed in the docker file that helps to upgrade the blockchain. The most important part is to run the command with the parameters **UPGRADE_NAME** and **CONTAINER_NAME** where:

- **UPGRADE_NAME**: This is the title name of the upgrade proposal. It must be the name of the proposal plan if this is different, cosmoviso won't upgrade the node.

```
$ make upgrade UPGRADE_NAME=<upgradeName> CONTAINER_NAME=<yourContainerName>
```

For instance upgrading the blockchain to the v2.0.0, running in a container called **kiichain-rpc-node**:

```
$ make upgrade UPGRADE_NAME=v2.0.0 CONTAINER_NAME=kiichain-rpc-node
```

# Contributing

All contributions are very welcome! Remember, contribution is not only PRs and code, but any help with docs or helping other developers solve their issues are very appreciated!

Read below to learn how you can take part in the Kiichain.

### Code of Conduct

Please be sure to read and follow our [Code of Conduct][coc]. By participating, you are expected to uphold this code.

### Issues, Questions and Discussions

We use [GitHub Issues][issues] for tracking requests and bugs, and for general questions and discussion.

# License

The Kiichain is licensed under [Apache License 2.0][license].

[coc]: ./CODE_OF_CONDUCT.md
[issues]: https://github.com/KiiChain/kiichain3/issues
[license]: ./LICENSE
