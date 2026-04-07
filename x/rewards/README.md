# Rewards

The rewards module adds the distribution of rewards from a specific community pool.
Anyone can fund the pool but to change or initiate a reward distribution, a proposal
needs to be passed.

## Flow:
1. Fund community pool with reward
2. Create and pass a proposal to create a release schedule
3. At the end of every block, a linear % of the reward will be forward to distribution
4. When the end time of the release is reached, all rewards will have been given away and it will go inactive

## Internal state:
To properly release on time, calculate rewards and keep track, we have a RewardReleaser with the following internal information:

```go
type RewardReleaser struct {
    // Total amount to be rewarded
    TotalAmount types.Coin `protobuf:"bytes,1,opt,name=total_amount,json=totalAmount,proto3" json:"total_amount" yaml:"total_amount"`
    // Amount released
    ReleasedAmount types.Coin `protobuf:"bytes,2,opt,name=released_amount,json=releasedAmount,proto3" json:"released_amount" yaml:"released_amount"`
    // Timestamp of end of release
    EndTime time.Time `protobuf:"bytes,3,opt,name=end_time,json=endTime,proto3,stdtime" json:"end_time" yaml:"end_time"`
    // Last height released
    LastReleaseTime time.Time `protobuf:"bytes,5,opt,name=last_release_time,json=lastReleaseTime,proto3,stdtime" json:"last_release_time" yaml:"last_release_time"`
    // If reward pool is active
    Active bool `protobuf:"varint,6,opt,name=active,proto3" json:"active,omitempty" yaml:"active"`
}
```

At the end of each block, if the releaser is active:
- It will calculate the amt to be distributed, linearly across time based on the last release and the current block time.
- If the amt to be distributed is zero, it goes inactive
- It sends the amt from the pool to the fee collector
- It increases the released amt, the last release time and the community pool with the changes.

## Edge cases and decisions

A few decisions were taken when building the module, which must be taken into account when creating governance proposals and interacting with the module:
- **Overriding schedules**: A new schedule will override the previous one and will not distribute any remaining funds from the previous schedule. This can be used to update amounts or end times, but if a schedule is changed midway it may result in leftover funds in the pool.
- **Changing denoms**: The token denom being distributed can be changed, but it was not meant to support multiple distributions at the same time. It was merely set as a future possibility for other tokens to be rewarded instead of the native token. If the denom needs to be changed, the correct approach is to wait for the current schedule to end so no tokens are left over, then change the base denom and create a new schedule. Otherwise, some tokens may be left over in the pool.
  - If a denom is changed, the `QueryParamsRequest` will return this new denom, even if there is still a current schedule on the old denom.
- **Leftover tokens**: If tokens are left over by any schedule due to overrides, they can be distributed in the future with a new schedule, but not recovered. With this in mind, the recommended approach when changing a schedule is to first halt it, then include its leftover funds in the next schedule, assuming it's on the same token.

## Messages

### FundPool

Sends funds to the community pool, to be used in a future release.

```go
message MsgFundPool {
  string sender = 1 [ (gogoproto.moretags) = "yaml:\"sender\"" ];
  cosmos.base.v1beta1.Coin amount = 2 [
    (gogoproto.moretags) = "yaml:\"amount\"",
    (amino.encoding) = "legacy_coin"
  ];
}
```

**State Modifications:**

- The community pool funds will increase, as well as the module's balance

### ChangeSchedule
Set the schedule to match what is sent. Only the governor can utilize this call, others need to pass a proposal.

```go
message MsgChangeSchedule {
  // authority is the address of the governance account.
  string authority = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];


  // New information for the schedule
  ReleaseSchedule schedule = 2 [ (gogoproto.nullable) = false ];
}

message ReleaseSchedule {
  // Total amount to be rewarded
  cosmos.base.v1beta1.Coin total_amount = 1 [
    (gogoproto.moretags) = "yaml:\"total_amount\""
  ];
  // Amount released
  cosmos.base.v1beta1.Coin released_amount = 2 [
    (gogoproto.moretags) = "yaml:\"released_amount\""
  ];
  // Timestamp of end of release
  google.protobuf.Timestamp end_time = 3 [
    (gogoproto.stdtime) = true,
    (gogoproto.moretags) = "yaml:\"end_time\""
  ];
  // Last height released
  google.protobuf.Timestamp last_release_time = 5 [
    (gogoproto.stdtime) = true,
    (gogoproto.moretags) = "yaml:\"last_release_time\""
  ];
  // If reward pool is active
  bool active = 6 [
    (gogoproto.moretags) = "yaml:\"active\""
  ];
}
```

**State Modifications:**

- Safety check the following
  - Denom of the amt must be the one being used
  - End time must be in the future
  - Funds must be available in the pool
- Changes the reward release schedule to match what is sent

### Update Params

Changes module params. Only the governor can utilize this call, others need to pass a proposal.

```go
message MsgUpdateParams {
  // authority is the address of the governance account.
  string authority = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];

  // params defines the x/rewards parameters to update.
  //
  // NOTE: All parameters must be supplied.
  Params params = 2 [ (gogoproto.nullable) = false ];
}
message Params {
  // Minimal deposit
  string governance_min_deposit = 1;

  // Denom used
  string token_denom = 2;
}
```

**State Modifications:**
- Changes the token_denom and min deposit

## Other important flows
The releaser has a few edge cases that happen when it is initializing or going inactive:

### First iteration:
- There is no previous release time, so we cannot calculate a reward distribution
- Instead of calculating the reward, we just set the last release as the block time
- Next iteration will cover it well

### Last iteration:
- As the first release is delayed, so will be the last one
- Once the EndTime is passed, all the remaining reward will be distributed
- The releaser will just go inactive a block after, when there is no amt to distribute
