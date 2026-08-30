# Rewards

The rewards module distributes prefunded tokens from a community pool into the
fee collector each block, using an inflation curve driven by the bonded ratio.
Anyone can fund the pool; governance enables emissions by setting `supply_base`.

## Flow

1. Fund the community pool (`MsgFundPool`)
2. Pass a gov proposal to set params, including a non-zero `supply_base` (`MsgUpdateParams`)
3. Each begin-block, release `inflation(bondedRatio) × supply_base / blocks_per_year` into `fee_collector`
4. Emissions continue until the pool is empty or governance sets `supply_base` back to `0`

## Emission formula

```text
inflation = clamp((1 - bondedRatio/goalBonded) × inflationRateChange × bondedRatio, inflationMin, inflationMax)
amount    = inflation × supplyBase / blocksPerYear
pay       = min(amount, poolBalance)
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

## Internal state

```go
type RewardPool struct {
    CommunityPool sdk.DecCoins
    TotalReleased sdk.Coin // cumulative observability counter
}
```

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

### FundPool

Sends funds to the community pool.

### UpdateParams

Governance-only. Sets all module params. Setting `supply_base > 0` enables emissions;
setting it to `0` disables them.
