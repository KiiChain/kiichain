# Rewards

The rewards module distributes prefunded tokens from a community pool into the
fee collector each block, using an inflation curve driven by the bonded ratio.
Anyone can fund the pool; governance enables emissions by setting `supply_base`.

## Flow

1. Fund the community pool (`MsgFundPool`)
2. Pass a gov proposal to set params, including a non-zero `supply_base` (`MsgUpdateParams`)
3. Each begin-block, release `inflation(bondedRatio) × supply_base × Δt / year` into `fee_collector`
4. Emissions continue until the pool is empty or governance sets `supply_base` back to `0`

## Emission formula

```
inflation = clamp((1 - bondedRatio/goalBonded) × 0.13 × bondedRatio, inflationMin, inflationMax)
amount    = inflation × supplyBase × elapsedNs / nsPerYear
pay       = min(amount, poolBalance)
```

`inflation_rate_change` is a fixed constant (`0.13`). `supply_base` defaults to `0` (emissions off).

## Internal state

```go
type RewardPool struct {
    CommunityPool   sdk.DecCoins
    LastReleaseTime time.Time
    TotalReleased   sdk.Coin // cumulative observability counter
}
```

Params:

```go
type Params struct {
    TokenDenom   string
    GoalBonded   math.LegacyDec // default 0.67
    InflationMin math.LegacyDec // default 0
    InflationMax math.LegacyDec // default 0.20
    SupplyBase   math.Int       // default 0
}
```

## Messages

### FundPool

Sends funds to the community pool.

### UpdateParams

Governance-only. Sets all module params. Setting `supply_base > 0` enables emissions;
setting it to `0` disables them.

## First iteration

When `last_release_time` is zero, BeginBlocker stamps the current block time and
skips release. The next block starts accruing.
