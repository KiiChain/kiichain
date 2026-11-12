# Rewards

The rewards module pays staking rewards from a prefunded pool. Each block it
moves tokens from its module account into the `fee_collector`, and
`x/distribution` then allocates them to validators and delegators in the same
way as transaction fees.

The per-block amount comes from an inflation curve driven by the bonded ratio,
the same model as cosmos-sdk `x/mint`. **No tokens are minted.** Every token
paid out must first be sent to the module account, and emissions stop when that
account is empty.

Emissions are off by default (`supply_base = 0`). Governance turns them on.

## Flow

1. Fund the pool: bank-send tokens to the rewards module account (anyone can do this)
2. Pass a governance proposal with `MsgUpdateParams` that sets `supply_base > 0`
3. Each begin-block, release `inflation(bondedRatio) × supply_base / blocks_per_year`
   into `fee_collector`
4. Emissions continue until the module account is empty or governance sets
   `supply_base` back to `0`

The rewards BeginBlocker runs **before** `x/distribution`, so a block's release
is allocated in that same block. Allocation only accrues rewards in
`x/distribution`; it does not change account balances. Validators and delegators
receive the tokens when they withdraw their rewards (for example with
`kiichaind tx distribution withdraw-rewards`).

### Module account

| Format | Address |
|---|---|
| Bech32 | `kii1245yut9zht8q4hz39sd0lzqtzkuw5us5rermc4` |
| Hex | `0x55684e2ca2bace0adc512c1aff880b15b8ea7214` |

The address is derived from the module name (`rewards`), so it is the same on
every KiiChain network. It is removed from the bank blocked-address list so it
can receive plain bank sends.

## Emission formula

```text
inflation = clamp((1 - bondedRatio/goalBonded) × inflationRateChange × bondedRatio,
                  inflationMin, inflationMax)
amount    = truncate(inflation × supplyBase / blocksPerYear)
pay       = min(amount, moduleBalance)
```

| Symbol | Source | Meaning |
|---|---|---|
| `bondedRatio` | `x/staking` `BondedRatio` | Bonded tokens ÷ total staking-token supply |
| `goalBonded` | param `goal_bonded` | Target bonded ratio |
| `inflationRateChange` | param `inflation_rate_change` | Steepness of the curve |
| `inflationMin` / `inflationMax` | params | Floor and ceiling on the annual emission rate |
| `supplyBase` | param `supply_base` | Notional base used to size annual provisions |
| `blocksPerYear` | param `blocks_per_year` | Expected blocks per year |
| `moduleBalance` | `x/bank` | Rewards module account balance in `token_denom` |

### Curve shape

The unclamped curve is a parabola in `bondedRatio`:

- It is `0` at `bondedRatio = 0` and at `bondedRatio = goalBonded`.
- It peaks at `bondedRatio = goalBonded / 2` with value
  `inflationRateChange × goalBonded / 4`.
- Above `goalBonded` it goes negative, so the clamp returns `inflationMin`.

With the defaults (`goal_bonded = 0.67`, `inflation_rate_change = 0.13`) the
peak is `0.13 × 0.67 / 4 = 2.1775%`. The default `inflation_max` of `20%` does
not bind unless `inflation_rate_change` or `goal_bonded` is raised.

The rate grows as the bonded ratio rises toward `goalBonded / 2`, then falls to
the floor as it approaches `goalBonded`. This differs from `x/mint`, which
raises inflation when the bonded ratio is low.

### `supply_base` is not the total supply

`annual provision = inflation × supply_base`. `supply_base` is a governance
knob that sets the scale of emissions. It is **not** read from the chain's
total supply. It is an integer in base units of `token_denom` (`akii`, 18
decimals).

### Per-block, not wall-clock

Like `x/mint`, each block releases a fixed share of the annual provision
(`annual / blocks_per_year`). It does not accrue by elapsed time. If blocks run
slower than `blocks_per_year` assumes, less is released per calendar year. Keep
`blocks_per_year` in line with the real block time (default `15778800`, which
assumes 2s blocks: `60 × 60 × 8766 / 2`).

### Worked example

Defaults, with `supply_base = 1e27` akii (1,000,000,000 KII):

| Bonded ratio | Inflation | Annual provision | Per block |
|---|---|---|---|
| 0.335 (peak) | 2.1775% | 21,775,000 KII | ~1.380 KII |
| 0.50 | ~1.6493% | ~16,492,537 KII | ~1.045 KII |
| ≥ 0.67 | 0% (floor) | 0 | 0 |

## Begin block

Each block, in order:

1. If `supply_base` is `0`, do nothing.
2. Read the module account balance in `token_denom`. If it is `0`, do nothing.
3. Read the bonded ratio from `x/staking`. On error, log it and skip the block.
4. Compute `amount` and `inflation`. If `amount` is `0`, do nothing.
5. Cap `amount` at the module balance. A short balance pays what is there.
6. Send `amount` from the rewards module account to `fee_collector`. On error,
   log it and skip the block. The chain does not halt.
7. Add `amount` to `RewardPool.total_released`. If the stored denom differs from
   `token_denom` (for example after a governance denom change), the counter
   resets to `amount`.
8. Emit `reward_distributed` and record telemetry.

## State

```go
type RewardPool struct {
    TotalReleased sdk.Coin // cumulative amount sent to fee_collector
}
```

