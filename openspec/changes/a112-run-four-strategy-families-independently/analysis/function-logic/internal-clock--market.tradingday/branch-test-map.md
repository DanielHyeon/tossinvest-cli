# Branch Test Map: `Market.TradingDay`

- Source: `internal/clock/market.go`, SHA-256 `85264b6e9b4e6b13ddee690ca62e0d30cbd73efeb6926e626abc3d20c622f7a9`; branch IDs follow `ast.json` (generated 2026-08-17).
- AST counts: branches 1, returns 2, calls 3, defers 0, go statements 0. Source range `158:1`–`164:2`. Signature `(m Market) TradingDay(t time.Time) (string, error)`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1b brief (official raw reader + bar producer). Any later body edit requires a fresh RED/BTM.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 160:2 — an unknown market (`nasdaq`) must return an empty day together with `ErrUnknownMarket`, never a host-local default | `TestUnknownMarketFailsClosed` | n/a (not edited) | existing suite green |

Fall-through at 163 (not a branch): `t.In(loc).Format("2006-01-02")` — `TestTradingDayBoundary` drives 2026-03-30T13:30:00Z and 2026-03-30T20:00:00Z through both markets and asserts KR rolls over (2026-03-30 then 2026-03-31) while US does not (2026-03-30 twice), together with the matching `SameTradingDay` answers. Zone correctness across DST is held by `TestMarketLocations`, `TestUTCOffsetAcrossDST` and `TestRegularSessionDSTTable`.

Verification: `go test ./internal/clock -count=1` green on 2026-08-17 (exit 0). No RED round applies — a112 does not edit this function.
