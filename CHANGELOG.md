# CHANGELOG

## Unreleased

### Fixed

- Reject `MsgCreateVestingAccount`, `MsgCreatePeriodicVestingAccount`, and `MsgCreatePermanentLockedAccount` in the Cosmos ante (top-level and nested in `authz.MsgExec`) so new vesting / locked accounts cannot be opened after the v7.3.2 upgrade
- Enable a bank `SendRestriction` for the 22 Aug 2026 incident addresses in the `v7.4.0` upgrade (after fund recovery) so Cosmos, precompile, and EVM native transfers cannot send from or to them after the upgrade height



### Dependencies

- Bump EVM fork to `v0.6.2-fork.1`, applied via the coordinated `v7.3.2` upgrade (private Cosmos EVM security hotfix; cannot build from public source until 2026-08-28)
- Bump EVM fork to `v0.6.2-fork.2`

## v7.3.1 - 2026-08-06

### Fixed

- Keep the gov module account on the bank blocked list so EVM BalanceHandler does not mirror gov deposits into StateDB (fixes Safe → gov precompile `deposit` failing on commit with `unauthorized`) ([#368](https://github.com/KiiChain/kiichain/pull/368))

## v7.3.0 - 2026-07-28

### Dependencies

- [EVM](https://github.com/KiiChain/evm) fork from v0.6.0-fork.1 to [v0.6.0-fork.2](https://github.com/KiiChain/evm/releases/tag/v0.6.0-fork.2): bounded internal EVM call gas limit, EVM fee refunds, distribution precompile 32-byte withdraw fix, and CosmWasm EVM query undercharge fix
- [EVM](https://github.com/KiiChain/evm) fork bump from `v0.6.0-fork.2` to [v0.6.1-fork.1](https://github.com/KiiChain/evm/releases/tag/v0.6.1-fork.1), applied via the coordinated `v7.3.0` upgrade: the July 2026 Cosmos EVM hotfix (precompile gas accounting alignment and StateDB locked-balance snapshotting), published upstream as [cosmos/evm v0.6.1](https://github.com/cosmos/evm/releases/tag/v0.6.1)

### Fixed

- Close an expedited-governance whitelist bypass in `GovExpeditedProposalsDecorator` where the check only inspected top-level messages: a non-whitelisted `MsgSubmitProposal` wrapped in `authz.MsgExec` could enter the expedited voting path. The decorator now recurses into `authz.MsgExec` (including nested execs) and applies the expedited whitelist validation to wrapped proposals ([#353](https://github.com/KiiChain/kiichain/pull/353))
- Compute the oracle ballot `StandardDeviation` as a stake-weighted variance (weight each squared deviation by the vote's power and divide by total voting power) instead of an unweighted average divided by the vote count, aligning the reward-band width with the stake-weighted median and preventing a group of low-stake validators from inflating the deviation to widen the accepted vote window ([#354](https://github.com/KiiChain/kiichain/pull/354))
- Close an oracle slashing bypass in the `EndBlocker` where validators were scored against the post-filtered `voteTargets` map: a denom that received votes but was pushed below the vote threshold (e.g. by a coordinated group abstaining) was dropped from the scoring denominator, letting the abstainers avoid miss penalties. Participation is now scored against the configured targets that received votes (passing targets plus below-threshold targets), crediting validators that voted on a below-threshold target while counting abstention on it as a miss; targets that received no votes at all are still excluded so a legitimately unpriceable denom cannot mass-slash the validator set ([#352](https://github.com/KiiChain/kiichain/pull/352))
- Allow EIP-7702 delegated EOAs to send direct EVM transactions by exempting delegation-designator code from the externally-owned-account-only check in `VerifyIfAccountExists`, so accounts that delegate via `SetCodeTx` can still manage (and revoke) their own delegation without a sponsored transaction ([#350](https://github.com/KiiChain/kiichain/pull/350))
- Remove the forced minimum 1-unit-per-block reward release in `CalculateReward` and skip (instead of deactivating) sub-unit blocks in the rewards `BeginBlocker`, so the proportional share accumulates and the pool follows the configured schedule independent of block time ([#347](https://github.com/KiiChain/kiichain/pull/347))
- Reject `MsgEthereumTx` from being dispatched through the authz keeper (including when nested inside `authz.MsgExec`), closing an EVM ante bypass on message-router execution paths that skip the ante handler ([#342](https://github.com/KiiChain/kiichain/pull/342))
- Fix feegrant denomination bypass in the cosmos fee ante handler by converting the fee before consuming the grant, so `UseGrantedFees` is checked against the same coins later deducted ([#343](https://github.com/KiiChain/kiichain/pull/343))
- Bound tokenfactory denom metadata size (`MaxDenomMetadataSize`) in `MsgSetDenomMetadata.ValidateBasic` and `msgServer.SetDenomMetadata` to prevent oversized metadata rewrites from forcing unbounded native store writes ([#341](https://github.com/KiiChain/kiichain/pull/341))
- Fix native token supply inflation from the stateful precompiles by wrapping the account address codec (`evmAddressCodec`) to reject non-20-byte accounts at decode time ([#340](https://github.com/KiiChain/kiichain/pull/340))
- Close governance vote minimum-stake bypass in `GovVoteDecorator` by enforcing the stake check on `MsgVoteWeighted` and recursing into nested `authz.MsgExec` messages ([#344](https://github.com/KiiChain/kiichain/pull/344))
- Prevent a chain halt in the rewards `BeginBlocker` by routing `SendCoinsFromModuleToModule` failures through `haltSchedule` instead of returning a fatal error ([#346](https://github.com/KiiChain/kiichain/pull/346))
- Add a `ValidateModuleAccounting` check (rewards module bank balance must cover the `CommunityPool`) and run it at genesis
- Fix CosmWasm EVM query path repeatable undercharged EVM execution ([#345](https://github.com/KiiChain/kiichain/pull/345))

## v7.2.0 - 2026-04-16

### Added

- Emit `update_params`, `fund_pool`, `change_schedule`, and `reward_distributed` events from x/rewards
- Emit `update_params` and `set_denom_metadata` events from tokenfactory
- Emit `update_params`, `update_fee_tokens`, `module_disabled` and `token_disabled` events on fee abstraction
- Emit `update_params` event on oracle module

### Fixed

- Refactor `PerformSetMetadata` in wasmbinding to delegate to `msgServer.SetDenomMetadata`, ensuring the `EnableSetMetadata` capability check is enforced ([#329](https://github.com/KiiChain/kiichain/pull/329))
- Ensure that `UpdateTokenMetadata.Decimals` matches the ERC20 or bank records ([#325](https://github.com/KiiChain/kiichain/pull/325))
- Fixed odd validation on tokenfactory change admin that blocked removing admin from the token ([#324](https://github.com/KiiChain/kiichain/pull/324))
- Fix division-by-zero chain halt in `CalculateReward` caused by sub-second schedule durations; replace `Seconds()` truncation with `Nanoseconds()` precision and release full remaining reward when `EndTime <= LastReleaseTime` ([#267](https://github.com/KiiChain/kiichain/issues/267))
- Add denom string length validation (max 128 bytes) to oracle precompile and query server to prevent memory exhaustion via oversized inputs
- Add result limits to oracle list queries (ExchangeRates, Actives, VoteTargets capped at 1000; PriceSnapshotHistory capped at 500) to prevent unbounded iteration
- Fix NewClaim constructor assigning power to Weight field instead of the weight parameter (x/oracle/types/ballot.go)
- Enforce community pool has sufficient balance before distributing rewards, returning an error instead of panicking on accounting divergence
- Apply whitelist bank denoms on oracle genesis so validators can submit votes from block 0
- Disable fee abstraction when base token price is 0 to prevent incorrect fee conversions
- Ensure native oracle denoms are always on whitelist and registered as vote targets when updating fee abstraction params
- Validate rewards baseDenom using sdk.ValidateDenom to enforce proper denom format (min 3 chars, valid characters, no leading digits)
- Ensure feeTokens is not nil at genesis
- Ensure feeTokenMetadata initial prices after updateFeeTokenMetadata is picked up from oracle
- Use `DecCoins.Validate()` on `RewardPool.ValidateGenesis` to catch malformed denom formats, duplicate denoms, bad ordering ([#323](https://github.com/KiiChain/kiichain/pull/323))
- Enforce denom consistency in `GenesisState.Validate` with `Params.TokenDenom`
- Limited tokenfactory queries, removing denial of service possibility ([#328](https://github.com/KiiChain/kiichain/pull/328))
- Indexed admins to reduce query space on tokenfactory denom queries ([#328](https://github.com/KiiChain/kiichain/pull/328))

### Removed

- Removed price field input in updateTokenMetadata request

## v7.1.0-mainnet - 2026-03-13

### Fixed
- Merged v7.0.0 and v7.1.0 in one upgrade for mainnet

## v7.1.0 - 2026-03-10

### Dependencies
- [EVM](https://github.com/cosmos/evm) from v0.5.1 to [v0.6.0](https://github.com/cosmos/evm/releases/tag/v0.6.0)

### Added

- Re enabled jailing on oracle module
- Re enabled ICS precompile

### Fixed

- Fixed misbehaviors on oracle module slashing

## v7.0.1 - 2026-02-11

### Fixed

- Pick evm-chain-id from genesis instead of config / flag

## v7.0.0 - 2026-02-03

## Removed

- Stripped out wasmd precompile

## v7.0.0 - 2026-02-03

## Dependencies
- Bump Go version from 1.23 to 1.24
- Bump [EVM](https://github.com/cosmos/evm) from v0.4.2 to [v0.5.1](https://github.com/cosmos/evm/releases/tag/v0.5.1)
- Bump [cosmos go-ethereum](github.com/cosmos/go-ethereum) to [v1.16.2](https://github.com/cosmos/go-ethereum/releases/tag/v1.16.2-cosmos-1)
- Bump golang crypto from 0.41.0 to 0.45.0
- Bump ulikunitz/xz from 0.5.11 to 0.5.14
- Bump opencontainers/runc from 1.1.14 to 1.2.8
- Bump Hashicorp/go-better to from v1.7.8 to v1.7.9
- Bump docker from 27.1.1 to 28.0.0
- Bump tendermint from v0.38.19 to v0.38.21

## Added
- Added telemetry for reward distribution

## Removed

- Removed dead code related to no gas consumption
- Removed ICS precompile

## Fixed
- Fix tokenfactory mintTo to check blocked address before minting ([#258](https://github.com/KiiChain/kiichain/issues/258))
- Return error instead of nil in RemoveExcessFeeds to properly propagate storage errors
- Fixed oracle module ConsensusVersion constant not being used ([#256](https://github.com/KiiChain/kiichain/issues/256))
- Fix oracle weighted median threshold to require >50% instead of >=50% for majority
- Make ERC20 fee conversion rounding explicit by always rounding up to avoid underpayment
- Handle ValAddressFromBech32 error in oracle EndBlocker [#234](https://github.com/KiiChain/kiichain/issues/234)
- Fix wrong nil check in ValidateFeeder function ([#250](https://github.com/KiiChain/kiichain/issues/250))
- Handle ignored error from TotalBondedTokens in pickReferenceDenom [#230](https://github.com/KiiChain/kiichain/issues/230)
- Handle error in pickReferenceDenom instead of panic to prevent consensus failure [#203](https://github.com/KiiChain/kiichain/issues/203)

- Fixed EVM mempool public nodes broadcast bug
- Fixed Ledger failure on linux nano
- Fixed ledger usage of default chain ID
- Remove references to time.Now() on release schedule validation
- Fix Oracle precompile ParseGetTwapsArgs missing validation for lookbackPeriod [#204](https://github.com/KiiChain/kiichain/issues/204)
- Remove unsafe math.Sqrt usage in oracle ballot standard deviation calculation
- Remove panic on oracle slash logic when validator is not found
- Updated swagger files to match current project

## v6.0.0 - 2025-11-25

## Added

- Cosmos EVM integrations tests added to the repo
- Added logs and telemetry for reentrance detection in wasmd precompile
- Add further validations to Wasm Oracle query bindings
- Moves Wasmd reentrance lock to the core of the wasmd contract to avoid reentrance attacks on queries and instantiations
- Added EVM mempool

## Removed

- Removed IBC precompile since ICS20 precompile also handles IBC transfers.
- Removed fallback native price param on the fee abstraction module. Instead of using hardcoded price, the module disables itself when price is lacking.
- Removed previous upgrades from main branch.

## DEPENDENCIES

-   Bump [EVM](github.com/cosmos/evm) to [v0.4.2](https://github.com/cosmos/evm/releases/tag/v0.4.2)

## Fixed

- Fix IBC unsafe log as reported on issue [#143](https://github.com/KiiChain/kiichain/issues/143)
- Fix wasmd precompile bad input handling for coins [#147](https://github.com/KiiChain/kiichain/issues/147)
- Fix Oracle module ante decorator to allow first vote and install oracle ante handlers
- Fix IBC validation of negative numbers happens a bit early [#144](https://github.com/KiiChain/kiichain/issues/144)
- Fix incorrect error passing on tokenfactory wasmbinding
- Fix blocked address checked after minting in wasmbinding tokenfactory 
- Fix account balance information being stale on fee abstraction's cosmos' ante

### Documentation

- Add further docs related to the reentrance key creation

## v5.1.0 — 2025-10-15

### Fixes

- Fix wasmd precompile against reentrance attacks

## v5.0.0 — 2025-09-26
- Swapped default micro oracle coins to their normal denom
- Correct some doc strings mentioning mint module

## DEPENDENCIES

-   Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.53.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.53.4)
-   Bump [Wasmd](https://github.com/CosmWasm/wasmd) to [v0.61.2](https://github.com/CosmWasm/wasmd/releases/tag/v0.61.2)
-   Bump [EVM](github.com/cosmos/evm) to [v0.4.1](https://github.com/cosmos/evm/releases/tag/v0.4.1)
-   Bump [IBC-go](https://github.com/cosmos/ibc-go/) to [v10.3.0](https://github.com/cosmos/ibc-go/releases/tag/v10.3.0)
-   Removed crisis module

### Fixes

-   Add missing address validation in `GetTokenfactoryDenomsByCreator` query to prevent potential crashes with malformed addresses
-   Override EVM chain ID if default
-   Change evm chain coin info mapping to always default

## v4.0.0 — 2025-08-06

### Added

-   Add the fee abstraction module to the chain

## v3.0.0 — 2025-07-01

No changes were made since the release candidate.

## v3.0.0-rc1 -- 2025-06-25

### Added

-   Add the oracle module to the chain
-   Add the oracle wasmbinding
-   Add the oracle EMV precompile
-   Add E2E tests to IBC precompile
-   Add E2E tests to wasmd precompile

## v2.0.0 -- 2025-06-18

### Added

-   Initial chain creation
-   Add EVM wasmbinding queries
-   Add bech32 wasmbinding queries
-   Add IBC precompile to transfer via EVM
-   Add correct ibc keepers to ibc precompiles
-   Add Rewards module

### Changed

-   Update pipelines by adding codeql, codecov and changelog diff checker
-   Refactor the tokenfactory wasmbinding into its own path
-   Refactor the wasmbinding implementation to allow multiple msg and query types
