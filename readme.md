# Kiichain3

![Banner!](assets/kii.png)

Kiichain Version 3 - Forked from Sei Chain

# Run Single Local Node (Docker)
```shell
git clone git@github.com:KiiChain/kiichain3.git
cd kiichain3
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

*This is the Kiichain kiichain3 Testnet*

> Genesis [Published](https://github.com/KiiChain/kiichain3/blob/main/remote/genesis.json)

## Hardware Requirements
**Minimum**
* 64 GB RAM
* 1 TB NVME SSD
* 16 Cores (modern CPU's)

## Operating System 

> Linux (x86_64) or Linux (amd64) Recommended Arch Linux

**Dependencies**
> Prerequisite: go1.18+ required.
* Arch Linux: `pacman -S go`
* Ubuntu: `sudo snap install go --classic`

> Prerequisite: git. 
* Arch Linux: `pacman -S git`
* Ubuntu: `sudo apt-get install git`

> Optional requirement: GNU make. 
* Arch Linux: `pacman -S make`
* Ubuntu: `sudo apt-get install make`
