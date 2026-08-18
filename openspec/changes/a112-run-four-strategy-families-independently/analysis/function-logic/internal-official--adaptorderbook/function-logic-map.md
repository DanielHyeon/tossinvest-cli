# Function Logic Map: `adaptOrderbook`

- Source: `internal/official/market_reads.go`
- Source SHA-256: `b356524b48df5dc550c07d74c00f6f7c695e6364c78036e04df8df94ef2ce1ea` (current worktree; verified with `sha256sum` 2026-08-18, equal to `source_sha256` in `ast.json`)
- Signature: `adaptOrderbook(symbol string, raw apiOrderbook) domain.OrderBook` (`ast.json`: `adaptOrderbook(params=2, results=1)`)
- Source range: `229:1`-`261:2`
- AST counts: branches 2, returns 1, calls 12, defers 0, go statements 0 (`ast.json` generated 2026-08-18 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1c brief (official quote producer) and for decision 32's correction. Any later body edit requires a fresh RED/BTM.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `symbol` | caller-supplied ticker | `(*Client).Orderbook`'s parameter | copied through; never validated |
| `raw.Asks` / `raw.Bids` | `[]apiOrderbookEntry{Price, Volume string}` | `/api/v1/orderbook` `result.asks` / `result.bids` | absent/empty arrays produce empty slices, not an error |
| `raw.Currency` | `"KRW"` / `"USD"` | the response | **dropped** - `domain.OrderBook` has no currency field |
| `raw.Timestamp` | RFC3339 string, **nullable** per the schema comment at 200 | the response | **dropped** - never referenced in this body |

- The AST call list contains no reference to `raw.Timestamp` and no `time.Parse`; the only clock call is `time.Now().UTC()` at 258:15, which stamps `FetchedAt` with the **local process clock**, not the broker's instant.
- Therefore the broker's orderbook instant is not observable anywhere downstream of this function, including in `tossctl quote orderbook --output json`. A human probe run through the console cannot report whether the field was present, absent or null. This is the structural reason L1c needs a raw-bytes reader (a112 decision 10's fail-closed rule).

## Branches and early returns

Exact AST return node: `252`. There is exactly one return and no early exit; every input path reaches it.

- B1 (`range` at 232:2) - iterates `raw.Asks`. Per element: `parseDecimal(a.Volume)` at 233:10, `totalOffer += vol` at 234:3, then `append` at 235:3 of `domain.OrderBookLevel{Price: parseDecimal(a.Price), Volume: vol}` (`parseDecimal` again at 236:12). Zero elements ⇒ empty `offers`, `totalOffer == 0`.
- B2 (`range` at 243:2) - the same shape over `raw.Bids`: `parseDecimal` at 244:10 and 247:12, `totalBid += vol` at 245:3, `append` at 246:3.
- Ordering, level count and level cap are not checked. The function preserves whatever order and length the response carried; it neither asserts a fixed depth nor detects a truncated ladder.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `parseDecimal` (x4: 233:10, 236:12, 244:10, 247:12) | decimal string -> `float64` | fail-open: any malformed value becomes `0` and is indistinguishable from a real `0` (`internal-official--parsedecimal` bundle) | `ast.json` calls |
| `make`/`len`/`append` | build the two level slices | none | `ast.json` calls at 230, 241, 235, 246 |
| `time.Now().UTC()` | stamp `FetchedAt` | none; process clock, no authority | `ast.json` call at 258:15 |

## State mutations and fallbacks

- `TotalOffer` and `TotalBid` are **computed here** by the `+=` accumulators at 234:3 and 245:3. They are not response fields: `apiOrderbook` (201-206) declares only `Asks`, `Bids`, `Currency`, `Timestamp`, and the doc comment at 227 states "TotalOffer / TotalBid: summed from levels (not directly provided by API)".
- Correction this bundle supports (a112 decision 32): an observation that `total_offer_volume` equals the sum of the printed levels is a **tautology of this function**, not a measurement of the broker, and can never prove that the visible ladder is the whole book. The contrast is visible one package away - the WTS adapter reads server-sent `offerVolume`/`bidVolume` fields (`internal/client/marketdata.go:597-606`), so on the WTS path the same equality would have been a real check.
- `ProductCode`, `Name`, `Close` are left zero by construction (comment at 259). `ProductCode == ""` is therefore the official path's fingerprint in console JSON, because the WTS adapter always sets it (`internal/client/marketdata.go:612-613`).

## Safety conclusion

- Safe edit boundary: a112 does not edit this function. It is the console/hybrid orderbook projection; changing it would change what operators see without changing what evidence records, so L1c adds a separate reader instead.
- High-risk impact: no by role - the function makes no order, sizing or protection decision and holds no state. Its risk to a112 is epistemic: it is the instrument through which the 2026-08-18 human probe was read, and it silently drops currency and the broker timestamp while synthesising the two totals. A brief that treated its output as a measurement of the broker would be citing this function's arithmetic as the broker's answer (decision 32).
- Untested branch: the empty-ladder path of B1/B2. Both populated paths and the whole-body path are pinned by `TestAdaptOrderbookUnit` and `TestOrderbookIntegration`; the package suite is green (`go test ./internal/official -count=1`, 351 tests, exit 0, 2026-08-18).
