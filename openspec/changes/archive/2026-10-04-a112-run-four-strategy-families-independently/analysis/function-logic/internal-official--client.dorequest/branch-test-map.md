# Branch Test Map: `Client.doRequest`

- Source: `internal/official/client.go`, SHA-256 `6d145916a6797f3a32842561346c8767585a951d8883ebd81bf8af4632106639`; branch IDs follow `ast.json` (generated 2026-08-17).
- AST counts: branches 2, returns 3, calls 20, defers 1, go statements 0. Source range `191:1`–`207:2`. Signature `(c *Client) doRequest(req *http.Request) (int, []byte, error)`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1b brief (official raw reader + bar producer). Any later body edit requires a fresh RED/BTM.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 194:2 — the round trip fails; a trace with `Err` is emitted and `(0, nil, ErrTransport)` is returned without a status or a rate record | not-applicable: transport-failure path, not injectable through this package's `httptest`-based suite | n/a (not edited) | not-applicable |
| B2 | if at 201:2 — the body read fails after the rate headers were already recorded; the status survives, the body does not | not-applicable: mid-body connection failure, not injectable here | n/a (not edited) | not-applicable |

Fall-through (the only path the suite exercises): rate record at 199, unbounded `io.ReadAll` at 200, trace with a copied body at 205, `(status, body, nil)` at 206 — covered by `TestAttemptObserverSeesBodyBeforeDecode` (trace body equals the raw response, `BodyReadComplete` at or after `RequestStart`), `TestAttemptObserverTraces401RefreshThenSuccessfulRetry` (one trace per attempt, 401 then 200), `TestAttemptObserverReturnPrecedesDecodeAndCallerReturn` (the observer runs before the caller returns) and `TestAttemptTracePublicSurfaceIsReadOnly`.

Uncovered wiring the L1b brief should note: no test asserts that a real request populates the client's rate-budget store, nor that the record precedes the body read. The pieces are unit-tested in isolation — `TestTheReportedBudgetIsKept`, `TestTheBudgetStoreIsPerPath`, `TestAResponseWithNoRateHeadersIsNotZeroRemaining`, `TestTheBudgetMapDoesNotGrowWithOrderIDs`, `TestTheBudgetKeyCollapsesObjectIdentifiers` — but never through this function.

Verification: `go test ./internal/official -count=1` green on 2026-08-17 (exit 0). No RED round applies — a112 does not edit this function.
