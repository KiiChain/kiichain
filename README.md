# Kiichain

![Banner!](assets/kii.png)

# Documentation

Our documentation can be found at:

- [Kiichain docs](https://docs.kiiglobal.io)

# Testnet Genesis

**How to validate on the Kiichain Testnet**

_This is the Kiichain kiichain Testnet_

> Genesis [Published](https://github.com/KiiChain/testnets/blob/main/testnet_oro/genesis.json)

## Hardware Requirements

**Minimum:**

- 8 GB RAM
- 1 TB NVMe SSD
- 4 Cores (modern CPUs)

## Operating System

> Linux (x86_64) or Linux (amd64) Recommended Arch Linux

**Dependencies**

> Prerequisite: go1.23+ required.

- Arch Linux: `pacman -S go`
- Ubuntu: `sudo snap install go --classic`

> Prerequisite: git.

- Arch Linux: `pacman -S git`
- Ubuntu: `sudo apt-get install git`

> Optional requirement: GNU make.

- Arch Linux: `pacman -S make`
- Ubuntu: `sudo apt-get install make`

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
[issues]: https://github.com/kiichain/kiichain/issues
[license]: ./LICENSE

---

## 🧹 Pre-commit Hook Support

This repository supports [pre-commit](https://pre-commit.com/) hooks for automatic linting before each commit. It helps enforce code quality and consistency across the codebase.

### 🔧 Setup Instructions

To enable the hooks on your local environment:

1. Install `pre-commit` (requires Python):

    ```bash
    pip install pre-commit
    ```

2. Install the hooks defined in the config:

    ```bash
    pre-commit install
    ```

3. (Optional) Run all hooks on all files:

    ```bash
    pre-commit run --all-files
    ```


### ✅ Enabled Hooks

- golangci-lint: Linter for Go code  
- markdownlint: Linter for Markdown files  
- yamllint: Linter for YAML files  

🗂️ Configuration file: `.pre-commit-config.yaml` (located at the project root).