`RewardPool` is an observability counter. Spendable funds are the module account
balance in `x/bank`. The module stores no balance of its own.

### Params

```go
type Params struct {
    TokenDenom          string         // default "akii"
    GoalBonded          math.LegacyDec // default 0.67
    InflationMin        math.LegacyDec // default 0
    InflationMax        math.LegacyDec // default 0.20
    SupplyBase          math.Int       // default 0 (emissions off)
    InflationRateChange math.LegacyDec // default 0.13
    BlocksPerYear       uint64         // default 15778800
}
```

Validation:

| Param | Rule |
|---|---|
| `token_denom` | Valid SDK denom |
| `goal_bonded` | `0 < x ≤ 1` |
| `inflation_min` | `≥ 0` |
| `inflation_max` | `≥ inflation_min` |
| `supply_base` | `≥ 0` |
| `inflation_rate_change` | `0 < x ≤ 1` |
| `blocks_per_year` | `> 0` |

### Genesis

```json
{
  "params": {
    "token_denom": "akii",
    "goal_bonded": "0.670000000000000000",
    "inflation_min": "0.000000000000000000",
    "inflation_max": "0.200000000000000000",
    "supply_base": "0",
    "inflation_rate_change": "0.130000000000000000",
    "blocks_per_year": "15778800"
  },
  "reward_pool": { ... }
}
```

The default `reward_pool` is empty. Genesis validation checks the params and
requires a non-zero `total_released` to use `token_denom`.

## Messages

### MsgUpdateParams

Governance-only (`authority` must be the `x/gov` module account). Replaces all
module params, so every field must be set.

```protobuf
message MsgUpdateParams {
  string authority = 1;
  Params params = 2;
}
```

- `supply_base > 0` turns emissions on.
- `supply_base = 0` turns them off. Funds stay in the module account.

Emits `update_params`.

### Funding

There is no fund message. Send tokens to the module account with a normal bank
send, in the current `token_denom`:

```bash
# Read the current emission denom
kiichaind q rewards params

kiichaind tx bank send <from> kii1245yut9zht8q4hz39sd0lzqtzkuw5us5rermc4 <amount><token_denom>
```

Only the module account balance in `token_denom` is released. Tokens sent in any
other denom stay in the module account and are never emitted. If governance
changes `token_denom`, balances in the old denom stop funding emissions.

## Enabling emissions

Proposal file (`proposal.json`):

```json
{
  "messages": [
    {
      "@type": "/kiichain.rewards.v1.MsgUpdateParams",
      "authority": "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
      "params": {
        "token_denom": "akii",
        "goal_bonded": "0.67",
        "inflation_min": "0",
        "inflation_max": "0.20",
        "supply_base": "1000000000000000000000000000",
        "inflation_rate_change": "0.13",
        "blocks_per_year": "15778800"
      }
    }
  ],
  "metadata": "",
  "deposit": "10000000akii",
  "title": "Enable utility rewards emissions",
  "summary": "Set supply_base to start x/rewards emissions"
}
```

```bash
kiichaind tx gov submit-proposal proposal.json --from <key>
```

`MsgUpdateParams` replaces **all** params. The values above are the defaults;
before submitting, query the live params with `kiichaind q rewards params` and
copy them into the proposal, changing only the fields you mean to change.
Otherwise any param that differs from the example is overwritten.

To turn emissions off, query the live params, copy them into the proposal
unchanged, and set only `"supply_base": "0"`.

## Queries

| CLI | REST | Returns |
|---|---|---|
| `kiichaind q rewards params` | `GET /kiichain/rewards/v1/params` | Module params |
| `kiichaind q rewards reward-pool` | `GET /kiichain/rewards/v1/reward-pool` | `reward_pool.total_released` and `balance` (module account balance in `token_denom`) |

## Events

| Type | Attributes | When |
|---|---|---|
| `update_params` | — | `MsgUpdateParams` succeeds |
| `reward_distributed` | `amount`, `total_released`, `inflation_rate`, `bonded_ratio` | A release is sent to `fee_collector` |

Telemetry gauges: `rewards_reward_released` (amount this block) and
`rewards_module_balance` (balance after the release).

## Legacy `v1beta1` API

The live API is `kiichain.rewards.v1`. The mainnet `kiichain.rewards.v1beta1`
types stay registered so existing transactions and proposals still decode.

| v1beta1 | Behavior |
|---|---|
| `MsgUpdateParams` | Updates only `token_denom`; other params are unchanged |
| `MsgFundPool` | Returns an error. Use a bank send. |
| `MsgChangeSchedule` | Returns an error. Schedules were removed. |
| `Query/Params` | Returns `token_denom` only |
| `Query/RewardPool` | Reports the module account balance as `community_pool` |
| `Query/ReleaseSchedule` | Returns an error |

## Migration (consensus version 1 → 2)

Consensus version 2 replaces the timed linear `ReleaseSchedule` with the
inflation curve. The store migration:

1. Deletes the `ReleaseSchedule` state.
2. Fills the new params with defaults and keeps the existing `token_denom`.
   `supply_base` stays `0`, so emissions stay off after the upgrade until
   governance enables them.
3. Initializes `RewardPool.total_released` if unset.

Removed in v2: `ReleaseSchedule`, `MsgChangeSchedule`, `MsgFundPool`, the
release-schedule query and CLI, and the end-date and total-budget schedule
semantics.
