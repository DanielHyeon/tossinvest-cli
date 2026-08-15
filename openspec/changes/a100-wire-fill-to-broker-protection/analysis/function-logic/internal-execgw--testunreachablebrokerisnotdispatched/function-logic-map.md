# Function Logic Map: `TestUnreachableBrokerIsNotDispatched`

- Source: `internal/execgw/gateway_test.go` (727-757)
- AST evidence: `ast.json` — AST branches 4 (B1-B4).
- Risk scan: `risk-pattern-report.md` — no configured pattern matched.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `preWriteFailureTransport` | Every `RoundTrip` returns an error before connection/write | Test-owned `http.Client.Transport` | Token acquisition fails; order POST is never constructed or sent |
| `transport.calls` | Exactly `1` | Test-owned recorder | Test fails when an unexpected retry or mutation request reaches transport |
| `transport.method` / `transport.path` | `POST` / `/oauth2/token` | First and only test transport request | Test fails if the only request is not token acquisition |
| journal outcome | `NOT_DISPATCHED` | `Gateway.Place` result | Test fails if the execution gateway enters another transport state |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `execgw.New` returns an error | None; fixture construction aborts | `t.Fatalf` | `TestUnreachableBrokerIsNotDispatched` |
| B2 | `Gateway.Place` outcome is not `NOT_DISPATCHED` | The test has attempted one gateway placement through the synthetic transport | `t.Errorf` | `TestUnreachableBrokerIsNotDispatched` |
| B3 | transport call count is not exactly one | Detects retry or any request beyond token acquisition | `t.Errorf` | `TestUnreachableBrokerIsNotDispatched` |
| B4 | sole transport request is not `POST /oauth2/token` | Detects a mutation endpoint reaching transport | `t.Errorf` | `TestUnreachableBrokerIsNotDispatched` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `official.New` with `WithHTTPClient` | Binds the deterministic in-memory transport to the official client | Every request fails before any socket I/O; no live endpoint is reachable | AST calls at 731-733; test-owned transport |
| `execgw.New` | Builds the normal execution-gateway path under test | Construction error is fatal (B1); no fallback | AST call at 737; B1 |
| `Gateway.Place` | Exercises the production gateway once | Transport error must map to `NOT_DISPATCHED`; the gateway must not retry mutation | AST call at 745; B2-B4 |
| `t.Fatalf` / `t.Errorf` | Makes every failed invariant visible to `go test` | No retry/fallback; test terminates at B1 and records assertions at B2-B4 | AST calls at 742, 747, 752, 755 |

## State mutations and fallbacks

- This is a test-only construction. Its only mutable state is the local transport recorder (`calls`, `method`, `path`).
- The synthetic transport refuses the OAuth token request before connection/write. Therefore it neither opens a broker connection nor authorizes a mutation request.
- There is no fallback from a pre-write failure: the expected result is the conservative `NOT_DISPATCHED` terminal outcome, not a live retry or resolution procedure.

## Safety conclusion

- Safe edit boundary: deterministic test transport and assertions only; no production trading, journal schema, runtime toggle, or broker endpoint changed.
- High-risk impact: test evidence protects the execution state boundary. Its invariant is conservative: only a provable pre-write failure may remain `NOT_DISPATCHED`; any mutation-path transport call fails the test.
