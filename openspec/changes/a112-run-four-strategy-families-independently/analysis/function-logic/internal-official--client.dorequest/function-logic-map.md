# Function Logic Map: `Client.doRequest`

- Source: `internal/official/client.go`
- Source SHA-256: `6d145916a6797f3a32842561346c8767585a951d8883ebd81bf8af4632106639` (current worktree; verified with `sha256sum` 2026-08-17, equal to `source_sha256` in `ast.json`)
- Signature: `(c *Client) doRequest(req *http.Request) (int, []byte, error)` (`ast.json`: `Client.doRequest(params=1, results=3)`)
- Source range: `191:1`–`207:2`
- AST counts: branches 2, returns 3, calls 20, defers 1, go statements 0 (`ast.json` generated 2026-08-17 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1b brief (official raw reader + bar producer). Any later body edit requires a fresh RED/BTM.

## Inputs and invariants

- The single HTTP hop shared by every official verb. `req` arrives fully built by the caller's `makeReq` closure (`Client.send` at 329 and 354); this function adds nothing to it and never rebuilds or retries it.
- Ordering is fixed and load-bearing: `started := time.Now()` (192) → `c.hc.Do` (193) → `defer resp.Body.Close()` (198) → rate headers recorded (199) → body read (200) → attempt trace emitted (205) → return. The rate-budget record happens BEFORE the body is read, so it survives every later failure, and it covers non-2xx responses too — 429 included, which is the single most informative response and the one the error mapping used to discard.
- `io.ReadAll(resp.Body)` at 200 has no byte cap and no `io.LimitReader`. The response size is bounded only by the peer and the client timeout (`defaultTimeout = 15 * time.Second`, client.go:20). This is the property a112's paged US reader must not inherit: its own reader bounds and validates before minting evidence.
- The success trace copies the body (`append([]byte(nil), body...)` at 205) so an observer cannot mutate what the caller decodes; the returned `body` is the original slice.
- Every error is wrapped as `ErrTransport`, and status code and body are returned as-is otherwise — no classification happens here (`send` B9 does that).

## Branches and early returns

Exact AST return nodes: `196, 203, 206`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 194:2 | `c.hc.Do` failed → emit `AttemptTrace{RequestStart, BodyReadComplete, Err}` and return `(0, nil, ErrTransport)`; no status, no rate record | not-applicable: transport-failure path; no test in this package makes the round-trip fail (all use `httptest` servers that answer) |
| B2 | if | 201:2 | `io.ReadAll` failed → emit `AttemptTrace{…, StatusCode, Err}` and return `(resp.StatusCode, nil, ErrTransport: reading body)`; the rate record at 199 has already happened | not-applicable: driver/connection failure mid-body; not injectable through the package's public surface in these tests |

Both branches are failure paths. The fall-through (199 record, 205 trace with a copied body, 206 return) is the only path any test in this package exercises, and it is covered by `TestAttemptObserverSeesBodyBeforeDecode`, `TestAttemptObserverTraces401RefreshThenSuccessfulRetry`, `TestAttemptObserverReturnPrecedesDecodeAndCallerReturn` and `TestAttemptTracePublicSurfaceIsReadOnly`.

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `time.Now()` | 192, 195, 199, 202, 205 | attempt start, trace completion stamps, and the rate-budget observation time |
| `c.hc.Do(req)` | 193 | the round trip; `hc` is the constructor-owned transport unless an option replaced it (`TestAuthorityOriginRejectsConfiguredTransport`) |
| `observeAttempt(req.Context(), AttemptTrace{…})` | 195, 202, 205 | one trace per attempt, delivered before the caller returns (`TestAttemptObserverReturnPrecedesDecodeAndCallerReturn`) |
| `req.Context()` | 195, 202, 205 | the observer is carried on the request context (`WithAttemptObserver`) |
| `fmt.Errorf("%w: …", ErrTransport, err)` | 196, 203 | both failure paths wrap `ErrTransport`, a `ShouldFallback` class |
| `resp.Body.Close` | 198 (deferred) | the only defer; runs after the body read and after the trace |
| `c.rates.record(readRateBudget(req.URL.Path, resp.Header, time.Now()))` | 199 | shared per-`Client` budget map; `readRateBudget` stores under `budgetKey(path)` (ratebudget.go:228), so object identifiers collapse to `{id}` |
| `io.ReadAll(resp.Body)` | 200 | unbounded body read |
| `append([]byte(nil), body...)` | 205 | defensive copy handed to the observer |

## State mutations and fallbacks

- Shared client state: `c.rates.record` (199) writes the `*Client`-scoped `rateBudgets` map under its own mutex (ratebudget.go:188–195). Every verb sharing one `*Client` therefore shares one budget view, keyed per collapsed path — that is the quota-sharing fact the L1b brief relies on, and it is why two families on one client cannot read independent allowances.
- Process-external effect: `observeAttempt` calls a caller-supplied function synchronously, so a slow observer blocks the request (asserted by `TestAttemptObserverReturnPrecedesDecodeAndCallerReturn`).
- Locals: `started`, `resp`, `body` (3 AST assignments). No fallback value is ever substituted — a failed read returns nil bytes with an error, never an empty body presented as a real one.
- Gap worth stating: no test drives `c.rates.record` end-to-end through this function. `readRateBudget` and `rateBudgets` are unit-tested directly (`TestTheReportedBudgetIsKept`, `TestTheBudgetStoreIsPerPath`, `TestAResponseWithNoRateHeadersIsNotZeroRemaining`, `TestTheBudgetMapDoesNotGrowWithOrderIDs`, `TestTheBudgetKeyCollapsesObjectIdentifiers`), so the wiring at line 199 — including its placement before the body read — is asserted by no test.

## Safety conclusion

- Safe edit boundary: the statement order is the contract. Moving the rate record after the body read would lose the budget on exactly the responses that matter (429 and truncated bodies), and adding a cap to `io.ReadAll` would change what every existing decoder sees.
- High-risk impact: yes by adjacency. It carries order, cancel and protection requests, and it is the only place a rate-limit observation exists; but it makes no trading decision and classifies nothing, so its own failure modes are all `ErrTransport` with fallback semantics.
- Untested: both AST branches (two transport-failure paths) and the rate-record wiring. The success path and the trace contract are covered. Package suite green (`go test ./internal/official -count=1`, 2026-08-17).
