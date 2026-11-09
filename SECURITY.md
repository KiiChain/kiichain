# Security Policy

## Introduction

Security is very important for Kii Global and its community. This document outlines security procedures and general policies for the `kiichain` project.

### Guidelines for Responsible Vulnerability Testing and Reporting

1. **Refrain from testing vulnerabilities on our publicly accessible environments**, including but not limited to:

- Kii mainnet
- Kii frontend
- Kii public testnets
- Kii testnet frontend

2. **Avoid reporting security vulnerabilities through public channels, including GitHub issues**

To privately report a security vulnerability, please choose one of the following options:

### 1. Email

Send your detailed vulnerability report to `devs@kiiglobal.io`.

### 2. GitHub Private Vulnerability Reporting

Utilize [GitHub's Private Vulnerability Reporting](https://github.com/kiichain/kiichain/security/advisories/new) for confidential disclosure.

## Submit Vulnerability Report

When reporting a vulnerability through either method, please include the following details to aid in our assessment:

- Type of vulnerability
- Description of the vulnerability
- Steps to reproduce the issue
- Impact of the issue
- Explanation on how an attacker could exploit it

## Vulnerability Disclosure Process

1. **Initial Report**: Submit the vulnerability via one of the above channels.
2. **Confirmation**: We will confirm receipt of your report within 48 hours.
3. **Assessment**: Our security team will evaluate the vulnerability and inform you of its severity and the estimated time frame for resolution.
4. **Resolution**: Once fixed, you will be contacted to verify the solution.
5. **Public Disclosure**: Details of the vulnerability may be publicly disclosed after ensuring it poses no further risk.

During the vulnerability disclosure process, we ask security researchers to keep vulnerabilities and communications around vulnerability submissions private and confidential until a patch is developed. Should a security issue require a network upgrade, additional time may be needed to raise a governance proposal and complete the upgrade.

During this time:

- Avoid exploiting any vulnerabilities you discover.
- Demonstrate good faith by not disrupting or degrading kii's services.

## Feature request

For a feature request, e.g. module inclusion, please make a GitHub issue. Clearly state your use case and what value it will bring to other users or developers on Kii.

## Severity Characterization

| Severity     | Description                                                                                                                                                                                                                                                                         |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **CRITICAL** | Direct theft of funds; permanent freezing of funds (irrecoverable protocol-level lock); total network shutdown, meaning the network can’t confirm transactions and recovery requires a hard fork; consensus manipulation or fork creation, including chain splits or invalid blocks |
| **HIGH**     | Application-level DoS that crashes any node under specific payloads; Temporary freezing of network transactions; Unintended temporary freezing of funds                                                                                                                             |
| **MEDIUM**   | Application-level DoS affecting a subset of nodes; unintended protocol behavior caused by protocol quirks or edge-case consensus rules; minor block reorg exploits affecting specific dApps                                                                                         |
| **LOW**      | Transaction fee modification or miscalculation without economic impact; P2P gossip layer issues such as propagation delays or non-malicious message flooding                                                                                                                        |

Examples of severity characterization:

- **CRITICAL**:
    - Panic in BeginBlocker/EndBlocker/FinalizeBlock that any user can trigger, halting the chain until a coordinated upgrade.
    - Unauthorized mint/burn through x/bank or a custom module.
    - Ante handler bypass that skips signature checks.
    - IBC light client or packet forgery.
    - Precompile or EVM-state desync that drains funds.
- **HIGH**:
    - A crafted tx or query that crashes nodes.
    - Liveness stalls that recover without a hard fork.
    - Stuck unbonding, or IBC escrow locked until governance acts.
- **MEDIUM**:
    - Gas metering gaps, such as unmetered loops in a module or iterators over unbounded stores, that let an attacker slow down some nodes.
    - EVM/precompile edge cases that diverge from Ethereum semantics.
    - Proposer-influenced BlockTime abuse.
- **LOW**:
    - Fee or gas estimation inaccuracies.
    - CometBFT peer-handling quirks.
    - Crashes limited to RPC or sentry nodes.

## Feedback on this Policy

For recommendations on how to improve this policy, either submit a pull request or email `devs@kiiglobal.io`.
