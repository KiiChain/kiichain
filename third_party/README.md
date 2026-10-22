# third_party

Vendored copies of IBC middleware that have no upstream release for Cosmos SDK v0.54 and ibc-go v11. `go.mod` replaces the upstream module paths with these directories:

```go
github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10 => ./third_party/packet-forward-middleware
github.com/cosmos/ibc-apps/modules/rate-limiting/v10 => ./third_party/rate-limiting
```

The module paths stay on `/v10` so the Go import paths, store keys (`packetfowardmiddleware`, `ratelimit`), and proto type URLs (`/packetforward.v1.*`, `/ratelimit.v1.*`) match the modules already running on chain.

## Sources

| Directory | Upstream module | Tag | Commit |
| --- | --- | --- | --- |
| `packet-forward-middleware` | `github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10` | [`middleware/packet-forward-middleware/v10.6.0`](https://github.com/cosmos/ibc-apps/tree/middleware/packet-forward-middleware/v10.6.0/middleware/packet-forward-middleware) | `ea0c803824041e4a4d55730a48a1f49be2686bbe` |
| `rate-limiting` | `github.com/cosmos/ibc-apps/modules/rate-limiting/v10` | [`modules/rate-limiting/v10.7.0`](https://github.com/cosmos/ibc-apps/tree/modules/rate-limiting/v10.7.0/modules/rate-limiting) | `b524d3eee2990fde1a28e807628d6414c3f7e547` |

## Local patches

Both modules:

- `github.com/cosmos/ibc-go/v10` imports moved to `github.com/cosmos/ibc-go/v11`.
- `cosmossdk.io/log` moved to `cosmossdk.io/log/v2`, and `cosmossdk.io/store/prefix` moved to `github.com/cosmos/cosmos-sdk/store/v2/prefix`.
- `go.mod` / `go.sum` bumped to cosmos-sdk v0.54.3 and ibc-go v11.2.0.
- `IBCMiddleware` gains `SetICS4Wrapper` and `SetUnderlyingApplication`, and `NewIBCMiddleware` returns `*IBCMiddleware`, so the middleware fits the ibc-go v11 stack builder.
- `Keeper` gains `SetICS4Wrapper`.
- Upstream `_test.go` files and the `test/` and `testing/` directories are removed. They depend on the ibc-go v10 testing package.

Generated `*.pb.go` files and `proto/` are unchanged from the tags above.

## Removing this directory

ibc-go v11 ships both middlewares as `github.com/cosmos/ibc-go/v11/modules/apps/packet-forward-middleware` and `.../rate-limiting`. They keep the same store keys and register their protos under `ibc.applications.packet_forward_middleware.v1` and `ibc.applications.rate_limiting.v1`. Moving to them changes the message type URLs, so it needs its own upgrade that accounts for stored governance proposals, authz grants, and circuit breaker entries that reference `/ratelimit.v1.*`.
