# Kiichain3

![Banner!](assets/SeiLogo.png)

Kiichain Version 3 - Forked from Sei Chain

# Run Single Local Node
```shell
git clone git@github.com:KiiChain/kiichain3.git
cd kiichain3
make run-local-node # this is to run a local node

# once the container is running, verify that you can reach the rpc
# by visiting http://localhost:26669 on your browser.  You should see a list of rpc endpoints

# execute commands once the container is running:
docker ps

# retrieve the container id
docker exec <container id> -it sh

# you will then shell into the container

kiichaind # you should now see a list of all the sub commands that can be used with kiichaind
```

# Testnet
## Get started (TODO)
**How to validate on the Sei Testnet**
*This is the Sei Atlantic-2 Testnet ()*

> Genesis [Published](https://github.com/sei-protocol/testnet/blob/main/atlantic-2/genesis.json)

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
