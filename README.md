# Kiichain

<div align="center">
  <img src="assets/kii.png" alt="Kiichain Logo" width="200"/>
</div>

<div align="center">

[![Project Status: Active](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#active)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/KiiChain/kiichain)](https://github.com/KiiChain/kiichain/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/KiiChain/kiichain)](https://goreportcard.com/report/github.com/KiiChain/kiichain)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Lines of Code](https://tokei.rs/b1/github/KiiChain/kiichain)](https://github.com/KiiChain/kiichain)
[![GitHub Super-Linter](https://github.com/KiiChain/kiichain/workflows/Lint/badge.svg)](https://github.com/marketplace/actions/super-linter)
[![GoDoc](https://godoc.org/github.com/KiiChain/kiichain?status.svg)](https://godoc.org/github.com/KiiChain/kiichain)

</div>

<div align="center">

[![Discord](https://img.shields.io/discord/123456789?color=7289da&label=Discord&logo=discord&logoColor=white)](https://discord.gg/kiichain)
[![Twitter Follow](https://img.shields.io/twitter/follow/kiichain?style=social)](https://x.com/KiiChainio)
[![TikTok](https://img.shields.io/badge/TikTok-@kiichain-ff0050?style=flat&logo=tiktok&logoColor=white)](https://www.tiktok.com/@kiichain_)

</div>

---

**KiiChain** is a CometBFT-based EVM-compatible blockchain providing a fast and scalable payment settlement layer for emerging market finance. As the first on-chain FX layer for stablecoins and RWA (Real World Assets), KiiChain is building the future of finance for emerging markets.

## 🌟 Features

- **100% EVM Compatible** - Leverage all EVM infrastructure and build with Solidity
- **High Performance** - Up to 12,000 TPS with ~1 second block times
- **Interoperable** - Connect with 100+ blockchain ecosystems
- **Custom Modules** - Built for RWA, PayFi, and CrediFi
- **Emerging Market Focus** - Gas fee scalability designed for micro-payments

## 🚀 Quick Links

### 🧪 Oro Testnet

- **[Join Oro Testnet](https://kiichain.io/testnet)** - Start validating on our testnet
- **[Testnet Explorer](https://explorer.kiichain.io)** - View transactions and blocks
- **[Testnet Faucet](https://explorer.kiichain.io/faucet)** - Get testnet tokens

### 📚 Documentation & Resources

- **[Official Documentation](https://docs.kiiglobal.io/docs)** - Comprehensive guides and API docs
- **[Developer Hub](https://docs.kiiglobal.io/docs/build-on-kiichain/developer-hub)** - Tools and resources for builders
- **[Whitepaper](https://docs.kiiglobal.io/docs/learn/kiichain/whitepaper)** - Technical specifications
- **[Blog](https://blog.kiiglobal.io/)** - Latest updates and insights

## 💻 Hardware Requirements

**Minimum**

- 8 GB RAM
- 1 TB NVME SSD
- 4 Cores (modern CPU's)

**Recommended**

- 16 GB RAM
- 2 TB NVME SSD
- 8 Cores (modern CPU's)

## 🔧 Operating System

> Linux (x86_64) or Linux (amd64) Recommended: Arch Linux or Ubuntu

### Dependencies

> **Prerequisite:** Go 1.23.6+ required (project uses Go 1.23.6 with toolchain 1.23.8)

**Install/Upgrade Go:**

- **macOS:** `brew install go@1.23` or download from [golang.org](https://golang.org/dl/)
- **Ubuntu:** `sudo snap install go --classic --channel=1.23/stable`
- **Arch Linux:** `pacman -S go`
- **Manual install:** Download from [golang.org/dl](https://golang.org/dl/)

**Verify Go version:**

```bash
go version  # Should show 1.23.6 or higher
```

**Configure Go PATH (important!):**

```bash
# Add Go bin to PATH (required for kiichaind to be found)
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc  # or restart terminal
```

> **Prerequisite:** git

- **Arch Linux:** `pacman -S git`
- **Ubuntu:** `sudo apt-get install git`

> **Optional:** GNU make

- **Arch Linux:** `pacman -S make`
- **Ubuntu:** `sudo apt-get install make`

## 🚀 Getting Started

### Installation

```bash
# Clone the repository
git clone https://github.com/KiiChain/kiichain.git
cd kiichain

# Build and install
make install

# Verify installation
kiichaind version
```

### Running a Node

Follow the official [Step-by-Step Guide](https://docs.kiiglobal.io/docs/validate-the-network/run-a-validator-full-node/step-by-step-guide) for joining Testnet Oro:

```bash
# Variables for Testnet Oro configuration
PERSISTENT_PEERS="5b6aa55124c0fd28e47d7da091a69973964a9fe1@uno.sentry.testnet.v3.kiivalidator.com:26656,5e6b283c8879e8d1b0866bda20949f9886aff967@dos.sentry.testnet.v3.kiivalidator.com:26656"
CHAIN_ID="oro_1336-1"
NODE_HOME=~/.kiichain
NODE_MONIKER=testnet_oro
GENESIS_URL=https://raw.githubusercontent.com/KiiChain/testnets/refs/heads/main/testnet_oro/genesis.json
MINIMUM_GAS_PRICES="1000000000akii"

# Initialize the chain
kiichaind init $NODE_MONIKER --chain-id $CHAIN_ID --home $NODE_HOME

# Configure persistent peers
sed -i.bak "s/^persistent_peers = \"\"/persistent_peers = \"$PERSISTENT_PEERS\"/" $NODE_HOME/config/config.toml

# Set minimum gas prices
sed -i -e "/minimum-gas-prices =/ s^= .*^= \"$MINIMUM_GAS_PRICES\"^" $NODE_HOME/config/app.toml

# Download official genesis file
curl -L $GENESIS_URL -o genesis.json
mv genesis.json $NODE_HOME/config/genesis.json

# Verify genesis file (optional but recommended)
sha256sum $NODE_HOME/config/genesis.json
# Expected: 2805ae1752dc8c3435afd6bdceea929b3bbd2883606f3f3589f4d62c99156d2d

# Start the node
kiichaind start --home $NODE_HOME
```

**📝 Configuration Files Overview:**

- **`~/.kiichain/config/app.toml`** - Application configuration (gas prices, API settings)
- **`~/.kiichain/config/config.toml`** - Node configuration (P2P, RPC, indexing)
- **`~/.kiichain/config/genesis.json`** - Genesis state (created during init)

**🔧 For production validator setup:**

```bash
# Create validator and user accounts
kiichaind keys add validator --keyring-backend test
kiichaind keys add user --keyring-backend test

# Add genesis accounts (for local testing)
kiichaind genesis add-genesis-account $(kiichaind keys show validator -a --keyring-backend test) 100000000000000000000akii
kiichaind genesis add-genesis-account $(kiichaind keys show user -a --keyring-backend test) 100000000000000000000akii

# Create genesis transaction
kiichaind genesis gentx validator 1000000000000000000akii --keyring-backend test --chain-id localchain_1010-1

# Collect genesis transactions
kiichaind genesis collect-gentxs

# Validate genesis file
kiichaind genesis validate-genesis
```

For detailed setup instructions, visit our [documentation](https://docs.kiiglobal.io/docs).

### 🔧 Troubleshooting

**Common Issues:**

- **"Minimum Go version 1.23 is required"** - Upgrade your Go installation to 1.23.6+
- **"unknown directive: toolchain"** - Your Go version is too old, upgrade to 1.23.6+
- **"invalid go version"** - Ensure you have Go 1.23.6+ installed and in your PATH
- **"command not found: kiichaind"** - Add `$HOME/go/bin` to your PATH (see setup above)
- **"set min gas price in app.toml"** - Use `1000000000akii` as shown in the setup above
- **"Wrong Block.Header.AppHash"** - Ensure you downloaded the correct genesis file for `oro_1336-1`
- **"failed to find any peers"** - Check that persistent peers are correctly configured

**Check your setup:**

```bash
go version                          # Should show 1.23.6+
echo $PATH | grep go                # Should include go/bin
which kiichaind                     # Should show path after 'make install'
kiichaind version                   # Should show version like v3.0.0-5-g239012d
```

**Quick fixes:**

```bash
export PATH="$HOME/go/bin:$PATH"                                                                              # Fix PATH issues (current session)
sed -i -e "/minimum-gas-prices =/ s^= .*^= \"1000000000akii\"^" ~/.kiichain/config/app.toml              # Fix gas price error
# On macOS, use curl instead of wget: curl -L <URL> -o filename
```

## 🤝 Contributing

All contributions are very welcome! Remember, contribution is not only PRs and code, but any help with docs or helping other developers solve their issues are very appreciated!

Read below to learn how you can take part in KiiChain:

### Code of Conduct

Please be sure to read and follow our [Code of Conduct](./CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

### Issues, Questions and Discussions

We use [GitHub Issues](https://github.com/KiiChain/kiichain/issues) for tracking requests and bugs, and for general questions and discussion.

## 🌍 Community

Join our vibrant community and stay connected:

- **[Discord](https://discord.gg/kiichain)** - Chat with developers and community
- **[X (Twitter)](https://x.com/KiiChainio)** - Follow for latest updates
- **[TikTok](https://www.tiktok.com/@kiichain_)** - Watch our latest content
- **[LinkedIn](https://www.linkedin.com/company/kiiglobal)** - Professional updates
- **[Instagram](https://www.instagram.com/kiichainofficial/)** - Behind the scenes

## 📜 License

The KiiChain is licensed under [Apache License 2.0](./LICENSE).

---

<div align="center">
  <strong>Building the future of finance for emerging markets 🌎</strong>
</div>
