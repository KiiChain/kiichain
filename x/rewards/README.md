# Rewards

The rewards module distributes prefunded tokens into the fee collector each
block, using an inflation curve driven by the bonded ratio. Anyone can fund
the module account with a bank send; governance enables emissions by setting
`supply_base`.

## Flow

1. Bank-send tokens to the rewards module account
2. Pass a gov proposal to set params, including a non-zero `supply_base` (`MsgUpdateParams`)
3. Each begin-block, release `inflation(bondedRatio) × supply_base / blocks_per_year` into `fee_collector`
4. Emissions continue until the module account is empty or governance sets `supply_base` back to `0`

If the calculated release is larger than the module account balance, the block
pays only that balance. A zero balance skips the block. A failed bank send is
logged and the chain continues.

## Emission formula

```text
inflation = clamp((1 - bondedRatio/goalBonded) × inflationRateChange × bondedRatio, inflationMin, inflationMax)
amount    = inflation × supplyBase / blocksPerYear
pay       = min(amount, moduleBalance)
```

This matches cosmos-sdk `x/mint`: a fixed per-block share of the annual provision
(`annual / blocks_per_year`), not a wall-clock Δt accrual.

Where:

- `bondedRatio` — fraction of the token supply currently staked (read from x/staking)
- `goalBonded` — target stake ratio; the curve peaks at `goalBonded/2` and hits zero at `goalBonded`
- `inflationRateChange` — curve steepness (gov param, default `0.13`)
- `inflationMin` / `inflationMax` — floor and ceiling on the emission rate
- `blocksPerYear` — expected blocks in a year used to size the per-block release
  (gov param, default `15778800` for a 2s block time)
- `supplyBase` — notional base that sizes annual provisions (`annual = inflation × supplyBase`).
  It is **not** the chain total supply; it is a governance knob for emission scale.
  Defaults to `0` (emissions off).
- `moduleBalance` — rewards module account balance for `token_denom` in x/bank

## Internal state

```go
type RewardPool struct {
    TotalReleased sdk.Coin // cumulative observability counter
}
```

Spendable funds are the module account balance. `RewardPool` does not track them.

Params:

```go
type Params struct {
    TokenDenom          string
    GoalBonded          math.LegacyDec // default 0.67
    InflationMin        math.LegacyDec // default 0
    InflationMax        math.LegacyDec // default 0.20
    SupplyBase          math.Int       // default 0
    InflationRateChange math.LegacyDec // default 0.13
    BlocksPerYear       uint64         // default 15778800
}
```

## Messages

### UpdateParams

Governance-only. Sets all module params. Setting `supply_base > 0` enables emissions;
setting it to `0` disables them.

The live API is `kiichain.rewards.v1`. Mainnet `kiichain.rewards.v1beta1` messages
stay registered so existing proposals still unpack. Legacy `MsgUpdateParams` only
updates `token_denom`. Legacy `MsgFundPool` and `MsgChangeSchedule` return an error.
